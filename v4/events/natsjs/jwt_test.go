package natsjs_test

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-micro/plugins/v4/events/natsjs"
	"github.com/nats-io/jwt/v2"
	nserver "github.com/nats-io/nats-server/v2/server"
	nats "github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"
	"github.com/stretchr/testify/require"
)

// newOperatorServer starts a NATS server in operator (JWT) mode with one account
// and returns its address and the path of a user creds file for that account.
func newOperatorServer(t *testing.T) (string, string) {
	t.Helper()

	operatorKP, err := nkeys.CreateOperator()
	require.NoError(t, err)
	operatorPub, err := operatorKP.PublicKey()
	require.NoError(t, err)
	operatorClaims := jwt.NewOperatorClaims(operatorPub)
	operatorJWT, err := operatorClaims.Encode(operatorKP)
	require.NoError(t, err)
	operator, err := jwt.DecodeOperatorClaims(operatorJWT)
	require.NoError(t, err)

	accountKP, err := nkeys.CreateAccount()
	require.NoError(t, err)
	accountPub, err := accountKP.PublicKey()
	require.NoError(t, err)
	accountJWT, err := jwt.NewAccountClaims(accountPub).Encode(operatorKP)
	require.NoError(t, err)

	resolver := &nserver.MemAccResolver{}
	require.NoError(t, resolver.Store(accountPub, accountJWT))

	natsAddr := getFreeLocalhostAddress()
	natsPort, _ := strconv.Atoi(strings.Split(natsAddr, ":")[1])
	server, err := nserver.NewServer(&nserver.Options{
		Host:             strings.Split(natsAddr, ":")[0],
		Port:             natsPort,
		TrustedOperators: []*jwt.OperatorClaims{operator},
		AccountResolver:  resolver,
	})
	require.NoError(t, err)
	go server.Start()
	t.Cleanup(server.Shutdown)
	require.True(t, server.ReadyForConnections(5*time.Second))

	return natsAddr, userCreds(t, accountKP)
}

// userCreds writes a decorated creds file for a new user of the given account
// and returns the path of the file.
func userCreds(t *testing.T, accountKP nkeys.KeyPair) string {
	t.Helper()

	userKP, err := nkeys.CreateUser()
	require.NoError(t, err)
	userPub, err := userKP.PublicKey()
	require.NoError(t, err)
	userClaims := jwt.NewUserClaims(userPub)
	userClaims.IssuerAccount, err = accountKP.PublicKey()
	require.NoError(t, err)
	userJWT, err := userClaims.Encode(accountKP)
	require.NoError(t, err)
	userSeed, err := userKP.Seed()
	require.NoError(t, err)
	creds, err := jwt.FormatUserConfig(userJWT, userSeed)
	require.NoError(t, err)

	credsFile := filepath.Join(t.TempDir(), "user.creds")
	require.NoError(t, os.WriteFile(credsFile, creds, 0o600))

	return credsFile
}

func TestNatsOptionsUserCredentials(t *testing.T) {
	natsAddr, creds := newOperatorServer(t)

	t.Run("no credentials", func(t *testing.T) {
		_, err := natsjs.NewStream(natsjs.Address(natsAddr))
		require.Error(t, err)
	})

	t.Run("valid credentials", func(t *testing.T) {
		_, err := natsjs.NewStream(
			natsjs.Address(natsAddr),
			natsjs.NatsOptions(nats.UserCredentials(creds)),
		)
		require.NoError(t, err)
	})

	t.Run("valid credentials bytes", func(t *testing.T) {
		content, err := os.ReadFile(creds)
		require.NoError(t, err)
		_, err = natsjs.NewStream(
			natsjs.Address(natsAddr),
			natsjs.NatsOptions(nats.UserCredentialBytes(content)),
		)
		require.NoError(t, err)
	})

	t.Run("credentials of an unknown account", func(t *testing.T) {
		otherAccountKP, err := nkeys.CreateAccount()
		require.NoError(t, err)
		_, err = natsjs.NewStream(
			natsjs.Address(natsAddr),
			natsjs.NatsOptions(nats.UserCredentials(userCreds(t, otherAccountKP))),
		)
		require.Error(t, err)
	})

	t.Run("invalid nats option", func(t *testing.T) {
		_, err := natsjs.NewStream(
			natsjs.Address(natsAddr),
			natsjs.NatsOptions(nats.UserJWT(nil, nil)),
		)
		require.ErrorIs(t, err, nats.ErrNoUserCB)
	})
}

// Package polaris provides an etcd service registry
// https://github.com/polarismesh/polaris
package polaris

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/polarismesh/polaris-go/api"
	"github.com/polarismesh/polaris-go/pkg/model"

	"go-micro.dev/v4/logger"
	"go-micro.dev/v4/registry"
	"go-micro.dev/v4/util/cmd"
)

var (
	prefix      = "/micro/registry/"
	defaultAddr = "127.0.0.1:8091"
)

type polarisRegistry struct {
	options registry.Options
	sync.RWMutex
	register map[string]string
	//p
	namespace   string
	serverToken string
	provider    api.ProviderAPI
	consumer    api.ConsumerAPI
	// only get one instance for use polaris's loadbalance
	getOneInstance bool
}

func init() {
	cmd.DefaultRegistries["polaris"] = NewRegistry
}

func NewRegistry(opts ...registry.Option) registry.Registry {
	e := &polarisRegistry{
		options: registry.Options{
			Timeout: time.Second * 5,
		},
		register:       make(map[string]string),
		getOneInstance: false,
	}
	token := os.Getenv("POLARIS_TOKEN")
	if token != "" {
		opts = append(opts, ServerToken(token))
	}
	ns := os.Getenv("POLARIS_NAMESPACE")
	if ns != "" {
		opts = append(opts, NameSpace(token))
	}
	address := os.Getenv("MICRO_REGISTRY_ADDRESS")
	if len(address) > 0 {
		opts = append(opts, registry.Addrs(address))
	}
	configure(e, opts...)

	return e
}

func configure(e *polarisRegistry, opts ...registry.Option) error {

	for _, o := range opts {
		o(&e.options)
	}

	if e.options.Context != nil {
		ns, ok := e.options.Context.Value(nameSpaceKey{}).(string)
		if ok {
			e.namespace = ns
		}
		token, ok := e.options.Context.Value(serverTokenKey{}).(string)
		if ok {
			e.serverToken = token
		}
		flag, ok := e.options.Context.Value(getOneInstanceKey{}).(bool)
		if ok {
			e.getOneInstance = flag
		}
	}
	addr := defaultAddr
	for _, a := range e.Options().Addrs {
		if a != "" {
			addr = a
		}
	}
	consumer, err := api.NewConsumerAPIByAddress(addr)
	if err != nil {
		return err
	}
	provider, err := api.NewProviderAPIByAddress(addr)
	if err != nil {
		return err
	}
	e.consumer = consumer
	e.provider = provider
	return nil
}

func encode(s *registry.Service) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func decode(ds []byte) *registry.Service {
	var s *registry.Service
	json.Unmarshal(ds, &s)
	return s
}

func nodePath(s, id string) string {
	service := strings.Replace(s, "/", "-", -1)
	node := strings.Replace(id, "/", "-", -1)
	return path.Join(prefix, service, node)
}

func servicePath(s string) string {
	return path.Join(prefix, strings.Replace(s, "/", "-", -1))
}

func (e *polarisRegistry) addInstance(nodeId, id string) {
	e.register[nodeId] = id
}
func (e *polarisRegistry) getInstance(nodeId string) string {
	if id, ok := e.register[nodeId]; ok {
		return id
	}
	return ""
}
func (e *polarisRegistry) delInstance(nodeId string) {
	delete(e.register, nodeId)
}

func (e *polarisRegistry) Init(opts ...registry.Option) error {
	return configure(e, opts...)
}

func (e *polarisRegistry) Options() registry.Options {
	return e.options
}

func (e *polarisRegistry) registerNode(s *registry.Service, node *registry.Node, opts ...registry.RegisterOption) error {

	service := &registry.Service{
		Name:      s.Name,
		Version:   s.Version,
		Metadata:  s.Metadata,
		Endpoints: s.Endpoints,
		Nodes:     []*registry.Node{node},
	}

	addrs := strings.Split(node.Address, ":")
	if len(addrs) != 2 {
		msg := fmt.Sprintf("fail to register instance, node.Address invalid %s", node.Address)
		logger.Fatal(msg)
		return errors.New(msg)
	}
	host := addrs[0]
	port, _ := strconv.Atoi(addrs[1])

	retryCount := 3

	e.Lock()
	defer e.Unlock()

	if id := e.getInstance(node.Id); id != "" {
		req := &api.InstanceHeartbeatRequest{}
		req.Service = s.Name
		req.Namespace = e.namespace
		req.Host = host
		req.Port = port
		req.ServiceToken = e.serverToken
		req.RetryCount = &retryCount
		req.InstanceID = id
		err := e.provider.Heartbeat(req)
		if err != nil {
			logger.Errorf("fail to heartbeat instance, err is %v %v", err, req)
			return err
		}
		return nil
	} else {
		var options registry.RegisterOptions
		for _, o := range opts {
			o(&options)
		}
		version := s.Version
		req := &api.InstanceRegisterRequest{}
		req.Service = s.Name
		req.Version = &version
		req.Namespace = e.namespace
		req.Host = host
		req.Port = port
		req.ServiceToken = e.serverToken
		// 不做心跳就不要设置,否则服务器会被置不健康
		req.SetTTL(int(options.TTL.Seconds()))
		req.RetryCount = &retryCount
		mm := map[string]string{}
		mm["node_path"] = nodePath(service.Name, node.Id)
		mm["Micro-Service"] = encode(service)

		req.Metadata = mm

		resp, err := e.provider.Register(req)
		if err != nil {
			logger.Fatalf("fail to register instance, err is %v %v", err, req)
		}
		e.addInstance(node.Id, resp.InstanceID)
		// logger.Infof("register response: instanceId %s %v", resp.InstanceID, options.TTL.Seconds())
	}

	return nil
}

func (e *polarisRegistry) Deregister(s *registry.Service, opts ...registry.DeregisterOption) error {
	if len(s.Nodes) != 1 {
		return errors.New("Require must one node")
	}

	e.Lock()
	defer e.Unlock()

	for _, node := range s.Nodes {
		addrs := strings.Split(node.Address, ":")
		if len(addrs) != 2 {
			msg := fmt.Sprintf("fail to deregister instance, node.Address invalid %s", node.Address)
			logger.Error(msg)
			return errors.New(msg)
		}
		host := addrs[0]
		port, _ := strconv.Atoi(addrs[1])

		timeout := e.options.Timeout
		retryCount := 3
		// logger.Infof("start to invoke deregister operation")
		req := &api.InstanceDeRegisterRequest{}
		req.Service = s.Name
		req.Namespace = e.namespace
		req.Host = host
		req.Port = port
		req.ServiceToken = e.serverToken
		req.Timeout = &timeout
		req.RetryCount = &retryCount

		e.delInstance(node.Id)

		if err := e.provider.Deregister(req); err != nil {
			msg := fmt.Sprintf("fail to deregister instance, err is %s", err)
			logger.Error(msg)
			return errors.New(msg)
		}
		logger.Infof("deregister successfully.")
	}

	return nil
}

func (e *polarisRegistry) Register(s *registry.Service, opts ...registry.RegisterOption) error {
	if len(s.Nodes) != 1 {
		return errors.New("Require must one node")
	}

	var gerr error

	// register each node individually
	for _, node := range s.Nodes {
		err := e.registerNode(s, node, opts...)
		if err != nil {
			gerr = err
		}
	}

	return gerr
}

func (e *polarisRegistry) GetService(name string, opts ...registry.GetOption) ([]*registry.Service, error) {
	timeout := e.options.Timeout
	retryCount := 3

	inss := []model.Instance{}
	// DiscoverEchoServer
	if false == e.getOneInstance {
		req := &api.GetInstancesRequest{}
		req.Service = name
		req.Namespace = e.namespace
		req.Timeout = &timeout
		req.RetryCount = &retryCount
		insResp, err := e.consumer.GetInstances(req)
		if err != nil {
			logger.Errorf("[error] fail to GetInstances, err is %v", err)
			return nil, err
		}
		inss = insResp.GetInstances()
	} else {
		req := &api.GetOneInstanceRequest{}
		req.Service = name
		req.Namespace = e.namespace
		req.Timeout = &timeout
		req.RetryCount = &retryCount
		insResp, err := e.consumer.GetOneInstance(req)
		if err != nil {
			logger.Errorf("[error] fail to GetOneInstance, err is %v", err)
			return nil, err
		}
		inss = insResp.GetInstances()
	}

	if len(inss) == 0 {
		return []*registry.Service{}, registry.ErrNotFound
	}

	serviceMap := map[string]*registry.Service{}

	for _, n := range inss {
		m := n.GetMetadata()
		if m == nil {
			logger.Errorf("[error] fail to GetMetadata, name is %v", name)
			return nil, errors.New("fail to GetMetadata")
		}

		microService := m["Micro-Service"]
		if sn := decode([]byte(microService)); sn != nil {
			if !n.IsHealthy() {
				msg := fmt.Sprintf("{Name: %s, Version: %v, Node Count: %v}", sn.Name, sn.Version, len(sn.Nodes))
				logger.Log(logger.WarnLevel, "Service not healthy according to Polaris", msg)

				continue
			}

			if n.IsIsolated() {
				msg := fmt.Sprintf("{Name: %s, Version: %v, Node Count: %v}", sn.Name, sn.Version, len(sn.Nodes))
				logger.Log(logger.WarnLevel, "Service is isolated according to Polaris", msg)

				continue
			}

			s, ok := serviceMap[sn.Version]
			if !ok {
				s = &registry.Service{
					Name:      sn.Name,
					Version:   sn.Version,
					Metadata:  sn.Metadata,
					Endpoints: sn.Endpoints,
				}
				serviceMap[s.Version] = s
			}

			s.Nodes = append(s.Nodes, sn.Nodes...)
		}
	}

	services := make([]*registry.Service, 0, len(serviceMap))
	for _, service := range serviceMap {
		services = append(services, service)
	}

	return services, nil
}

func (e *polarisRegistry) ListServices(opts ...registry.ListOption) ([]*registry.Service, error) {
	services := make([]*registry.Service, 0)
	return services, errors.New("not support")
}

func (e *polarisRegistry) Watch(opts ...registry.WatchOption) (registry.Watcher, error) {
	return newPoWatcher(e, e.options.Timeout, opts...)
}

func (e *polarisRegistry) String() string {
	return "polaris"
}
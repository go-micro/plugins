package http

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"testing"

	"github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/registry/memory"
	"github.com/micro/go-micro/v2/server"
)

type testHandler struct{}

func (t *testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`{"hello": "world"}`))
}

func TestHTTPRouter(t *testing.T) {
	t.Log("skip broken test")
	t.Skip()
	c, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Error(err)
	}
	defer c.Close()
	addr := c.Addr().String()

	url := fmt.Sprintf("http://%s", addr)

	testCases := []struct {
		// local url e.g http://localhost:9090
		url string
		// http endpoint to call e.g /foo/bar
		httpEp string
		// rpc endpoint called e.g Foo.Bar
		rpcEp string
		// should be an error
		err bool
	}{
		{addr, "/foo/bar", "Foo.Bar", false},
		{addr, "/foo/baz", "Foo.Baz", true},
		{addr, "/helloworld", "Hello.World", false},
		{addr, "/greeter", "Greeter.Hello", false},
		{addr, "/", "Fail.Hard", true},
	}

	// handler
	http.Handle("/foo/bar", new(testHandler))
	http.Handle("/helloworld", new(testHandler))
	http.Handle("/greeter", new(testHandler))

	// new proxy
	p := NewSingleHostRouter(url)

	// register a route
	p.RegisterEndpoint("Hello.World", "/helloworld")
	p.RegisterEndpoint("Greeter.Hello", url+"/greeter")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)

	// new micro service
	service := micro.NewService(
		micro.Context(ctx),
		micro.Name("foobar"),
		micro.Registry(memory.NewRegistry()),
		micro.AfterStart(func() error {
			wg.Done()
			return nil
		}),
	)

	// set router
	service.Server().Init(
		server.WithRouter(p),
	)

	// run service
	// server
	go http.Serve(c, nil)
	go service.Run()

	// wait till service is started
	wg.Wait()

	for _, test := range testCases {
		req := service.Client().NewRequest("foobar", test.rpcEp, map[string]string{"foo": "bar"}, client.WithContentType("application/json"))
		var rsp map[string]string
		err := service.Client().Call(ctx, req, &rsp)
		if err != nil && test.err == false {
			t.Error(err)
		}
		if err == nil && test.err == true {
			t.Errorf("Expected error for %v:%v got %v and response %v", test.rpcEp, test.httpEp, err, rsp)
		} else {
			continue
		}
		if v := rsp["hello"]; v != "world" {
			t.Errorf("Expected hello world got %s from %s", v, test.rpcEp)
		}
	}
}

func TestHTTPRouterOptions(t *testing.T) {
	// test endpoint
	service := NewService(
		WithBackend("http://foo.bar"),
	)

	r := service.Server().Options().Router
	httpRouter, ok := r.(*Router)
	if !ok {
		t.Error("Expected http router to be installed")
	}
	if httpRouter.Backend != "http://foo.bar" {
		t.Errorf("Expected endpoint http://foo.bar got %v", httpRouter.Backend)
	}

	// test router
	service = NewService(
		WithRouter(&Router{Backend: "http://foo2.bar"}),
	)
	r = service.Server().Options().Router
	httpRouter, ok = r.(*Router)
	if !ok {
		t.Error("Expected http router to be installed")
	}
	if httpRouter.Backend != "http://foo2.bar" {
		t.Errorf("Expected endpoint http://foo2.bar got %v", httpRouter.Backend)
	}
}

package k8sheadlesssvc

import (
	"go-micro.dev/v4/registry"
)

// about services within the registry.
type k8sSvcWatcher struct {
}

func (k *k8sSvcWatcher) Next() (*registry.Result, error) {
	return &registry.Result{}, nil
}
func (k *k8sSvcWatcher) Stop() {}

// Service  saves the microservice u call.
type Service struct {
	Namespace string // namespace of microservice u call in k8s
	SvcName   string // Service name of microservice u call in k8s
	PodPort   int32  // the port of  container u deploy in k8s which is the value of containerPort
}
type k8sSvcRegister struct {
	k8sService []*Service
	opts       *registry.Options
}

func (k *k8sSvcRegister) Init(opts ...registry.Option) error {
	for _, o := range opts {
		o(k.opts)
	}

	return nil
}
func (k *k8sSvcRegister) Options() registry.Options {
	return registry.Options{}
}

// Register The resolution dns returns the pod id
// Since we intend to register self-discovery endpoints with k8s service,
// we do not need to write the registration discovery logic ourselves.
func (k *k8sSvcRegister) Register(*registry.Service, ...registry.RegisterOption) error {
	return nil
}

// Deregister The resolution dns returns the pod id
// Since we intend to register self-discovery endpoints with k8s service,
// we do not need to write the registration discovery logic ourselves.
func (k *k8sSvcRegister) Deregister(*registry.Service, ...registry.DeregisterOption) error {
	return nil
}

// GetService get service from endpoints of Service.
func (k *k8sSvcRegister) GetService(string, ...registry.GetOption) ([]*registry.Service, error) {
	service := make([]*registry.Service, 64)
	nodes := make([]*registry.Node, 64)

	ipMaps, err := getDNSForPodIP(k.k8sService)
	if err != nil {
		return service, err
	}

	for svcName, ips := range ipMaps {
		for _, ip := range ips {
			nodes = append(nodes, &registry.Node{Address: ip})
		}

		service = append(service, &registry.Service{Name: svcName, Version: "latest", Nodes: nodes})
	}

	return service, nil
}

// ListServices get service from endpoints of Service.
func (k *k8sSvcRegister) ListServices(...registry.ListOption) ([]*registry.Service, error) {
	service := make([]*registry.Service, 64)
	nodes := make([]*registry.Node, 64)
	ipMaps, err := getDNSForPodIP(k.k8sService)

	if err != nil {
		return []*registry.Service{}, err
	}

	for svcName, ips := range ipMaps {
		for _, ip := range ips {
			nodes = append(nodes, &registry.Node{Address: ip})
		}

		service = append(service, &registry.Service{Name: svcName, Version: "latest", Nodes: nodes})
	}

	return service, nil
}

// Watch Since we intend to register self-discovery endpoints with k8s service,
// we do not need to write the registration discovery logic ourselves.
func (k *k8sSvcRegister) Watch(_ ...registry.WatchOption) (registry.Watcher, error) {
	return &k8sSvcWatcher{}, nil
}

func (k *k8sSvcRegister) String() string {
	return "k8s-headless-svc"
}

// NewRegistry creates a kubernetes registry.
func NewRegistry(k8sService []*Service, opts ...registry.Option) registry.Registry {
	k := k8sSvcRegister{
		k8sService: k8sService,
		opts:       &registry.Options{},
	}

	return &k
}

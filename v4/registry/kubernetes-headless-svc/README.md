Kubernetes Registry Plugin for micro
---

## about kubernetes-headless-svc registry

The current project is a go-micro register plug-in.
When we deploy the go-micro grpc server with k8s, if we use the built-in Service kind of k8s
to deploy the pod of grpc server, it is not possible to perform grpc http2.0, so I use the 
headless Service mode to deploy, and we can use the headless Service as a dns host, so we 
can  use go package  `net` to do dns resolution, and then obtain the podIp record of headless 
Service endpoints return, realize the function of grpc service discovery.
## how to use  kubernetes-headless-svc registry
```go
package main

import (
	"fmt"
	"github.com/go-micro-v4-demo/frontend/handler"
	helloworldPb "github.com/go-micro-v4-demo/helloworld/proto"
	userPb "github.com/go-micro-v4-demo/user/proto"
	mgrpc "github.com/go-micro/plugins/v4/client/grpc"
	mhttp "github.com/go-micro/plugins/v4/server/http"
	"github.com/gorilla/mux"
	k8sHeadlessSvc "github.com/gsmini/k8s-headless-svc"
	"go-micro.dev/v4/logger"
	"net/http"
)

var (
	service = "frontend"
	version = "latest"
)

const K8sSvcName = "user-svc"

const UserSvcName = "user-svc"        //the name of Service.meta.name in k8s
const HelloWordSvcName = "helloworld" //the name of Service.meta.name in k8s
func main() {
	UserSvc := &k8sHeadlessSvc.Service{Namespace: "default", SvcName: UserSvcName, PodPort: 8080}
	//HelloWordSvc := &k8sHeadlessSvc.Service{Namespace: "default", SvcName: HelloWordSvcName, PodPort: 9090}
	reg := k8sHeadlessSvc.NewRegistry([]*k8sHeadlessSvc.Service{UserSvc})
	// when frontend has many grpc server, u can use like this
	//reg := k8sHeadlessSvc.NewRegistry([]*k8sHeadlessSvc.Service{UserSvc},[]*k8sHeadlessSvc.Service{HelloWordSvcName})
	srv := micro.NewService()
	srv.Init(
		micro.Name(service),
		micro.Version(version),
		micro.Address("0.0.0.0:8080"), 
		micro.Registry(reg),//registry our k8sHeadlessSvc registry
	)
	//Omit unimportant code ...
}
```

## Core  code of  kubernetes-headless-svc
```go
package main
import (
	"fmt"
	"net"
)

func main() {
	//we just use net.LookupIP to get ip address of www.twitter.com
	ipRecords, err := net.LookupIP("www.twitter.com")
	if err != nil {
		panic(err)
	}
	for _, value := range ipRecords {
		fmt.Println(value.String())
	}
}
```

```shell
104.244.42.193
```
> it is the same of this command: 'nslookup www.twitter.com'
```shell
Server:         8.8.8.8
Address:        8.8.8.8#53

Non-authoritative answer:
www.twitter.com canonical name = twitter.com.
Name:   twitter.com
Address: 104.244.42.193

```


When deploying the grpc server, configure the sessionAffinity(session affinity) for the Service
to ensure that the grpc server can return messages properly after receiving them
```yaml
apiVersion: v1
kind: Service
metadata:
  name: user-svc
  namespace: default
spec:
  clusterIP: None
  ports:
    - port: 8080
  selector:
    app:  user

  sessionAffinity: ClientIP
  sessionAffinityConfig:
    clientIP:
      timeoutSeconds: 3600
```
## examples
See the fronted and user projects under examples directory
### deploy microservice
```shell
kubectl apply -f examples/fronted/k8s.yaml
kubectl apply -f examples/user/k8s.yaml
```
> deploy our microservice  both user and fronted service

### list svc in k8s
```shell
root@hecs-410147:# kubectl  get svc
NAME             TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)    AGE
frontend-svc     ClusterIP   10.108.199.130   <none>        80/TCP     40h
```
### requests frontend-svc clusterIp
```shell

curl http://10.108.199.130/index
```
### check log of user microservice
```shell
root@hecs-410147:~# kubectl logs user-5cdd5697f-vr5db
2023-04-09 22:22:21  file=build/main.go:33 level=info Starting [service] user
2023-04-09 22:22:21  file=v4@v4.9.0/service.go:96 level=info Transport [http] Listening on [::]:8080
2023-04-09 22:22:21  file=v4@v4.9.0/service.go:96 level=info Broker [http] Connected to 127.0.0.1:33039
2023-04-09 22:22:21  file=server/rpc_server.go:832 level=info Registry [memory] Registering node: user-defaaa6b-7314-4757-bb47-9a1ea6043d0d
2023-04-11 20:46:35  file=handler/user.go:16 level=info Received User.Call request: name:"gsmini@sina.cn"
2023-04-11 21:23:35  file=handler/user.go:16 level=info Received User.Call request: name:"gsmini@sina.cn"
2023-04-11 21:25:00  file=handler/user.go:16 level=info Received User.Call request: name:"gsmini@sina.cn"
2023-04-11 21:35:39  file=handler/user.go:16 level=info Received User.Call request: name:"gsmini@sina.cn"
2023-04-11 21:35:49  file=handler/user.go:16 level=info Received User.Call request: name:"gsmini@sina.cn"
```
> so we can find the request log from fronted application
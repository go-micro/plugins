// Code generated by protoc-gen-micro. DO NOT EDIT.
// source: proto/user.proto

package user

import (
	fmt "fmt"
	proto "google.golang.org/protobuf/proto"
	math "math"
)

import (
	context "context"
	api "go-micro.dev/v4/api"
	client "go-micro.dev/v4/client"
	server "go-micro.dev/v4/server"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// Reference imports to suppress errors if they are not otherwise used.
var _ api.Endpoint
var _ context.Context
var _ client.Option
var _ server.Option

// Api Endpoints for User service

func NewUserEndpoints() []*api.Endpoint {
	return []*api.Endpoint{}
}

// Client API for User service

type UserService interface {
	Call(ctx context.Context, in *CallRequest, opts ...client.CallOption) (*CallResponse, error)
	ClientStream(ctx context.Context, opts ...client.CallOption) (User_ClientStreamService, error)
	ServerStream(ctx context.Context, in *ServerStreamRequest, opts ...client.CallOption) (User_ServerStreamService, error)
	BidiStream(ctx context.Context, opts ...client.CallOption) (User_BidiStreamService, error)
}

type userService struct {
	c    client.Client
	name string
}

func NewUserService(name string, c client.Client) UserService {
	return &userService{
		c:    c,
		name: name,
	}
}

func (c *userService) Call(ctx context.Context, in *CallRequest, opts ...client.CallOption) (*CallResponse, error) {
	req := c.c.NewRequest(c.name, "User.Call", in)
	out := new(CallResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userService) ClientStream(ctx context.Context, opts ...client.CallOption) (User_ClientStreamService, error) {
	req := c.c.NewRequest(c.name, "User.ClientStream", &ClientStreamRequest{})
	stream, err := c.c.Stream(ctx, req, opts...)
	if err != nil {
		return nil, err
	}
	return &userServiceClientStream{stream}, nil
}

type User_ClientStreamService interface {
	Context() context.Context
	SendMsg(interface{}) error
	RecvMsg(interface{}) error
	CloseSend() error
	Close() error
	Send(*ClientStreamRequest) error
}

type userServiceClientStream struct {
	stream client.Stream
}

func (x *userServiceClientStream) CloseSend() error {
	return x.stream.CloseSend()
}

func (x *userServiceClientStream) Close() error {
	return x.stream.Close()
}

func (x *userServiceClientStream) Context() context.Context {
	return x.stream.Context()
}

func (x *userServiceClientStream) SendMsg(m interface{}) error {
	return x.stream.Send(m)
}

func (x *userServiceClientStream) RecvMsg(m interface{}) error {
	return x.stream.Recv(m)
}

func (x *userServiceClientStream) Send(m *ClientStreamRequest) error {
	return x.stream.Send(m)
}

func (c *userService) ServerStream(ctx context.Context, in *ServerStreamRequest, opts ...client.CallOption) (User_ServerStreamService, error) {
	req := c.c.NewRequest(c.name, "User.ServerStream", &ServerStreamRequest{})
	stream, err := c.c.Stream(ctx, req, opts...)
	if err != nil {
		return nil, err
	}
	if err := stream.Send(in); err != nil {
		return nil, err
	}
	return &userServiceServerStream{stream}, nil
}

type User_ServerStreamService interface {
	Context() context.Context
	SendMsg(interface{}) error
	RecvMsg(interface{}) error
	CloseSend() error
	Close() error
	Recv() (*ServerStreamResponse, error)
}

type userServiceServerStream struct {
	stream client.Stream
}

func (x *userServiceServerStream) CloseSend() error {
	return x.stream.CloseSend()
}

func (x *userServiceServerStream) Close() error {
	return x.stream.Close()
}

func (x *userServiceServerStream) Context() context.Context {
	return x.stream.Context()
}

func (x *userServiceServerStream) SendMsg(m interface{}) error {
	return x.stream.Send(m)
}

func (x *userServiceServerStream) RecvMsg(m interface{}) error {
	return x.stream.Recv(m)
}

func (x *userServiceServerStream) Recv() (*ServerStreamResponse, error) {
	m := new(ServerStreamResponse)
	err := x.stream.Recv(m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (c *userService) BidiStream(ctx context.Context, opts ...client.CallOption) (User_BidiStreamService, error) {
	req := c.c.NewRequest(c.name, "User.BidiStream", &BidiStreamRequest{})
	stream, err := c.c.Stream(ctx, req, opts...)
	if err != nil {
		return nil, err
	}
	return &userServiceBidiStream{stream}, nil
}

type User_BidiStreamService interface {
	Context() context.Context
	SendMsg(interface{}) error
	RecvMsg(interface{}) error
	CloseSend() error
	Close() error
	Send(*BidiStreamRequest) error
	Recv() (*BidiStreamResponse, error)
}

type userServiceBidiStream struct {
	stream client.Stream
}

func (x *userServiceBidiStream) CloseSend() error {
	return x.stream.CloseSend()
}

func (x *userServiceBidiStream) Close() error {
	return x.stream.Close()
}

func (x *userServiceBidiStream) Context() context.Context {
	return x.stream.Context()
}

func (x *userServiceBidiStream) SendMsg(m interface{}) error {
	return x.stream.Send(m)
}

func (x *userServiceBidiStream) RecvMsg(m interface{}) error {
	return x.stream.Recv(m)
}

func (x *userServiceBidiStream) Send(m *BidiStreamRequest) error {
	return x.stream.Send(m)
}

func (x *userServiceBidiStream) Recv() (*BidiStreamResponse, error) {
	m := new(BidiStreamResponse)
	err := x.stream.Recv(m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// Server API for User service

type UserHandler interface {
	Call(context.Context, *CallRequest, *CallResponse) error
	ClientStream(context.Context, User_ClientStreamStream) error
	ServerStream(context.Context, *ServerStreamRequest, User_ServerStreamStream) error
	BidiStream(context.Context, User_BidiStreamStream) error
}

func RegisterUserHandler(s server.Server, hdlr UserHandler, opts ...server.HandlerOption) error {
	type user interface {
		Call(ctx context.Context, in *CallRequest, out *CallResponse) error
		ClientStream(ctx context.Context, stream server.Stream) error
		ServerStream(ctx context.Context, stream server.Stream) error
		BidiStream(ctx context.Context, stream server.Stream) error
	}
	type User struct {
		user
	}
	h := &userHandler{hdlr}
	return s.Handle(s.NewHandler(&User{h}, opts...))
}

type userHandler struct {
	UserHandler
}

func (h *userHandler) Call(ctx context.Context, in *CallRequest, out *CallResponse) error {
	return h.UserHandler.Call(ctx, in, out)
}

func (h *userHandler) ClientStream(ctx context.Context, stream server.Stream) error {
	return h.UserHandler.ClientStream(ctx, &userClientStreamStream{stream})
}

type User_ClientStreamStream interface {
	Context() context.Context
	SendMsg(interface{}) error
	RecvMsg(interface{}) error
	Close() error
	Recv() (*ClientStreamRequest, error)
}

type userClientStreamStream struct {
	stream server.Stream
}

func (x *userClientStreamStream) Close() error {
	return x.stream.Close()
}

func (x *userClientStreamStream) Context() context.Context {
	return x.stream.Context()
}

func (x *userClientStreamStream) SendMsg(m interface{}) error {
	return x.stream.Send(m)
}

func (x *userClientStreamStream) RecvMsg(m interface{}) error {
	return x.stream.Recv(m)
}

func (x *userClientStreamStream) Recv() (*ClientStreamRequest, error) {
	m := new(ClientStreamRequest)
	if err := x.stream.Recv(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (h *userHandler) ServerStream(ctx context.Context, stream server.Stream) error {
	m := new(ServerStreamRequest)
	if err := stream.Recv(m); err != nil {
		return err
	}
	return h.UserHandler.ServerStream(ctx, m, &userServerStreamStream{stream})
}

type User_ServerStreamStream interface {
	Context() context.Context
	SendMsg(interface{}) error
	RecvMsg(interface{}) error
	Close() error
	Send(*ServerStreamResponse) error
}

type userServerStreamStream struct {
	stream server.Stream
}

func (x *userServerStreamStream) Close() error {
	return x.stream.Close()
}

func (x *userServerStreamStream) Context() context.Context {
	return x.stream.Context()
}

func (x *userServerStreamStream) SendMsg(m interface{}) error {
	return x.stream.Send(m)
}

func (x *userServerStreamStream) RecvMsg(m interface{}) error {
	return x.stream.Recv(m)
}

func (x *userServerStreamStream) Send(m *ServerStreamResponse) error {
	return x.stream.Send(m)
}

func (h *userHandler) BidiStream(ctx context.Context, stream server.Stream) error {
	return h.UserHandler.BidiStream(ctx, &userBidiStreamStream{stream})
}

type User_BidiStreamStream interface {
	Context() context.Context
	SendMsg(interface{}) error
	RecvMsg(interface{}) error
	Close() error
	Send(*BidiStreamResponse) error
	Recv() (*BidiStreamRequest, error)
}

type userBidiStreamStream struct {
	stream server.Stream
}

func (x *userBidiStreamStream) Close() error {
	return x.stream.Close()
}

func (x *userBidiStreamStream) Context() context.Context {
	return x.stream.Context()
}

func (x *userBidiStreamStream) SendMsg(m interface{}) error {
	return x.stream.Send(m)
}

func (x *userBidiStreamStream) RecvMsg(m interface{}) error {
	return x.stream.Recv(m)
}

func (x *userBidiStreamStream) Send(m *BidiStreamResponse) error {
	return x.stream.Send(m)
}

func (x *userBidiStreamStream) Recv() (*BidiStreamRequest, error) {
	m := new(BidiStreamRequest)
	if err := x.stream.Recv(m); err != nil {
		return nil, err
	}
	return m, nil
}
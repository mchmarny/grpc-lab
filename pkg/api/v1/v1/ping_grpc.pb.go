// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package v1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion7

// ServiceClient is the client API for Service service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ServiceClient interface {
	// Ping method on the service.
	Ping(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PingResponse, error)
	// Stream is like Ping but with stream
	Stream(ctx context.Context, opts ...grpc.CallOption) (Service_StreamClient, error)
}

type serviceClient struct {
	cc grpc.ClientConnInterface
}

func NewServiceClient(cc grpc.ClientConnInterface) ServiceClient {
	return &serviceClient{cc}
}

func (c *serviceClient) Ping(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PingResponse, error) {
	out := new(PingResponse)
	err := c.cc.Invoke(ctx, "/io.thingz.grpc.v1.Service/Ping", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) Stream(ctx context.Context, opts ...grpc.CallOption) (Service_StreamClient, error) {
	stream, err := c.cc.NewStream(ctx, &_Service_serviceDesc.Streams[0], "/io.thingz.grpc.v1.Service/Stream", opts...)
	if err != nil {
		return nil, err
	}
	x := &serviceStreamClient{stream}
	return x, nil
}

type Service_StreamClient interface {
	Send(*PingRequest) error
	Recv() (*PingResponse, error)
	grpc.ClientStream
}

type serviceStreamClient struct {
	grpc.ClientStream
}

func (x *serviceStreamClient) Send(m *PingRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *serviceStreamClient) Recv() (*PingResponse, error) {
	m := new(PingResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// ServiceServer is the server API for Service service.
// All implementations must embed UnimplementedServiceServer
// for forward compatibility
type ServiceServer interface {
	// Ping method on the service.
	Ping(context.Context, *PingRequest) (*PingResponse, error)
	// Stream is like Ping but with stream
	Stream(Service_StreamServer) error
	mustEmbedUnimplementedServiceServer()
}

// UnimplementedServiceServer must be embedded to have forward compatible implementations.
type UnimplementedServiceServer struct {
}

func (UnimplementedServiceServer) Ping(context.Context, *PingRequest) (*PingResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}
func (UnimplementedServiceServer) Stream(Service_StreamServer) error {
	return status.Errorf(codes.Unimplemented, "method Stream not implemented")
}
func (UnimplementedServiceServer) mustEmbedUnimplementedServiceServer() {}

// UnsafeServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ServiceServer will
// result in compilation errors.
type UnsafeServiceServer interface {
	mustEmbedUnimplementedServiceServer()
}

func RegisterServiceServer(s grpc.ServiceRegistrar, srv ServiceServer) {
	s.RegisterService(&_Service_serviceDesc, srv)
}

func _Service_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PingRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ServiceServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/io.thingz.grpc.v1.Service/Ping",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ServiceServer).Ping(ctx, req.(*PingRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Service_Stream_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(ServiceServer).Stream(&serviceStreamServer{stream})
}

type Service_StreamServer interface {
	Send(*PingResponse) error
	Recv() (*PingRequest, error)
	grpc.ServerStream
}

type serviceStreamServer struct {
	grpc.ServerStream
}

func (x *serviceStreamServer) Send(m *PingResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *serviceStreamServer) Recv() (*PingRequest, error) {
	m := new(PingRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

var _Service_serviceDesc = grpc.ServiceDesc{
	ServiceName: "io.thingz.grpc.v1.Service",
	HandlerType: (*ServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Ping",
			Handler:    _Service_Ping_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Stream",
			Handler:       _Service_Stream_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "v1/ping.proto",
}

// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v4.25.1
// source: position.proto

package pb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// PositionClient is the client API for Position service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type PositionClient interface {
	Positioner(ctx context.Context, opts ...grpc.CallOption) (Position_PositionerClient, error)
}

type positionClient struct {
	cc grpc.ClientConnInterface
}

func NewPositionClient(cc grpc.ClientConnInterface) PositionClient {
	return &positionClient{cc}
}

func (c *positionClient) Positioner(ctx context.Context, opts ...grpc.CallOption) (Position_PositionerClient, error) {
	stream, err := c.cc.NewStream(ctx, &Position_ServiceDesc.Streams[0], "/pb.Position/Positioner", opts...)
	if err != nil {
		return nil, err
	}
	x := &positionPositionerClient{stream}
	return x, nil
}

type Position_PositionerClient interface {
	Send(*RequestPositioner) error
	CloseAndRecv() (*emptypb.Empty, error)
	grpc.ClientStream
}

type positionPositionerClient struct {
	grpc.ClientStream
}

func (x *positionPositionerClient) Send(m *RequestPositioner) error {
	return x.ClientStream.SendMsg(m)
}

func (x *positionPositionerClient) CloseAndRecv() (*emptypb.Empty, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(emptypb.Empty)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// PositionServer is the server API for Position service.
// All implementations must embed UnimplementedPositionServer
// for forward compatibility
type PositionServer interface {
	Positioner(Position_PositionerServer) error
	mustEmbedUnimplementedPositionServer()
}

// UnimplementedPositionServer must be embedded to have forward compatible implementations.
type UnimplementedPositionServer struct {
}

func (UnimplementedPositionServer) Positioner(Position_PositionerServer) error {
	return status.Errorf(codes.Unimplemented, "method Positioner not implemented")
}
func (UnimplementedPositionServer) mustEmbedUnimplementedPositionServer() {}

// UnsafePositionServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to PositionServer will
// result in compilation errors.
type UnsafePositionServer interface {
	mustEmbedUnimplementedPositionServer()
}

func RegisterPositionServer(s grpc.ServiceRegistrar, srv PositionServer) {
	s.RegisterService(&Position_ServiceDesc, srv)
}

func _Position_Positioner_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(PositionServer).Positioner(&positionPositionerServer{stream})
}

type Position_PositionerServer interface {
	SendAndClose(*emptypb.Empty) error
	Recv() (*RequestPositioner, error)
	grpc.ServerStream
}

type positionPositionerServer struct {
	grpc.ServerStream
}

func (x *positionPositionerServer) SendAndClose(m *emptypb.Empty) error {
	return x.ServerStream.SendMsg(m)
}

func (x *positionPositionerServer) Recv() (*RequestPositioner, error) {
	m := new(RequestPositioner)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Position_ServiceDesc is the grpc.ServiceDesc for Position service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Position_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "pb.Position",
	HandlerType: (*PositionServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Positioner",
			Handler:       _Position_Positioner_Handler,
			ClientStreams: true,
		},
	},
	Metadata: "position.proto",
}

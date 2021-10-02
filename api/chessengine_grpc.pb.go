// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package api

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// RustChessClient is the client API for RustChess service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type RustChessClient interface {
	Game(ctx context.Context, opts ...grpc.CallOption) (RustChess_GameClient, error)
}

type rustChessClient struct {
	cc grpc.ClientConnInterface
}

func NewRustChessClient(cc grpc.ClientConnInterface) RustChessClient {
	return &rustChessClient{cc}
}

func (c *rustChessClient) Game(ctx context.Context, opts ...grpc.CallOption) (RustChess_GameClient, error) {
	stream, err := c.cc.NewStream(ctx, &RustChess_ServiceDesc.Streams[0], "/rustchess.RustChess/Game", opts...)
	if err != nil {
		return nil, err
	}
	x := &rustChessGameClient{stream}
	return x, nil
}

type RustChess_GameClient interface {
	Send(*GameMessage) error
	Recv() (*GameMessage, error)
	grpc.ClientStream
}

type rustChessGameClient struct {
	grpc.ClientStream
}

func (x *rustChessGameClient) Send(m *GameMessage) error {
	return x.ClientStream.SendMsg(m)
}

func (x *rustChessGameClient) Recv() (*GameMessage, error) {
	m := new(GameMessage)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// RustChessServer is the server API for RustChess service.
// All implementations must embed UnimplementedRustChessServer
// for forward compatibility
type RustChessServer interface {
	Game(RustChess_GameServer) error
	mustEmbedUnimplementedRustChessServer()
}

// UnimplementedRustChessServer must be embedded to have forward compatible implementations.
type UnimplementedRustChessServer struct {
}

func (UnimplementedRustChessServer) Game(RustChess_GameServer) error {
	return status.Errorf(codes.Unimplemented, "method Game not implemented")
}
func (UnimplementedRustChessServer) mustEmbedUnimplementedRustChessServer() {}

// UnsafeRustChessServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to RustChessServer will
// result in compilation errors.
type UnsafeRustChessServer interface {
	mustEmbedUnimplementedRustChessServer()
}

func RegisterRustChessServer(s grpc.ServiceRegistrar, srv RustChessServer) {
	s.RegisterService(&RustChess_ServiceDesc, srv)
}

func _RustChess_Game_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(RustChessServer).Game(&rustChessGameServer{stream})
}

type RustChess_GameServer interface {
	Send(*GameMessage) error
	Recv() (*GameMessage, error)
	grpc.ServerStream
}

type rustChessGameServer struct {
	grpc.ServerStream
}

func (x *rustChessGameServer) Send(m *GameMessage) error {
	return x.ServerStream.SendMsg(m)
}

func (x *rustChessGameServer) Recv() (*GameMessage, error) {
	m := new(GameMessage)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// RustChess_ServiceDesc is the grpc.ServiceDesc for RustChess service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var RustChess_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "rustchess.RustChess",
	HandlerType: (*RustChessServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Game",
			Handler:       _RustChess_Game_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "api/chessengine.proto",
}
// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion7

// RoomClient is the client API for GetRoom service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type RoomClient interface {
	Create(ctx context.Context, in *CreateRoomRequest, opts ...grpc.CallOption) (*CreateRoomResponse, error)
	Subscribe(ctx context.Context, in *RoomSubscribeRequest, opts ...grpc.CallOption) (*RoomSubscribeResponse, error)
	GetByCriteria(ctx context.Context, in *GetRoomsByCriteriaRequest, opts ...grpc.CallOption) (*GetRoomsByCriteriaResponse, error)
	CloseRoom(ctx context.Context, in *CloseRoomRequest, opts ...grpc.CallOption) (*CloseRoomResponse, error)
}

type roomClient struct {
	cc grpc.ClientConnInterface
}

func NewRoomClient(cc grpc.ClientConnInterface) RoomClient {
	return &roomClient{cc}
}

func (c *roomClient) Create(ctx context.Context, in *CreateRoomRequest, opts ...grpc.CallOption) (*CreateRoomResponse, error) {
	out := new(CreateRoomResponse)
	err := c.cc.Invoke(ctx, "/proto.GetRoom/Create", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *roomClient) Subscribe(ctx context.Context, in *RoomSubscribeRequest, opts ...grpc.CallOption) (*RoomSubscribeResponse, error) {
	out := new(RoomSubscribeResponse)
	err := c.cc.Invoke(ctx, "/proto.GetRoom/Subscribe", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *roomClient) GetByCriteria(ctx context.Context, in *GetRoomsByCriteriaRequest, opts ...grpc.CallOption) (*GetRoomsByCriteriaResponse, error) {
	out := new(GetRoomsByCriteriaResponse)
	err := c.cc.Invoke(ctx, "/proto.GetRoom/GetByCriteria", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *roomClient) CloseRoom(ctx context.Context, in *CloseRoomRequest, opts ...grpc.CallOption) (*CloseRoomResponse, error) {
	out := new(CloseRoomResponse)
	err := c.cc.Invoke(ctx, "/proto.GetRoom/CloseRoom", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// RoomServer is the server API for GetRoom service.
// All implementations must embed UnimplementedRoomServer
// for forward compatibility
type RoomServer interface {
	Create(context.Context, *CreateRoomRequest) (*CreateRoomResponse, error)
	Subscribe(context.Context, *RoomSubscribeRequest) (*RoomSubscribeResponse, error)
	GetByCriteria(context.Context, *GetRoomsByCriteriaRequest) (*GetRoomsByCriteriaResponse, error)
	CloseRoom(context.Context, *CloseRoomRequest) (*CloseRoomResponse, error)
	mustEmbedUnimplementedRoomServer()
}

// UnimplementedRoomServer must be embedded to have forward compatible implementations.
type UnimplementedRoomServer struct {
}

func (UnimplementedRoomServer) Create(context.Context, *CreateRoomRequest) (*CreateRoomResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Create not implemented")
}
func (UnimplementedRoomServer) Subscribe(context.Context, *RoomSubscribeRequest) (*RoomSubscribeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Subscribe not implemented")
}
func (UnimplementedRoomServer) GetByCriteria(context.Context, *GetRoomsByCriteriaRequest) (*GetRoomsByCriteriaResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetByCriteria not implemented")
}
func (UnimplementedRoomServer) CloseRoom(context.Context, *CloseRoomRequest) (*CloseRoomResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CloseRoom not implemented")
}
func (UnimplementedRoomServer) mustEmbedUnimplementedRoomServer() {}

// UnsafeRoomServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to RoomServer will
// result in compilation errors.
type UnsafeRoomServer interface {
	mustEmbedUnimplementedRoomServer()
}

func RegisterRoomServer(s grpc.ServiceRegistrar, srv RoomServer) {
	s.RegisterService(&_Room_serviceDesc, srv)
}

func _Room_Create_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateRoomRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RoomServer).Create(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.GetRoom/Create",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RoomServer).Create(ctx, req.(*CreateRoomRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Room_Subscribe_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RoomSubscribeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RoomServer).Subscribe(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.GetRoom/Subscribe",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RoomServer).Subscribe(ctx, req.(*RoomSubscribeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Room_GetByCriteria_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRoomsByCriteriaRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RoomServer).GetByCriteria(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.GetRoom/GetByCriteria",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RoomServer).GetByCriteria(ctx, req.(*GetRoomsByCriteriaRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Room_CloseRoom_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CloseRoomRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RoomServer).CloseRoom(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.GetRoom/CloseRoom",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RoomServer).CloseRoom(ctx, req.(*CloseRoomRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Room_serviceDesc = grpc.ServiceDesc{
	ServiceName: "proto.GetRoom",
	HandlerType: (*RoomServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Create",
			Handler:    _Room_Create_Handler,
		},
		{
			MethodName: "Subscribe",
			Handler:    _Room_Subscribe_Handler,
		},
		{
			MethodName: "GetByCriteria",
			Handler:    _Room_GetByCriteria_Handler,
		},
		{
			MethodName: "CloseRoom",
			Handler:    _Room_CloseRoom_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "roomService.proto",
}
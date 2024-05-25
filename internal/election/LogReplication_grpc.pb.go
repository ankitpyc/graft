// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v5.26.1
// source: internal/election/LogReplication.proto

package election

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

const (
	RaftLogReplication_AppendEntries_FullMethodName = "/election.RaftLogReplication/AppendEntries"
)

// RaftLogReplicationClient is the client API for RaftLogReplication service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type RaftLogReplicationClient interface {
	AppendEntries(ctx context.Context, in *AppendEntriesRequest, opts ...grpc.CallOption) (*AppendEntriesResponse, error)
}

type raftLogReplicationClient struct {
	cc grpc.ClientConnInterface
}

func NewRaftLogReplicationClient(cc grpc.ClientConnInterface) RaftLogReplicationClient {
	return &raftLogReplicationClient{cc}
}

func (c *raftLogReplicationClient) AppendEntries(ctx context.Context, in *AppendEntriesRequest, opts ...grpc.CallOption) (*AppendEntriesResponse, error) {
	out := new(AppendEntriesResponse)
	err := c.cc.Invoke(ctx, RaftLogReplication_AppendEntries_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// RaftLogReplicationServer is the server API for RaftLogReplication service.
// All implementations must embed UnimplementedRaftLogReplicationServer
// for forward compatibility
type RaftLogReplicationServer interface {
	AppendEntries(context.Context, *AppendEntriesRequest) (*AppendEntriesResponse, error)
	mustEmbedUnimplementedRaftLogReplicationServer()
}

// UnimplementedRaftLogReplicationServer must be embedded to have forward compatible implementations.
type UnimplementedRaftLogReplicationServer struct {
}

func (UnimplementedRaftLogReplicationServer) AppendEntries(context.Context, *AppendEntriesRequest) (*AppendEntriesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AppendEntries not implemented")
}
func (UnimplementedRaftLogReplicationServer) mustEmbedUnimplementedRaftLogReplicationServer() {}

// UnsafeRaftLogReplicationServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to RaftLogReplicationServer will
// result in compilation errors.
type UnsafeRaftLogReplicationServer interface {
	mustEmbedUnimplementedRaftLogReplicationServer()
}

func RegisterRaftLogReplicationServer(s grpc.ServiceRegistrar, srv RaftLogReplicationServer) {
	s.RegisterService(&RaftLogReplication_ServiceDesc, srv)
}

func _RaftLogReplication_AppendEntries_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AppendEntriesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RaftLogReplicationServer).AppendEntries(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RaftLogReplication_AppendEntries_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RaftLogReplicationServer).AppendEntries(ctx, req.(*AppendEntriesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// RaftLogReplication_ServiceDesc is the grpc.ServiceDesc for RaftLogReplication service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var RaftLogReplication_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "election.RaftLogReplication",
	HandlerType: (*RaftLogReplicationServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "AppendEntries",
			Handler:    _RaftLogReplication_AppendEntries_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "internal/election/LogReplication.proto",
}
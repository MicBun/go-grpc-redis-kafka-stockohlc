// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v3.6.1
// source: stock.proto

package pb

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
	DataStock_GetOneSummary_FullMethodName = "/DataStock/GetOneSummary"
)

// DataStockClient is the client API for DataStock service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type DataStockClient interface {
	GetOneSummary(ctx context.Context, in *GetOneSummaryRequest, opts ...grpc.CallOption) (*Stock, error)
}

type dataStockClient struct {
	cc grpc.ClientConnInterface
}

func NewDataStockClient(cc grpc.ClientConnInterface) DataStockClient {
	return &dataStockClient{cc}
}

func (c *dataStockClient) GetOneSummary(ctx context.Context, in *GetOneSummaryRequest, opts ...grpc.CallOption) (*Stock, error) {
	out := new(Stock)
	err := c.cc.Invoke(ctx, DataStock_GetOneSummary_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DataStockServer is the server API for DataStock service.
// All implementations must embed UnimplementedDataStockServer
// for forward compatibility
type DataStockServer interface {
	GetOneSummary(context.Context, *GetOneSummaryRequest) (*Stock, error)
	mustEmbedUnimplementedDataStockServer()
}

// UnimplementedDataStockServer must be embedded to have forward compatible implementations.
type UnimplementedDataStockServer struct {
}

func (UnimplementedDataStockServer) GetOneSummary(context.Context, *GetOneSummaryRequest) (*Stock, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetOneSummary not implemented")
}
func (UnimplementedDataStockServer) mustEmbedUnimplementedDataStockServer() {}

// UnsafeDataStockServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to DataStockServer will
// result in compilation errors.
type UnsafeDataStockServer interface {
	mustEmbedUnimplementedDataStockServer()
}

func RegisterDataStockServer(s grpc.ServiceRegistrar, srv DataStockServer) {
	s.RegisterService(&DataStock_ServiceDesc, srv)
}

func _DataStock_GetOneSummary_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetOneSummaryRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DataStockServer).GetOneSummary(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: DataStock_GetOneSummary_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DataStockServer).GetOneSummary(ctx, req.(*GetOneSummaryRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// DataStock_ServiceDesc is the grpc.ServiceDesc for DataStock service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var DataStock_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "DataStock",
	HandlerType: (*DataStockServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetOneSummary",
			Handler:    _DataStock_GetOneSummary_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "stock.proto",
}

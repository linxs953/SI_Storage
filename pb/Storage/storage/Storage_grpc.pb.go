// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v3.21.12
// source: Storage.proto

package storage

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	Storage_CreateTask_FullMethodName       = "/Storage.Storage/CreateTask"
	Storage_DeleteTask_FullMethodName       = "/Storage.Storage/DeleteTask"
	Storage_UpdateTask_FullMethodName       = "/Storage.Storage/UpdateTask"
	Storage_ViewTask_FullMethodName         = "/Storage.Storage/ViewTask"
	Storage_ListTasks_FullMethodName        = "/Storage.Storage/ListTasks"
	Storage_BatchDeleteTasks_FullMethodName = "/Storage.Storage/BatchDeleteTasks"
)

// StorageClient is the client API for Storage service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type StorageClient interface {
	// 添加新任务
	CreateTask(ctx context.Context, in *CreateTaskRequest, opts ...grpc.CallOption) (*CreateTaskResponse, error)
	// 删除任务
	DeleteTask(ctx context.Context, in *DeleteTaskRequest, opts ...grpc.CallOption) (*OperationResponse, error)
	// 更新任务
	UpdateTask(ctx context.Context, in *UpdateTaskRequest, opts ...grpc.CallOption) (*OperationResponse, error)
	// 查看单个任务详情
	ViewTask(ctx context.Context, in *ViewTaskRequest, opts ...grpc.CallOption) (*Task, error)
	// 列出所有任务
	ListTasks(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*TaskList, error)
	// 批量删除任务
	BatchDeleteTasks(ctx context.Context, in *BatchDeleteTasksRequest, opts ...grpc.CallOption) (*OperationResponse, error)
}

type storageClient struct {
	cc grpc.ClientConnInterface
}

func NewStorageClient(cc grpc.ClientConnInterface) StorageClient {
	return &storageClient{cc}
}

func (c *storageClient) CreateTask(ctx context.Context, in *CreateTaskRequest, opts ...grpc.CallOption) (*CreateTaskResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CreateTaskResponse)
	err := c.cc.Invoke(ctx, Storage_CreateTask_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *storageClient) DeleteTask(ctx context.Context, in *DeleteTaskRequest, opts ...grpc.CallOption) (*OperationResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(OperationResponse)
	err := c.cc.Invoke(ctx, Storage_DeleteTask_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *storageClient) UpdateTask(ctx context.Context, in *UpdateTaskRequest, opts ...grpc.CallOption) (*OperationResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(OperationResponse)
	err := c.cc.Invoke(ctx, Storage_UpdateTask_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *storageClient) ViewTask(ctx context.Context, in *ViewTaskRequest, opts ...grpc.CallOption) (*Task, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Task)
	err := c.cc.Invoke(ctx, Storage_ViewTask_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *storageClient) ListTasks(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*TaskList, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(TaskList)
	err := c.cc.Invoke(ctx, Storage_ListTasks_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *storageClient) BatchDeleteTasks(ctx context.Context, in *BatchDeleteTasksRequest, opts ...grpc.CallOption) (*OperationResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(OperationResponse)
	err := c.cc.Invoke(ctx, Storage_BatchDeleteTasks_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// StorageServer is the server API for Storage service.
// All implementations must embed UnimplementedStorageServer
// for forward compatibility.
type StorageServer interface {
	// 添加新任务
	CreateTask(context.Context, *CreateTaskRequest) (*CreateTaskResponse, error)
	// 删除任务
	DeleteTask(context.Context, *DeleteTaskRequest) (*OperationResponse, error)
	// 更新任务
	UpdateTask(context.Context, *UpdateTaskRequest) (*OperationResponse, error)
	// 查看单个任务详情
	ViewTask(context.Context, *ViewTaskRequest) (*Task, error)
	// 列出所有任务
	ListTasks(context.Context, *Empty) (*TaskList, error)
	// 批量删除任务
	BatchDeleteTasks(context.Context, *BatchDeleteTasksRequest) (*OperationResponse, error)
	mustEmbedUnimplementedStorageServer()
}

// UnimplementedStorageServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedStorageServer struct{}

func (UnimplementedStorageServer) CreateTask(context.Context, *CreateTaskRequest) (*CreateTaskResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateTask not implemented")
}
func (UnimplementedStorageServer) DeleteTask(context.Context, *DeleteTaskRequest) (*OperationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteTask not implemented")
}
func (UnimplementedStorageServer) UpdateTask(context.Context, *UpdateTaskRequest) (*OperationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateTask not implemented")
}
func (UnimplementedStorageServer) ViewTask(context.Context, *ViewTaskRequest) (*Task, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ViewTask not implemented")
}
func (UnimplementedStorageServer) ListTasks(context.Context, *Empty) (*TaskList, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListTasks not implemented")
}
func (UnimplementedStorageServer) BatchDeleteTasks(context.Context, *BatchDeleteTasksRequest) (*OperationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BatchDeleteTasks not implemented")
}
func (UnimplementedStorageServer) mustEmbedUnimplementedStorageServer() {}
func (UnimplementedStorageServer) testEmbeddedByValue()                 {}

// UnsafeStorageServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to StorageServer will
// result in compilation errors.
type UnsafeStorageServer interface {
	mustEmbedUnimplementedStorageServer()
}

func RegisterStorageServer(s grpc.ServiceRegistrar, srv StorageServer) {
	// If the following call pancis, it indicates UnimplementedStorageServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&Storage_ServiceDesc, srv)
}

func _Storage_CreateTask_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateTaskRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StorageServer).CreateTask(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Storage_CreateTask_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StorageServer).CreateTask(ctx, req.(*CreateTaskRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Storage_DeleteTask_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteTaskRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StorageServer).DeleteTask(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Storage_DeleteTask_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StorageServer).DeleteTask(ctx, req.(*DeleteTaskRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Storage_UpdateTask_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateTaskRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StorageServer).UpdateTask(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Storage_UpdateTask_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StorageServer).UpdateTask(ctx, req.(*UpdateTaskRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Storage_ViewTask_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ViewTaskRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StorageServer).ViewTask(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Storage_ViewTask_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StorageServer).ViewTask(ctx, req.(*ViewTaskRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Storage_ListTasks_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StorageServer).ListTasks(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Storage_ListTasks_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StorageServer).ListTasks(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _Storage_BatchDeleteTasks_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BatchDeleteTasksRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StorageServer).BatchDeleteTasks(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Storage_BatchDeleteTasks_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StorageServer).BatchDeleteTasks(ctx, req.(*BatchDeleteTasksRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Storage_ServiceDesc is the grpc.ServiceDesc for Storage service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Storage_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "Storage.Storage",
	HandlerType: (*StorageServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateTask",
			Handler:    _Storage_CreateTask_Handler,
		},
		{
			MethodName: "DeleteTask",
			Handler:    _Storage_DeleteTask_Handler,
		},
		{
			MethodName: "UpdateTask",
			Handler:    _Storage_UpdateTask_Handler,
		},
		{
			MethodName: "ViewTask",
			Handler:    _Storage_ViewTask_Handler,
		},
		{
			MethodName: "ListTasks",
			Handler:    _Storage_ListTasks_Handler,
		},
		{
			MethodName: "BatchDeleteTasks",
			Handler:    _Storage_BatchDeleteTasks_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "Storage.proto",
}
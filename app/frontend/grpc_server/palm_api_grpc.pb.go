// Code generated by protoc-gen-go-grpc_server. DO NOT EDIT.

package grpc_server

import (
	context "context"
	empty "github.com/golang/protobuf/ptypes/empty"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// AccountsClient is the client API for Accounts service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type AccountsClient interface {
	Login(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*LoginResult, error)
}

type accountsClient struct {
	cc grpc.ClientConnInterface
}

func NewAccountsClient(cc grpc.ClientConnInterface) AccountsClient {
	return &accountsClient{cc}
}

func (c *accountsClient) Login(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*LoginResult, error) {
	out := new(LoginResult)
	err := c.cc.Invoke(ctx, "/frontend.Accounts/login", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AccountsServer is the server API for Accounts service.
// All implementations must embed UnimplementedAccountsServer
// for forward compatibility
type AccountsServer interface {
	Login(context.Context, *empty.Empty) (*LoginResult, error)
	mustEmbedUnimplementedAccountsServer()
}

// UnimplementedAccountsServer must be embedded to have forward compatible implementations.
type UnimplementedAccountsServer struct {
}

func (UnimplementedAccountsServer) Login(context.Context, *empty.Empty) (*LoginResult, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Login not implemented")
}
func (UnimplementedAccountsServer) mustEmbedUnimplementedAccountsServer() {}

// UnsafeAccountsServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to AccountsServer will
// result in compilation errors.
type UnsafeAccountsServer interface {
	mustEmbedUnimplementedAccountsServer()
}

func RegisterAccountsServer(s grpc.ServiceRegistrar, srv AccountsServer) {
	s.RegisterService(&Accounts_ServiceDesc, srv)
}

func _Accounts_Login_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(empty.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AccountsServer).Login(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/frontend.Accounts/login",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AccountsServer).Login(ctx, req.(*empty.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

// Accounts_ServiceDesc is the grpc_server.ServiceDesc for Accounts service.
// It's only intended for direct use with grpc_server.RegisterService,
// and not to be introspected or modified (even as a copy)
var Accounts_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "frontend.Accounts",
	HandlerType: (*AccountsServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "login",
			Handler:    _Accounts_Login_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "palm_api.proto",
}

// ContactsClient is the client API for Contacts service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ContactsClient interface {
	CreateOrUpdate(ctx context.Context, in *Contact, opts ...grpc.CallOption) (*ContactResult, error)
	Delete(ctx context.Context, in *Contact, opts ...grpc.CallOption) (*ContactResult, error)
	Search(ctx context.Context, in *Contact, opts ...grpc.CallOption) (*ContactListResult, error)
}

type contactsClient struct {
	cc grpc.ClientConnInterface
}

func NewContactsClient(cc grpc.ClientConnInterface) ContactsClient {
	return &contactsClient{cc}
}

func (c *contactsClient) CreateOrUpdate(ctx context.Context, in *Contact, opts ...grpc.CallOption) (*ContactResult, error) {
	out := new(ContactResult)
	err := c.cc.Invoke(ctx, "/frontend.Contacts/createOrUpdate", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *contactsClient) Delete(ctx context.Context, in *Contact, opts ...grpc.CallOption) (*ContactResult, error) {
	out := new(ContactResult)
	err := c.cc.Invoke(ctx, "/frontend.Contacts/delete", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *contactsClient) Search(ctx context.Context, in *Contact, opts ...grpc.CallOption) (*ContactListResult, error) {
	out := new(ContactListResult)
	err := c.cc.Invoke(ctx, "/frontend.Contacts/search", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ContactsServer is the server API for Contacts service.
// All implementations must embed UnimplementedContactsServer
// for forward compatibility
type ContactsServer interface {
	CreateOrUpdate(context.Context, *Contact) (*ContactResult, error)
	Delete(context.Context, *Contact) (*ContactResult, error)
	Search(context.Context, *Contact) (*ContactListResult, error)
	mustEmbedUnimplementedContactsServer()
}

// UnimplementedContactsServer must be embedded to have forward compatible implementations.
type UnimplementedContactsServer struct {
}

func (UnimplementedContactsServer) CreateOrUpdate(context.Context, *Contact) (*ContactResult, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateOrUpdate not implemented")
}
func (UnimplementedContactsServer) Delete(context.Context, *Contact) (*ContactResult, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Delete not implemented")
}
func (UnimplementedContactsServer) Search(context.Context, *Contact) (*ContactListResult, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Search not implemented")
}
func (UnimplementedContactsServer) mustEmbedUnimplementedContactsServer() {}

// UnsafeContactsServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ContactsServer will
// result in compilation errors.
type UnsafeContactsServer interface {
	mustEmbedUnimplementedContactsServer()
}

func RegisterContactsServer(s grpc.ServiceRegistrar, srv ContactsServer) {
	s.RegisterService(&Contacts_ServiceDesc, srv)
}

func _Contacts_CreateOrUpdate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Contact)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ContactsServer).CreateOrUpdate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/frontend.Contacts/createOrUpdate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ContactsServer).CreateOrUpdate(ctx, req.(*Contact))
	}
	return interceptor(ctx, in, info, handler)
}

func _Contacts_Delete_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Contact)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ContactsServer).Delete(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/frontend.Contacts/delete",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ContactsServer).Delete(ctx, req.(*Contact))
	}
	return interceptor(ctx, in, info, handler)
}

func _Contacts_Search_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Contact)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ContactsServer).Search(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/frontend.Contacts/search",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ContactsServer).Search(ctx, req.(*Contact))
	}
	return interceptor(ctx, in, info, handler)
}

// Contacts_ServiceDesc is the grpc_server.ServiceDesc for Contacts service.
// It's only intended for direct use with grpc_server.RegisterService,
// and not to be introspected or modified (even as a copy)
var Contacts_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "frontend.Contacts",
	HandlerType: (*ContactsServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "createOrUpdate",
			Handler:    _Contacts_CreateOrUpdate_Handler,
		},
		{
			MethodName: "delete",
			Handler:    _Contacts_Delete_Handler,
		},
		{
			MethodName: "search",
			Handler:    _Contacts_Search_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "palm_api.proto",
}

// TasksClient is the client API for Tasks service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TasksClient interface {
	Delete(ctx context.Context, in *Task, opts ...grpc.CallOption) (*TaskResult, error)
	Search(ctx context.Context, in *Task, opts ...grpc.CallOption) (*TaskListResult, error)
	CreateOrUpdate(ctx context.Context, in *Task, opts ...grpc.CallOption) (*TaskResult, error)
}

type tasksClient struct {
	cc grpc.ClientConnInterface
}

func NewTasksClient(cc grpc.ClientConnInterface) TasksClient {
	return &tasksClient{cc}
}

func (c *tasksClient) Delete(ctx context.Context, in *Task, opts ...grpc.CallOption) (*TaskResult, error) {
	out := new(TaskResult)
	err := c.cc.Invoke(ctx, "/frontend.Tasks/delete", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tasksClient) Search(ctx context.Context, in *Task, opts ...grpc.CallOption) (*TaskListResult, error) {
	out := new(TaskListResult)
	err := c.cc.Invoke(ctx, "/frontend.Tasks/search", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tasksClient) CreateOrUpdate(ctx context.Context, in *Task, opts ...grpc.CallOption) (*TaskResult, error) {
	out := new(TaskResult)
	err := c.cc.Invoke(ctx, "/frontend.Tasks/createOrUpdate", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TasksServer is the server API for Tasks service.
// All implementations must embed UnimplementedTasksServer
// for forward compatibility
type TasksServer interface {
	Delete(context.Context, *Task) (*TaskResult, error)
	Search(context.Context, *Task) (*TaskListResult, error)
	CreateOrUpdate(context.Context, *Task) (*TaskResult, error)
	mustEmbedUnimplementedTasksServer()
}

// UnimplementedTasksServer must be embedded to have forward compatible implementations.
type UnimplementedTasksServer struct {
}

func (UnimplementedTasksServer) Delete(context.Context, *Task) (*TaskResult, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Delete not implemented")
}
func (UnimplementedTasksServer) Search(context.Context, *Task) (*TaskListResult, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Search not implemented")
}
func (UnimplementedTasksServer) CreateOrUpdate(context.Context, *Task) (*TaskResult, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateOrUpdate not implemented")
}
func (UnimplementedTasksServer) mustEmbedUnimplementedTasksServer() {}

// UnsafeTasksServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TasksServer will
// result in compilation errors.
type UnsafeTasksServer interface {
	mustEmbedUnimplementedTasksServer()
}

func RegisterTasksServer(s grpc.ServiceRegistrar, srv TasksServer) {
	s.RegisterService(&Tasks_ServiceDesc, srv)
}

func _Tasks_Delete_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Task)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TasksServer).Delete(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/frontend.Tasks/delete",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TasksServer).Delete(ctx, req.(*Task))
	}
	return interceptor(ctx, in, info, handler)
}

func _Tasks_Search_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Task)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TasksServer).Search(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/frontend.Tasks/search",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TasksServer).Search(ctx, req.(*Task))
	}
	return interceptor(ctx, in, info, handler)
}

func _Tasks_CreateOrUpdate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Task)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TasksServer).CreateOrUpdate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/frontend.Tasks/createOrUpdate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TasksServer).CreateOrUpdate(ctx, req.(*Task))
	}
	return interceptor(ctx, in, info, handler)
}

// Tasks_ServiceDesc is the grpc_server.ServiceDesc for Tasks service.
// It's only intended for direct use with grpc_server.RegisterService,
// and not to be introspected or modified (even as a copy)
var Tasks_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "frontend.Tasks",
	HandlerType: (*TasksServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "delete",
			Handler:    _Tasks_Delete_Handler,
		},
		{
			MethodName: "search",
			Handler:    _Tasks_Search_Handler,
		},
		{
			MethodName: "createOrUpdate",
			Handler:    _Tasks_CreateOrUpdate_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "palm_api.proto",
}

package main

import (
	context "context"
	"net/http"

	v1 "github.com/MikkelHJuul/ld/gen/ld/v1"
	connect_v1 "github.com/MikkelHJuul/ld/gen/ld/v1/ldv1connect"
	connect_go "github.com/bufbuild/connect-go"
)

type handlerGen func(string, connect_v1.LdServiceHandler, ...connect_go.HandlerOption) http.Handler

type adminService struct {
	handler       Handler
	dbProvisioner DbProvisioner
	registry      protoRegistry

	handlerOpts []connect_go.HandlerOption
}

var _ connect_v1.AdminServiceHandler = (*adminService)(nil)

type DBMeta struct {
	Name    string
	Methods method
}

type DbProvisioner interface {
	Create(meta DBMeta) (connect_v1.LdServiceHandler, error)
	Alter(new DBMeta) (handler connect_v1.LdServiceHandler, err error)
	Delete(dbName string) error
}

type protoRegistry interface {
	Create(*v1.RegistryInfo) error
	Alter(*v1.RegistryInfo) error
	Delete(*v1.RegistryInfo) error
	Get(string) (*v1.RegistryInfo, error)
	All() ([]*v1.RegistryInfo, error)
}

// Register implements ldv1connect.AdminServiceHandler
func (a *adminService) Register(ctx context.Context, req *connect_go.Request[v1.RegisterRequest]) (*connect_go.Response[v1.RegisterResponse], error) {
	methodGenerators, meth := methodsFrom(req.Msg.Entry.Methods)
	methods, action, err := a.handleDatabaseRegistering(req.Msg, methodGenerators, meth)
	if err != nil {
		return nil, err // v1.RegisterResponse
	}

	err = a.handler.Handle(RoutingRequest{
		Action: action,
		Routes: Routing{
			BaseName: req.Msg.Entry.ProtoName,
			Methods:  methods,
		},
	})
	if err != nil {
		return nil, err // v1.RegisterResponse
	}

	return connect_go.NewResponse(&v1.RegisterResponse{}), nil
}

func (a *adminService) handleDatabaseRegistering(req *v1.RegisterRequest, methodGenerators map[method]handlerGen, allMethods method) (mHandlers map[method]http.Handler, action Action, err error) {
	var handler connect_v1.LdServiceHandler
	switch req.Type {
	case v1.RegisterRequest_REMOVE:
		action = REMOVE
		err = a.dbProvisioner.Delete(req.Entry.ProtoName)
		if err != nil {
			return
		}
		err = a.registry.Delete(req.Entry)
		return
	case v1.RegisterRequest_ALTER:
		action = ALTER
		handler, err = a.dbProvisioner.Alter(DBMeta{req.Entry.ProtoName, allMethods})
		if err != nil {
			return
		}
		err = a.registry.Alter(req.Entry)
	case v1.RegisterRequest_ADD, v1.RegisterRequest_UNSPECIFIED:
		action = ADD
		handler, err = a.dbProvisioner.Create(DBMeta{req.Entry.ProtoName, allMethods})
		if err != nil {
			return
		}
		err = a.registry.Create(req.Entry)
	}
	mHandlers = make(map[method]http.Handler, len(methodGenerators))
	for k, v := range methodGenerators {
		mHandlers[k] = v(req.Entry.ProtoName, handler, a.handlerOpts...)
	}
	return
}

func methodsFrom(reqMethods []v1.RegistryInfo_Method) (map[method]handlerGen, method) {
	methods := 0
	for _, m := range reqMethods {
		methods += int(m.Number())
	}
	if methods == 0 {
		methods = int(All)
	}
	handlers := make(map[method]handlerGen)
	for m, h := range handlerForMethod {
		if int(m)&methods == int(m) {
			handlers[m] = h
		}
	}
	return handlers, method(methods)
}

// Registry implements ldv1connect.AdminServiceHandler
func (a *adminService) Registry(context.Context, *connect_go.Request[v1.RegistryRequest]) (*connect_go.Response[v1.RegistryResponse], error) {
	regs, err := a.registry.All()
	if err != nil {
		return nil, err
	}
	return connect_go.NewResponse(&v1.RegistryResponse{Registries: regs}), nil
}

// RegistryInfo implements ldv1connect.AdminServiceHandler
func (a *adminService) RegistryInfo(_ context.Context, req *connect_go.Request[v1.RegistryInfoRequest]) (*connect_go.Response[v1.RegistryInfoResponse], error) {
	reg, err := a.registry.Get(req.Msg.ProtoName)
	if err != nil {
		return nil, err
	}
	return connect_go.NewResponse(&v1.RegistryInfoResponse{Entry: reg}), nil
}

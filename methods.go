package main

import (
	"fmt"
	"net/http"

	v1 "github.com/MikkelHJuul/ld/gen/ld/v1/ldv1connect"
	connect_go "github.com/bufbuild/connect-go"
)

type method int

const All method = 255

const (
	Get method = 1 << iota
	GetMany
	GetRange
	Set
	SetMany
	Delete
	DeleteMany
	DeleteRange
)

var names = [...]string{
	"Get",
	"GetMany",
	"GetRange",
	"Set",
	"SetMany",
	"Delete",
	"DeleteMany",
	"DeleteRange",
}

var handlerForMethod = map[method]handlerGen{
	Set: func(serviceName string, svc v1.LdServiceHandler, opts ...connect_go.HandlerOption) http.Handler {
		return connect_go.NewUnaryHandler(
			"/"+serviceName+"/Set",
			svc.Set,
			opts...,
		)
	},
	SetMany: func(serviceName string, svc v1.LdServiceHandler, opts ...connect_go.HandlerOption) http.Handler {
		return connect_go.NewBidiStreamHandler(
			"/"+serviceName+"/SetMany",
			svc.SetMany,
			opts...,
		)
	},
	Get: func(serviceName string, svc v1.LdServiceHandler, opts ...connect_go.HandlerOption) http.Handler {
		return connect_go.NewUnaryHandler(
			"/"+serviceName+"/Get",
			svc.Get,
			opts...,
		)
	},
	GetMany: func(serviceName string, svc v1.LdServiceHandler, opts ...connect_go.HandlerOption) http.Handler {
		return connect_go.NewBidiStreamHandler(
			"/"+serviceName+"/GetMany",
			svc.GetMany,
			opts...,
		)
	},
	GetRange: func(serviceName string, svc v1.LdServiceHandler, opts ...connect_go.HandlerOption) http.Handler {
		return connect_go.NewServerStreamHandler(
			"/"+serviceName+"/GetRange",
			svc.GetRange,
			opts...,
		)
	},
	Delete: func(serviceName string, svc v1.LdServiceHandler, opts ...connect_go.HandlerOption) http.Handler {
		return connect_go.NewUnaryHandler(
			"/"+serviceName+"/Delete",
			svc.Delete,
			opts...,
		)
	},
	DeleteMany: func(serviceName string, svc v1.LdServiceHandler, opts ...connect_go.HandlerOption) http.Handler {
		return connect_go.NewBidiStreamHandler(
			"/"+serviceName+"/DeleteMany",
			svc.DeleteMany,
			opts...,
		)
	},
	DeleteRange: func(serviceName string, svc v1.LdServiceHandler, opts ...connect_go.HandlerOption) http.Handler {
		return connect_go.NewServerStreamHandler(
			"/"+serviceName+"/DeleteRange",
			svc.DeleteRange,
			opts...,
		)
	},
}

func (m method) String() string {
	return names[m]
}

func MethodFromText(name string) (method, error) {
	for i, n := range names {
		if n == name {
			return method(1 << i), nil
		}
	}
	return 0, fmt.Errorf("no such method, %s", name)
}

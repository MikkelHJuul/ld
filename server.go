package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
)

type Routing struct {
	BaseName string
	Methods  map[method]http.Handler
}

type Action int

const (
	_ Action = iota
	ADD
	ALTER
	REMOVE
)

type RoutingRequest struct {
	Action Action
	Routes Routing
}

type ldMux struct {
	mu       sync.RWMutex
	handlers map[string]map[method]http.Handler
}

type Handler interface {
	Handle(RoutingRequest) error
}

var _ http.Handler = (*ldMux)(nil)
var _ Handler = (*ldMux)(nil)

func NewInternalLdMux(routeRequestChan <-chan RoutingRequest) *ldMux {
	h := make(map[string]map[method]http.Handler)
	mux := &ldMux{handlers: h}
	return mux
}

func (lm *ldMux) Handle(routeReq RoutingRequest) error {
	switch routeReq.Action {
	case ADD:
		return lm.addHandle(routeReq.Routes)
	case ALTER:
		return lm.alterHandle(routeReq.Routes)
	case REMOVE:
		return lm.removeHandle(routeReq.Routes)
	default:
		return fmt.Errorf("Unkown Action, %v", routeReq.Action)
	}

}

func (lm *ldMux) addHandle(route Routing) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if _, ok := lm.handlers[route.BaseName]; ok {
		return fmt.Errorf("%s already exist", route.BaseName)
	}

	if route.Methods == nil {
		return fmt.Errorf("routing must not be nil, in %v", route)
	}

	lm.handlers[route.BaseName] = route.Methods
	return nil
}

func (lm *ldMux) removeHandle(route Routing) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if _, ok := lm.handlers[route.BaseName]; !ok {
		return fmt.Errorf("%s does not exist", route.BaseName)
	}
	delete(lm.handlers, route.BaseName)
	return nil
}

func (lm *ldMux) alterHandle(route Routing) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if _, ok := lm.handlers[route.BaseName]; !ok {
		return fmt.Errorf("%s does not exist", route.BaseName)
	}
	lm.handlers[route.BaseName] = route.Methods
	return nil
}

func (lm *ldMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	base, meth, ok := strings.Cut(r.RequestURI, "/")

	notFound := http.NotFoundHandler()

	if !ok {
		notFound.ServeHTTP(w, r)
		return
	}

	m, err := MethodFromText(meth)
	if err != nil {
		notFound.ServeHTTP(w, r) //probably not 404?
		return
	}

	lm.mu.RLock()

	methods, ok := lm.handlers[base]

	if !ok {
		lm.mu.RUnlock()
		notFound.ServeHTTP(w, r)
		return
	}

	h, ok := methods[m]
	lm.mu.RUnlock()
	if !ok {
		notFound.ServeHTTP(w, r)
		return
	}

	h.ServeHTTP(w, r)
}

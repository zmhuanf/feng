package router

import (
	"reflect"
	"strings"
	"sync"
)

type Middleware struct {
	Route string
	Fn    any
}

type Router struct {
	contextType reflect.Type
	routes      map[string]any
	middlewares []Middleware
	lock        sync.RWMutex
}

func New(contextType reflect.Type) *Router {
	return &Router{
		contextType: contextType,
		routes:      make(map[string]any),
		middlewares: make([]Middleware, 0),
	}
}

func (r *Router) Handle(route string, fn any) error {
	if err := CheckHandler(fn, r.contextType); err != nil {
		return err
	}
	r.lock.Lock()
	defer r.lock.Unlock()
	r.routes[route] = fn
	return nil
}

func (r *Router) Use(route string, fn any) error {
	if err := CheckHandler(fn, r.contextType); err != nil {
		return err
	}
	r.lock.Lock()
	defer r.lock.Unlock()
	r.middlewares = append(r.middlewares, Middleware{Route: route, Fn: fn})
	return nil
}

func (r *Router) Handler(route string) (any, bool) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	fn, ok := r.routes[route]
	return fn, ok
}

func (r *Router) Middlewares(route string) []Middleware {
	r.lock.RLock()
	defer r.lock.RUnlock()

	matched := make([]Middleware, 0, len(r.middlewares))
	for _, middleware := range r.middlewares {
		if strings.HasPrefix(route, middleware.Route) {
			matched = append(matched, middleware)
		}
	}
	return matched
}

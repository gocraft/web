package web

import (
	"reflect"
	"strings"
)

type HttpMethod string

const (
	HttpMethodGet    = HttpMethod("GET")
	HttpMethodPost   = HttpMethod("POST")
	HttpMethodPut    = HttpMethod("PUT")
	HttpMethodDelete = HttpMethod("DELETE")
	HttpMethodPatch  = HttpMethod("PATCH")
)

var HttpMethods = []HttpMethod{HttpMethodGet, HttpMethodPost, HttpMethodPut, HttpMethodDelete, HttpMethodPatch}

type Router struct {
	// Hierarchy:
	parent           *Router // nil if root router.
	children         []*Router
	maxChildrenDepth int

	// For each request we'll create one of these objects
	contextType reflect.Type

	// Eg, "/" or "/admin". Any routes added to this router will be prefixed with this.
	pathPrefix string

	// Routeset contents:
	middleware []*middlewareHandler
	routes     []*Route

	// The root pathnode is the same for a tree of Routers
	root map[HttpMethod]*PathNode

	// This can can be set on any router. The target's ErrorHandler will be invoked if it exists
	errorHandler reflect.Value

	// This can only be set on the root handler, since by virtue of not finding a route, we don't have a target.
	// (That being said, in the future we could investigate namespace matches)
	notFoundHandler reflect.Value
}

type Route struct {
	Router  *Router
	Method  HttpMethod
	Path    string
	Handler *actionHandler
}

type middlewareHandler struct {
	Generic           bool
	DynamicMiddleware reflect.Value
	GenericMiddleware GenericMiddleware
}

type actionHandler struct {
	Generic        bool
	DynamicHandler reflect.Value
	GenericHandler GenericHandler
}

type NextMiddlewareFunc func(ResponseWriter, *Request)
type GenericMiddleware func(ResponseWriter, *Request, NextMiddlewareFunc)
type GenericHandler func(ResponseWriter, *Request)

func New(ctx interface{}) *Router {
	validateContext(ctx, nil)

	r := &Router{}
	r.contextType = reflect.TypeOf(ctx)
	r.pathPrefix = "/"
	r.maxChildrenDepth = 1
	r.root = make(map[HttpMethod]*PathNode)
	for _, method := range HttpMethods {
		r.root[method] = newPathNode()
	}
	return r
}

func NewWithPrefix(ctx interface{}, pathPrefix string) *Router {
	r := New(ctx)
	r.pathPrefix = pathPrefix

	return r
}

func (r *Router) Subrouter(ctx interface{}, pathPrefix string) *Router {
	validateContext(ctx, r.contextType)

	// Create new router, link up hierarchy
	newRouter := &Router{parent: r}
	r.children = append(r.children, newRouter)

	// Increment maxChildrenDepth if this is the first child of the router
	if len(r.children) == 1 {
		curParent := r
		for curParent != nil {
			curParent.maxChildrenDepth = curParent.depth()
			curParent = curParent.parent
		}
	}

	newRouter.contextType = reflect.TypeOf(ctx)
	newRouter.pathPrefix = appendPath(r.pathPrefix, pathPrefix)
	newRouter.root = r.root

	return newRouter
}

func (r *Router) Middleware(fn interface{}) *Router {
	vfn := reflect.ValueOf(fn)
	validateMiddleware(vfn, r.contextType)
	if vfn.Type().NumIn() == 3 {
		r.middleware = append(r.middleware, &middlewareHandler{Generic: true, GenericMiddleware: fn.(func(ResponseWriter, *Request, NextMiddlewareFunc))})
	} else {
		r.middleware = append(r.middleware, &middlewareHandler{Generic: false, DynamicMiddleware: vfn})
	}

	return r
}

func (r *Router) Error(fn interface{}) {
	vfn := reflect.ValueOf(fn)
	validateErrorHandler(vfn, r.contextType)
	r.errorHandler = vfn
}

func (r *Router) NotFound(fn interface{}) {
	if r.parent != nil {
		panic("You can only set a NotFoundHandler on the root router.")
	}
	vfn := reflect.ValueOf(fn)
	validateNotFoundHandler(vfn, r.contextType)
	r.notFoundHandler = vfn
}

func (r *Router) Get(path string, fn interface{}) *Router {
	return r.addRoute(HttpMethodGet, path, fn)
}

func (r *Router) Post(path string, fn interface{}) *Router {
	return r.addRoute(HttpMethodPost, path, fn)
}

func (r *Router) Put(path string, fn interface{}) *Router {
	return r.addRoute(HttpMethodPut, path, fn)
}

func (r *Router) Delete(path string, fn interface{}) *Router {
	return r.addRoute(HttpMethodDelete, path, fn)
}

func (r *Router) Patch(path string, fn interface{}) *Router {
	return r.addRoute(HttpMethodPatch, path, fn)
}

//
//
//
func (r *Router) addRoute(method HttpMethod, path string, fn interface{}) *Router {
	vfn := reflect.ValueOf(fn)
	validateHandler(vfn, r.contextType)
	fullPath := appendPath(r.pathPrefix, path)
	route := &Route{Method: method, Path: fullPath, Router: r}
	if vfn.Type().NumIn() == 2 {
		route.Handler = &actionHandler{Generic: true, GenericHandler: fn.(func(ResponseWriter, *Request))}
	} else {
		route.Handler = &actionHandler{Generic: false, DynamicHandler: vfn}
	}
	r.routes = append(r.routes, route)
	r.root[method].add(fullPath, route)
	return r
}

// Calculates the max child depth of the node. Leaves return 1. For Parent->Child, Parent is 2.
func (r *Router) depth() int {
	max := 0
	for _, child := range r.children {
		childDepth := child.depth()
		if childDepth > max {
			max = childDepth
		}
	}
	return max + 1
}

//
// Private methods:
//

// Panics unless validation is correct
func validateContext(ctx interface{}, parentCtxType reflect.Type) {
	ctxType := reflect.TypeOf(ctx)

	if ctxType.Kind() != reflect.Struct {
		panic("web: Context needs to be a struct type")
	}

	if parentCtxType != nil && parentCtxType != ctxType {
		if ctxType.NumField() == 0 {
			panic("web: Context needs to have first field be a pointer to parent context")
		}

		fldType := ctxType.Field(0).Type

		// Ensure fld is a pointer to parentCtxType
		if fldType != reflect.PtrTo(parentCtxType) {
			panic("web: Context needs to have first field be a pointer to parent context")
		}
	}
}

// Panics unless fn is a proper handler wrt ctxType
// eg, func(ctx *ctxType, writer, request)
func validateHandler(vfn reflect.Value, ctxType reflect.Type) {
	var req *Request
	var resp func() ResponseWriter
	if !isValidHandler(vfn, ctxType, reflect.TypeOf(resp).Out(0), reflect.TypeOf(req)) {
		panic(instructiveMessage(vfn, "a handler", "handler", "rw web.ResponseWriter, req *web.Request", ctxType))
	}
}

func validateErrorHandler(vfn reflect.Value, ctxType reflect.Type) {
	var req *Request
	var resp func() ResponseWriter
	var interfaceType func() interface{} // This is weird. I need to get an interface{} reflect.Type; var x interface{}; TypeOf(x) doesn't work, because it returns nil
	if !isValidHandler(vfn, ctxType, reflect.TypeOf(resp).Out(0), reflect.TypeOf(req), reflect.TypeOf(interfaceType).Out(0)) {
		panic(instructiveMessage(vfn, "an error handler", "error handler", "rw web.ResponseWriter, req *web.Request, err interface{}", ctxType))
	}
}

func validateNotFoundHandler(vfn reflect.Value, ctxType reflect.Type) {
	var req *Request
	var resp func() ResponseWriter
	if !isValidHandler(vfn, ctxType, reflect.TypeOf(resp).Out(0), reflect.TypeOf(req)) {
		panic(instructiveMessage(vfn, "a 'not found' handler", "not found handler", "rw web.ResponseWriter, req *web.Request", ctxType))
	}
}

func validateMiddleware(vfn reflect.Value, ctxType reflect.Type) {
	var req *Request
	var resp func() ResponseWriter
	var n NextMiddlewareFunc
	if !isValidHandler(vfn, ctxType, reflect.TypeOf(resp).Out(0), reflect.TypeOf(req), reflect.TypeOf(n)) {
		panic(instructiveMessage(vfn, "middleware", "middleware", "rw web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc", ctxType))
	}
}

// Ensures vfn is a function, that optionally takes a *ctxType as the first argument, followed by the specified types. Handlers have no return value.
// Returns true if valid, false otherwise.
func isValidHandler(vfn reflect.Value, ctxType reflect.Type, types ...reflect.Type) bool {
	fnType := vfn.Type()

	if fnType.Kind() != reflect.Func {
		return false
	}

	typesStartIdx := 0
	typesLen := len(types)
	numIn := fnType.NumIn()
	numOut := fnType.NumOut()

	if numOut != 0 {
		return false
	}

	if numIn == typesLen {
		// No context
	} else if numIn == (typesLen + 1) {
		// context, types
		if fnType.In(0) != reflect.PtrTo(ctxType) {
			return false
		}
		typesStartIdx = 1
	} else {
		return false
	}

	for _, typeArg := range types {
		if fnType.In(typesStartIdx) != typeArg {
			return false
		}
		typesStartIdx += 1
	}

	return true
}

// Since it's easy to pass the wrong method to a middleware/handler route, and since the user can't rely on static type checking since we use reflection,
// lets be super helpful about what they did and what they need to do.
// Arguments:
//  - vfn is the failed method
//  - addingType is for "You are adding {addingType} to a router...". Eg, "middleware" or "a handler" or "an error handler"
//  - yourType is for "Your {yourType} function can have...". Eg, "middleware" or "handler" or "error handler"
//  - args is like "rw web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc"
//    - NOTE: args can be calculated if you pass in each type. BUT, it doesn't have example argument name, so it has less copy/paste value.
func instructiveMessage(vfn reflect.Value, addingType string, yourType string, args string, ctxType reflect.Type) string {
	// Get context type without package.
	ctxString := ctxType.String()
	splitted := strings.Split(ctxString, ".")
	if len(splitted) <= 1 {
		ctxString = splitted[0]
	} else {
		ctxString = splitted[1]
	}

	str := "\n" + strings.Repeat("*", 120) + "\n"
	str += "* You are adding " + addingType + " to a router with context type '" + ctxString + "'\n"
	str += "*\n*\n"
	str += "* Your " + yourType + " function can have one of these signatures:\n"
	str += "*\n"
	str += "* // If you don't need context:\n"
	str += "* func YourFunctionName(" + args + ")\n"
	str += "*\n"
	str += "* // If you want your " + yourType + " to accept a context:\n"
	str += "* func (c *" + ctxString + ") YourFunctionName(" + args + ")  // or,\n"
	str += "* func YourFunctionName(c *" + ctxString + ", " + args + ")\n"
	str += "*\n"
	str += "* Unfortunately, your function has this signature: " + vfn.Type().String() + "\n"
	str += "*\n"
	str += strings.Repeat("*", 120) + "\n"

	return str
}

// Both rootPath/childPath are like "/" and "/users"
// Assumption is that both are well-formed paths.
func appendPath(rootPath, childPath string) string {
	if rootPath == "/" {
		return childPath
	}

	return rootPath + childPath
}

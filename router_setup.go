package web

import (
  "reflect"
  "fmt"
)

type HttpMethod string
const (
  HttpMethodGet = HttpMethod("GET")
  HttpMethodPost = HttpMethod("POST")
  HttpMethodPut = HttpMethod("PUT")
  HttpMethodDelete = HttpMethod("DELETE")
  HttpMethodPatch = HttpMethod("PATCH")
)
var HttpMethods = []HttpMethod{HttpMethodGet, HttpMethodPost, HttpMethodPut, HttpMethodDelete, HttpMethodPatch}

type Router struct {
  
  // Hierarchy:
  parent *Router      // nil if root router.
  children []*Router  
  
  // For each request we'll create one of these objects
  contextType reflect.Type
  
  // Eg, "/" or "/admin". Any routes added to this router will be prefixed with this.
  pathPrefix string
  
  // Routeset contents:
  middleware []reflect.Value
  routes []*Route
  
  // The root pathnode is the same for a tree of Routers
  root map[HttpMethod]*PathNode
  
  // This can can be set on any router. The target's ErrorHandler will be invoked if it exists
  ErrorHandler func(*ResponseWriter, *Request, interface{})
  
  // This can only be set on the root handler, since by virtue of not finding a route, we don't have a target.
  // (That being said, in the future we could investigate namespace matches)
  NotFoundHandler func(*ResponseWriter, *Request)
}

type Route struct {
  Router *Router
  Method HttpMethod
  Path string
  Handler reflect.Value // Dynamic method sig.
}

type NextMiddlewareFunc func()

func New(ctx interface{}) *Router {
  validateContext(ctx, nil)
  
  r := &Router{}
  r.contextType = reflect.TypeOf(ctx)
  r.pathPrefix = "/"
  r.root = make(map[HttpMethod]*PathNode)
  for _, method := range HttpMethods {
    r.root[method] = newPathNode()
  }
  return r
}

func (r *Router) NewSubrouter(ctx interface{}) *Router {
  
  // First, we need to make sure that ctx includes a pointer to the parent context in the first slot
  validateContext(ctx, r.contextType)
  
  // Create new router, link up hierarchy
  newRouter := &Router{parent: r}
  r.children = append(r.children, newRouter)
  
  newRouter.contextType = reflect.TypeOf(ctx)
  newRouter.pathPrefix = r.pathPrefix
  newRouter.root = r.root
  
  fmt.Println("newRouter: ", newRouter) // Keep this to allow fmt
  
  return newRouter
}

func (r *Router) AddMiddleware(fn interface{}) *Router {
  validateMiddleware(fn, r.contextType)
  r.middleware = append(r.middleware, reflect.ValueOf(fn))
  return r
}

func (r *Router) SetNamespace(ns string) *Router {
  // TODO: do we need to re-eval all the routes ?
  // TODO: validate pathPrefix
  r.pathPrefix = ns
  return r
}

func (r *Router) Get(path string, fn interface{}) {
  r.addRoute(HttpMethodGet, path, fn)
}

func (r *Router) Post(path string, fn interface{}) {
  r.addRoute(HttpMethodPost, path, fn)
}

func (r *Router) Put(path string, fn interface{}) {
  r.addRoute(HttpMethodPut, path, fn)
}

func (r *Router) Delete(path string, fn interface{}) {
  r.addRoute(HttpMethodDelete, path, fn)
}

func (r *Router) Patch(path string, fn interface{}) {
  r.addRoute(HttpMethodPatch, path, fn)
}

// 
// 
// 
func (r *Router) addRoute(method HttpMethod, path string, fn interface{}) {
  
  // First, let's validate that fn is the proper type
  validateHandler(fn, r.contextType)
  
  fullPath := appendPath(r.pathPrefix, path)
  
  route := &Route{Method: method, Path: fullPath, Handler: reflect.ValueOf(fn), Router: r}
  r.routes = append(r.routes, route)
  
  r.root[method].add(fullPath, route)
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
  
  if parentCtxType != nil {
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
func validateHandler(fn interface{}, ctxType reflect.Type) {
  message := "web: handler must be a function with signature TODO"
  
  fnType := reflect.TypeOf(fn)
  
  if fnType.Kind() != reflect.Func {
    panic(message)
  }
  
  numIn := fnType.NumIn()
  numOut := fnType.NumOut()
  
  if numIn != 3 || numOut != 0 {
    panic(message)
  }
  
  if fnType.In(0) != reflect.PtrTo(ctxType) {
    panic(message)
  }
  
  var req *Request
  var resp *ResponseWriter
  if fnType.In(1) != reflect.TypeOf(resp) || fnType.In(2) != reflect.TypeOf(req) {
    panic(message)
  }
}

// Either of:
//    f(*context, *web.ResponseWriter, *web.Request, NextMiddlewareFunc)
//    f(*web.ResponseWriter, *web.Request, NextMiddlewareFunc)
func validateMiddleware(fn interface{}, ctxType reflect.Type) {
  message := "web: middlware must be a function with signature TODO"
  
  fnType := reflect.TypeOf(fn)
  
  if fnType.Kind() != reflect.Func {
    panic(message)
  }
  
  numIn := fnType.NumIn()
  numOut := fnType.NumOut()
  
  
  var i0, i1, i2 int
  if numOut != 0 {
    panic(message)
  }
  
  if numIn == 3 {
    i0, i1, i2 = 0, 1, 2
  } else if numIn == 4 {
    if fnType.In(0) != reflect.PtrTo(ctxType) {
      panic(message)
    }
    i0, i1, i2 = 1, 2, 3
  } else {
    panic(message)
  }
  
  var req *Request
  var resp *ResponseWriter
  var n NextMiddlewareFunc
  if fnType.In(i0) != reflect.TypeOf(resp) || fnType.In(i1) != reflect.TypeOf(req) || fnType.In(i2) != reflect.TypeOf(n) {
    panic(message)
  }
}

// Both rootPath/childPath are like "/" and "/users"
// Assumption is that both are well-formed paths.
func appendPath(rootPath, childPath string) string {
  if rootPath == "/" {
    return childPath
  }
  
  return rootPath + childPath
}


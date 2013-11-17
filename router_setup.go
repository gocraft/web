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
  errorHandler reflect.Value
  
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

func (r *Router) Subrouter(ctx interface{}, pathPrefix string) *Router {
  
  // First, we need to make sure that ctx includes a pointer to the parent context in the first slot
  validateContext(ctx, r.contextType)
  
  // Create new router, link up hierarchy
  newRouter := &Router{parent: r}
  r.children = append(r.children, newRouter)
  
  newRouter.contextType = reflect.TypeOf(ctx)
  newRouter.pathPrefix = appendPath(r.pathPrefix, pathPrefix)
  newRouter.root = r.root
  
  fmt.Println("newRouter: ", newRouter) // Keep this to allow fmt
  
  return newRouter
}

func (r *Router) Middleware(fn interface{}) *Router {
  fnv := reflect.ValueOf(fn)
  validateMiddleware(fnv, r.contextType)
  r.middleware = append(r.middleware, fnv)
  return r
}

func (r *Router) ErrorHandler(fn interface{}) {
  vfn := reflect.ValueOf(fn)
  validateErrorHandler(vfn, r.contextType)
  r.errorHandler = vfn
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
  fnv := reflect.ValueOf(fn)
  validateHandler(fnv, r.contextType)
  
  fullPath := appendPath(r.pathPrefix, path)
  
  route := &Route{Method: method, Path: fullPath, Handler: fnv, Router: r}
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
func validateHandler(fnv reflect.Value, ctxType reflect.Type) {
  var req *Request
  var resp *ResponseWriter
  if !isValidateHandler(fnv, ctxType, reflect.TypeOf(resp), reflect.TypeOf(req)) {
    panic("web: handler be a function with signature TODO")
  }
}

func validateErrorHandler(fnv reflect.Value, ctxType reflect.Type) {
  var req *Request
  var resp *ResponseWriter
  var wat func() interface{}  // This is weird. I need to get an interface{} reflect.Type; var x interface{}; TypeOf(x) doesn't work, because it returns nil
  if !isValidateHandler(fnv, ctxType, reflect.TypeOf(resp), reflect.TypeOf(req), reflect.TypeOf(wat).Out(0)) {
    panic("web: error handler be a function with signature TODO")
  }
}

// Either of:
//    f(*context, *web.ResponseWriter, *web.Request, NextMiddlewareFunc)
//    f(*web.ResponseWriter, *web.Request, NextMiddlewareFunc)
func validateMiddleware(fnv reflect.Value, ctxType reflect.Type) {
  var req *Request
  var resp *ResponseWriter
  var n NextMiddlewareFunc
  if !isValidateHandler(fnv, ctxType, reflect.TypeOf(resp), reflect.TypeOf(req), reflect.TypeOf(n)) {
    panic("web: middlware must be a function with signature TODO")
  }
}

// Ensures fnv is a function, that optionally takes a ctxType as the first argument, followed by the specified types. Handles have no return value.
// Returns true if valid, false otherwise.
func isValidateHandler(fnv reflect.Value, ctxType reflect.Type, types ...reflect.Type) bool {
  fnType := fnv.Type()
  
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

// Both rootPath/childPath are like "/" and "/users"
// Assumption is that both are well-formed paths.
func appendPath(rootPath, childPath string) string {
  if rootPath == "/" {
    return childPath
  }
  
  return rootPath + childPath
}


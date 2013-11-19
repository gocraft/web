package web

import (
  "reflect"
  "net/http"
  "fmt"
  "runtime"
)

func (rootRouter *Router) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
  
  // Wrap the request and writer.
  responseWriter := &ResponseWriter{rw}
  request := &Request{Request: r}
  
  // Handle errors
  defer func() {
    if recovered := recover(); recovered != nil {
      rootRouter.handlePanic(responseWriter, request, recovered) // TODO: that's wrong (used to be route.Router)
    }
  }()
  
  middlewareStack := rootRouter.MiddlewareStack(responseWriter, request)
  middlewareStack()
}

// r should be the root router
func (r *Router) MiddlewareStack(rw *ResponseWriter, req *Request) NextMiddlewareFunc {
  // Where are we in the stack
  routers := []*Router{r}
  contexts := []reflect.Value{reflect.New(r.contextType)}
  currentMiddlewareIndex := 0
  currentRouterIndex := 0
  currentMiddlewareLen := len(r.middleware)
  
  // Pre-make some Values
  vrw := reflect.ValueOf(rw)
  vreq := reflect.ValueOf(req)
  
  var next NextMiddlewareFunc // create self-referential anonymous function
  var nextValue reflect.Value
  next = func() {
    if currentRouterIndex >= len(routers) {
      return
    }
    
    // Find middleware to invoke. The goal of this block is to set the middleware variable. If it can't be done, it will be the zero value.
    // Side effects of this loop: set currentMiddlewareIndex, currentRouterIndex, currentMiddlewareLen
    var middleware reflect.Value
    if currentMiddlewareIndex < currentMiddlewareLen {
      middleware = routers[currentRouterIndex].middleware[currentMiddlewareIndex]
    } else {
      // We ran out of middleware on the current router
      if currentRouterIndex == 0 {
        // If we're still on the root router, it's time to actually figure out what the route is.
        // Do so, and update the various variables.
        // We could also 404 at this point: if so, run NotFound handlers and return.
        route, wildcardMap := calculateRoute(r, req)
        if route == nil {
          if r.notFoundHandler != nil {
            rw.WriteHeader(http.StatusNotFound)
            r.notFoundHandler(rw, req)
          } else {
            rw.WriteHeader(http.StatusNotFound)
            fmt.Fprintf(rw, DefaultNotFoundResponse)
          }
          return
        }
        
        req.route = route
        req.UrlVariables = wildcardMap
        
        routers = routersFor(route)
        contexts = contextsFor(contexts, routers)
      }
      
      currentMiddlewareIndex = 0
      currentRouterIndex += 1
      routersLen := len(routers)
      for currentRouterIndex < routersLen {
        currentMiddlewareLen = len(routers[currentRouterIndex].middleware)
        if currentMiddlewareLen > 0 {
          break
        }
        currentRouterIndex += 1
      }
      if currentRouterIndex < routersLen {
        middleware = routers[currentRouterIndex].middleware[currentMiddlewareIndex]
      } else {
        // We're done! invoke the action
        invoke(req.route.Handler, contexts[len(contexts) - 1], []reflect.Value{vrw, vreq})
      }
    }
    
    currentMiddlewareIndex += 1
    
    // Invoke middleware. Reflect on the function to call the context or no-context variant.
    if middleware.IsValid() {
      invoke(middleware, contexts[currentRouterIndex], []reflect.Value{vrw, vreq, nextValue})
    }
  }
  nextValue = reflect.ValueOf(next)
  
  return next
}

func calculateRoute(rootRouter *Router, req *Request) (*Route, map[string]string) {
  var leaf *PathLeaf
  var wildcardMap map[string]string
  tree, ok := rootRouter.root[HttpMethod(req.Method)]
  if ok {
    leaf, wildcardMap = tree.Match(req.URL.Path)
  }
  if leaf == nil {
    return nil, nil
  } else {
    return leaf.route, wildcardMap
  }
}

// given the route (and target router), return [root router, child router, ..., leaf route's router]
func routersFor(route *Route) []*Router {
  var routers []*Router
  curRouter := route.Router
  for curRouter != nil {
    routers = append(routers, curRouter)
    curRouter = curRouter.parent
  }
  
  // Reverse the slice
  s := 0
  e := len(routers) - 1
  for s < e {
    routers[s], routers[e] = routers[e], routers[s]
    s += 1
    e -= 1
  }
  
  return routers
}

// contexts is initially filled with a single context for the root
// routers is [root, child, ..., leaf] with at least 1 element
// Returns [ctx for root, ... ctx for leaf]
func contextsFor(contexts []reflect.Value, routers []*Router) []reflect.Value {
  routersLen := len(routers)
  
  for i := 1; i < routersLen; i += 1 {
    ctx := reflect.New(routers[i].contextType)
    contexts = append(contexts, ctx)
    
    // set the first field to the parent
    f := reflect.Indirect(ctx).Field(0)
    f.Set(contexts[i - 1])
  }
  
  return contexts
}

// This is called against the *target* router
func (targetRouter *Router) handlePanic(rw *ResponseWriter, req *Request, err interface{}) {
  
  // Find the first router that has an errorHandler
  // We also need to get the context corresponding to that router.
  curRouter := targetRouter
  curContextPtr := req.context
  for !curRouter.errorHandler.IsValid() && curRouter.parent != nil {
    curRouter = curRouter.parent
    
    // Need to set curContext to the next context, UNLESS the context is the same type.
    curContextStruct := reflect.Indirect(curContextPtr)
    if curRouter.contextType != curContextStruct.Type() {
      curContextPtr = curContextStruct.Field(0)
      if reflect.Indirect(curContextPtr).Type() != curRouter.contextType {
        panic("oshit why")
      }
    }
  }
  
  if curRouter.errorHandler.IsValid() {
    rw.WriteHeader(http.StatusInternalServerError)
    invoke(curRouter.errorHandler, curContextPtr, []reflect.Value{reflect.ValueOf(rw), reflect.ValueOf(req), reflect.ValueOf(err)})
  } else {
    http.Error(rw, DefaultPanicResponse, http.StatusInternalServerError)
  }
  
  const size = 4096
  stack := make([]byte, size)
  stack = stack[:runtime.Stack(stack, false)]
  
  ERROR.Printf("%v\n", err)
  ERROR.Printf("%s\n", string(stack))
}

func invoke(handler reflect.Value, ctx reflect.Value, values []reflect.Value) {
  handlerType := handler.Type()
  numIn := handlerType.NumIn()
  if numIn == len(values) {
    handler.Call(values)
  } else {
    values = append([]reflect.Value{ctx}, values...)
    handler.Call(values)
  }
}

var DefaultNotFoundResponse string = "Not Found"
var DefaultPanicResponse string = "Application Error"



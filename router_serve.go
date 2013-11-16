package web

import (
  "reflect"
  "net/http"
  "fmt"
  "runtime"
)

// This is the main entry point for a request from the built-in Go http library.
// router should be the root router.
func (rootRouter *Router) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
  
  // Wrap the request and writer.
  responseWriter := &ResponseWriter{rw}
  request := &Request{Request: r}
  
  // Do routing
  // TODO: fix bad method
  leaf, wildcardMap := rootRouter.root[HttpMethod(r.Method)].Match(r.URL.Path)
  if leaf == nil {
    fmt.Println("Not Found: ", r.URL.Path)
    return
  }
  
  route := leaf.route
  request.route = route
  request.UrlVariables = wildcardMap
  
  // Handle errors
  defer func() {
    if recovered := recover(); recovered != nil {
      route.Router.handlePanic(responseWriter, request, recovered)
    }
  }()
  
  middlewareStack := route.Router.MiddlewareStack(responseWriter, request)
  middlewareStack()
}

func (router *Router) handlePanic(rw *ResponseWriter, req *Request, err interface{}) {
  
  // Find the first router that has an errorHandler
  curRouter := router
  for !curRouter.errorHandler.IsValid() && curRouter.parent != nil {
    curRouter = curRouter.parent
  }
  
  if curRouter.errorHandler.IsValid() {
    rw.WriteHeader(http.StatusInternalServerError)
    invoke(curRouter.errorHandler, req.currentContext, []reflect.Value{reflect.ValueOf(rw), reflect.ValueOf(req), reflect.ValueOf(err)})
  } else {
    http.Error(rw, "Something went wrong", http.StatusInternalServerError)
  }
  
  const size = 4096
  stack := make([]byte, size)
  stack = stack[:runtime.Stack(stack, false)]
  
  ERROR.Printf("%v\n", err)
  ERROR.Printf("%s\n", string(stack))
}

// This is the last middleware. It will just invoke the action
func RouteInvokingMiddleware(rw *ResponseWriter, req *Request, next NextMiddlewareFunc) {
  req.route.Handler.Call([]reflect.Value{req.context, reflect.ValueOf(rw), reflect.ValueOf(req)})
}

// Routers is a [leaf, child, ... , root]. Return [ctx for leaf, ctx for child, ..., ctx for root]
// Routers must have at least one element
func createContexts(routers []*Router) (contexts []reflect.Value) {
  routersLen := len(routers)
  
  contexts = make([]reflect.Value, routersLen)
  
  for i := routersLen - 1; i >= 0; i -= 1 {
    ctx := reflect.New(routers[i].contextType)
    contexts[i] = ctx
    
    // If we're not the root context, then set the first field to the parent
    if i < routersLen - 1 {
      f := reflect.Indirect(ctx).Field(0)
      f.Set(contexts[i + 1])
    }
  }
  
  return
}

//
func (r *Router) MiddlewareStack(rw *ResponseWriter, req *Request) NextMiddlewareFunc {
  // r is the target router (could be a leaf router, or the root router, or somewhere in between)
  // Construct routers, being [leaf, child, ..., root]
  var routers []*Router
  curRouter := r
  for curRouter != nil {
    routers = append(routers, curRouter)
    curRouter = curRouter.parent
  }
  
  // contexts are parallel to routers. We're going to pre-emptively create all contexts
  var contexts []reflect.Value
  contexts = createContexts(routers)
  req.context = contexts[0]
  
  // Inputs into next():
  // routers: 1 or more routers in reverse order
  // currentRouterIndex: N-1, ..., 0. If -1, then we're done
  // currentMiddlwareLen: len(routers[currentRouterIndex].middleware)
  // currentMiddlewareIndex: 0, ..., len(routers[currentRounterIndex]). We *CAN* enter next() with this out of bounds. That's expected.
  currentRouterIndex := len(routers) - 1
  currentMiddlewareLen := len(routers[currentRouterIndex].middleware)
  currentMiddlewareIndex := 0
  
  var next NextMiddlewareFunc // create self-referential anonymous function
  var nextValue reflect.Value
  
  // Pre-make some Values
  vrw := reflect.ValueOf(rw)
  vreq := reflect.ValueOf(req)
  
  next = func() {
    if currentRouterIndex < 0 {
      return
    }
    
    // Find middleware to invoke. The goal of the if statement is to set the middleware variable. If it can't be done, it will be the zero value.
    // Side effects of this loop: set currentMiddlewareIndex, currentRouterIndex
    var middleware reflect.Value
    if currentMiddlewareIndex < currentMiddlewareLen {
      // It's in bounds? Cool, use it
      middleware = routers[currentRouterIndex].middleware[currentMiddlewareIndex]
    } else {
      // Out of bounds. Find next router with middleware. If none, use the invoking middleware
      currentMiddlewareIndex = 0
      for {
        currentRouterIndex -= 1
        
        if currentRouterIndex < 0 {
          // If we're at the end of the routers, invoke the final virtual middleware: the handler invoker.
          // (next() wont execute on future calls b/c we'll return at the top)
          middleware = reflect.ValueOf(RouteInvokingMiddleware)
          break
        }
        
        // So currentRouterIndex >= 0 b/c we didn't break.
        currentMiddlewareLen = len(routers[currentRouterIndex].middleware)
        
        if currentMiddlewareLen > 0 {
          middleware = routers[currentRouterIndex].middleware[currentMiddlewareIndex]
          break
        }
        // didn't break? loop
      }
    }
    
    // Make sure we increment the index for the next time
    currentMiddlewareIndex += 1
    
    // Invoke middleware. Reflect on the function to call the context or no-context variant.
    if middleware.IsValid() {
      var ctx reflect.Value
      if currentRouterIndex >= 0 {
        ctx := contexts[currentRouterIndex]
        req.currentContext = ctx
      }
      invoke(middleware, ctx, []reflect.Value{vrw, vreq, nextValue})
    }
  }
  nextValue = reflect.ValueOf(next)
  
  return next
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

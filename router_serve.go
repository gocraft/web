package web

import (
	"fmt"
	"net/http"
	"reflect"
	"runtime"
)

// This is the entry point for servering all requests
func (rootRouter *Router) ServeHTTP(rw http.ResponseWriter, r *http.Request) {

	// Wrap the request and writer.
	responseWriter := &AppResponseWriter{ResponseWriter: rw}
	request := &Request{Request: r}

	// Handle errors
	defer func() {
		if recovered := recover(); recovered != nil {
			rootRouter.handlePanic(responseWriter, request, recovered)
		}
	}()

	next := rootRouter.MiddlewareStack(request)
	next(responseWriter, request)
}

// r should be the root router
// This function executes the middleware stack. It does so creating/returning an anonymous function/closure.
// This closure can be called multiple times (eg, next()). Each time it is called, the next middleware is called.
// Each time a middleware is called, this 'next' function is passed into it, which will/might call it again.
// There are two 'virtual' middlewares in this stack: the route choosing middleware, and the action invoking middleware.
// The route choosing middleware is executed after all root middleware. It picks the route.
// The action invoking middleware is executed after all middleware. It executes the final handler.
func (r *Router) MiddlewareStack(request *Request) NextMiddlewareFunc {
	// Where are we in the stack
	routers := make([]*Router, 1, r.maxChildrenDepth)
	routers[0] = r
	contexts := make([]reflect.Value, 1, r.maxChildrenDepth)
	contexts[0] = reflect.New(r.contextType)
	currentMiddlewareIndex := 0
	currentRouterIndex := 0
	currentMiddlewareLen := len(r.middleware)

	request.rootContext = contexts[0]

	var next NextMiddlewareFunc // create self-referential anonymous function
	var nextValue reflect.Value
	next = func(rw ResponseWriter, req *Request) {
		if currentRouterIndex >= len(routers) {
			return
		}

		// Find middleware to invoke. The goal of this block is to set the middleware variable. If it can't be done, it will be the zero value.
		// Side effects of this block:
		//  - set currentMiddlewareIndex, currentRouterIndex, currentMiddlewareLen
		//  - calculate route, setting routers/contexts, and fields in req.
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
					if r.notFoundHandler.IsValid() {
						invoke(r.notFoundHandler, contexts[0], []reflect.Value{reflect.ValueOf(rw), reflect.ValueOf(req)})
					} else {
						rw.WriteHeader(http.StatusNotFound)
						fmt.Fprintf(rw, DefaultNotFoundResponse)
					}
					return
				}

				routers = routersFor(route, routers)
				contexts = contextsFor(contexts, routers)

				req.targetContext = contexts[len(contexts)-1]
				req.route = route
				req.PathParams = wildcardMap
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
				invoke(req.route.Handler, contexts[len(contexts)-1], []reflect.Value{reflect.ValueOf(rw), reflect.ValueOf(req)})
			}
		}

		currentMiddlewareIndex += 1

		// Invoke middleware.
		if middleware.IsValid() {
			invoke(middleware, contexts[currentRouterIndex], []reflect.Value{reflect.ValueOf(rw), reflect.ValueOf(req), nextValue})
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
// Use the memory in routers to store this information
func routersFor(route *Route, routers []*Router) []*Router {
	routers = routers[:0]
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
// NOTE: if two routers have the same contextType, then they'll share the exact same context.
func contextsFor(contexts []reflect.Value, routers []*Router) []reflect.Value {
	routersLen := len(routers)

	for i := 1; i < routersLen; i += 1 {
		var ctx reflect.Value
		if routers[i].contextType == routers[i-1].contextType {
			ctx = contexts[i-1]
		} else {
			ctx = reflect.New(routers[i].contextType)
			// set the first field to the parent
			f := reflect.Indirect(ctx).Field(0)
			f.Set(contexts[i-1])
		}
		contexts = append(contexts, ctx)
	}

	return contexts
}

// If there's a panic in the root middleware (so that we don't have a route/target), then invoke the root handler or default.
// If there's a panic in other middleware, then invoke the target action's function.
// If there's a panic in the action handler, then invoke the target action's function.
func (rootRouter *Router) handlePanic(rw *AppResponseWriter, req *Request, err interface{}) {
	var targetRouter *Router  // This will be set to the router we want to use the errorHandler on.
	var context reflect.Value // this is the context of the target router

	if req.route == nil {
		targetRouter = rootRouter
		context = req.rootContext
	} else {
		targetRouter = req.route.Router
		context = req.targetContext

		for !targetRouter.errorHandler.IsValid() && targetRouter.parent != nil {
			targetRouter = targetRouter.parent

			// Need to set context to the next context, UNLESS the context is the same type.
			curContextStruct := reflect.Indirect(context)
			if targetRouter.contextType != curContextStruct.Type() {
				context = curContextStruct.Field(0)
				if reflect.Indirect(context).Type() != targetRouter.contextType {
					panic("oshit why")
				}
			}
		}
	}

	if targetRouter.errorHandler.IsValid() {
		invoke(targetRouter.errorHandler, context, []reflect.Value{reflect.ValueOf(rw), reflect.ValueOf(req), reflect.ValueOf(err)})
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

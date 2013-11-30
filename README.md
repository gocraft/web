# Mars Web

Mars Web is a Go mux + middleware tool. We deal with casting and reflection so YOUR code can be statically typed. We play nicely with Go's built-in HTTP tools.

## Getting Started


## Features
* **Super Fast and Scalable**. Added latency is from 3-9μs per request. Routing performance is O(log(N)) in the number of routes.
* **Your own contexts**. Easily pass information between your middleware and handler with strong static typing.
* **Easy and Powerful routing**. Capture path variables. Validate path segments with regexps. Lovely API.
* **Middleware**. Middleware can express almost any web-layer feature. We make it easy.
* **Nested routers, contexts, and middleware**. Your app has an API, and admin area, and a logged out view. Each view needs different contexts and different middleware. We let you express this hierarchy naturally.
* **Embrace Go's net/http package**. Start your server with http.ListenAndServe(), and work directly with http.ResponseWriter and http.Request.
* **Minimal**. The core of Mars Web is lightweight and minimal. Add optional functionality with our built-in middleware, or write your own middleware.

## Performance
Performance is a first class concern. Every update to this package has its performance measured and tracked in [BENCHMARK_RESULTS](https://github.com/cypriss/mars_web/blob/master/BENCHMARK_RESULTS).

For minimal 'hello world' style apps, added latency is about 3μs. This grows to about 10μs for more complex apps (6 middleware functions, 3 levels of contexts, 150+ routes).


One key design choice we've made is our choice of routing algorithm. Most competing libraries use simple O(N) iteration over all routes to find a match. This is fine if you have only a handful of routes, but starts to break down as your app gets bigger. Mars Web uses a tree-based router which grows in complexity at O(log(N)).

## Application Structure

### Making your router
The first thing you need to do is make a new router. Routers serve requests and execute middleware.

```go
router := web.New(YourContext{})
```

### Your context
Wait, what is YourContext{} and why do you need it? It can be any struct you want it to be. Here's an example of one:

```go
type YourContext struct {
  User *User // Assumes you've defined a User type as well
}
```

Your context can be empty or it can have various fields in it. The fields can be whatever you want - it's your type! When a new request comes into the router, we'll allocate an instance of this struct and pass it to your middleware and handlers. This allows, for instance, a SetUser middleware to set a User field that can be read in the handlers.

### Routes and handlers
Once you have your router, you can add routes to it. Standard HTTP verbs are supported.

```go
router := web.New(YourContext{})
router.Get("/users", (*YourContext).UsersList)
router.Post("/users", (*YourContext).UsersCreate)
router.Put("/users/:id", (*YourContext).UsersUpdate)
router.Delete("/users/:id", (*YourContext).UsersDelete)
router.Patch("/users/:id", (*YourContext).UsersUpdate)
router.Get("/", (*YourContext).Root)
```

What is that funny ```(*YourContext).Root``` notation? It's called a method expression. It lets your handlers look like this:

```go
func (c *YourContext) Root(rw web.ResponseWriter, req *web.Request) {
  if c.User != nil {
    fmt.Fprint(rw, "Hello,", c.User.Name)
  } else {
    fmt.Fprint(rw, "Hello, anonymous person")
  }
}
```

All method expressions do is return a function that accepts the type as the first argument. So your handler can also look like this:

```go
func Root(c *YourContext, rw web.ResponseWriter, req *web.Request) {}
```

Of course, if you don't need a context for a particluar action, you can also do that:

```go
func Root(rw web.ResponseWriter, req *web.Request) {}
```

Note that handlers always need to accept two input parameters: web.ResponseWriter, and *web.Request, both of which wrap the standard http.ResponseWriter and *http.Request, respectively.

### Middleware
You can add middleware to a router:

```go
router := web.New(YourContext{})
router.Middleware((*YourContext).UserRequired)
// add routes, more middleware
```

This is what a middleware handler looks like:

```go
func (c *YourContext) UserRequired(rw web.ResponseWriter, r *web.Request, next web.NextMiddlewareFunc) {
  user := userFromSession(r)  // Pretend like this is defined. It reads a session cookie and returns a *User or nil.
  if user != nil {
    c.User = user
    next(rw, r)
  } else {
		rw.Header().Set("Location", "/")
		rw.WriteHeader(http.StatusMovedPermanently)
    // do NOT call next()
  }
}
```

Some things to note about the above example:
*  We set fields in the context for future middleware / handlers to use.
*  We can call next(), or not. Not calling next() effectively stops the middleware stack.

Of course, generic middleware without contexts are supported:

```go
func GenericMiddleware(rw web.ResponseWriter, r *web.Request, next web.NextMiddlewareFunc) {
  // ...
}
```

### Nested routers
Nested routers let you run different middleware and use different contexts for different parts of your app. Some common scenarios:
*  You want to run an AdminRequired middleware on all your admin routes, but not on API routes. Your context needs a CurrentAdmin field.
*  You want to run an OAuth middleware on your API routes. Your context needs an AccessToken field.
*  You want to run session handling middleware on ALL your routes. Your context needs a Session field.

Let's implement that. Your contexts would look like this:

```go
type Context struct {
  Session map[string]string
}

type AdminContext struct {
  *Context
  CurrentAdmin *User
}

type ApiContext struct {
  *Context
  AccessToken string
}
```

Note that we embed a pointer to the root context in each subcontext. This is required.

Now that we have our contexts, let's create our routers:

```go
rootRouter := web.New(Context{})
rootRouter.Middleware((*Context).LoadSession)

apiRouter := rootRouter.Subrouter(ApiContext{}, "/api")
apiRouter.Middleware((*ApiContext).OAuth)
apiRouter.Get("/tickets", (*ApiContext).TicketsIndex)

adminRouter := rootRouter.Subrouter(AdminContext{}, "/admin")
adminRouter.Middleware((*AdminContext).AdminRequired)
adminRouter.Get("/reports", (*AdminContext).Reports)  // Given the path namesapce for this router is "/admin", the full path of this route is "/admin/reports"
```

Note that each time we make a subrouter, we need to supply the context as well as a path namespace. The context CAN be the same as the parent context, and the namespace CAN just be "/" for no namespace.

### Request lifecycle
1.  Wrap the default Go http.ResponseWriter and http.Request in a web.ResponseWriter and web.Request, respectively (via structure embedding).
2.  Allocate a new root context. This context is passed into your root middleware.
3.  Execute middleware on the root router. We do this before we find a route!
4.  After all of the root router's middleware is executed, we'll run a 'virtual' routing middleware that determines the target route.
    *  If the there's no route found, we'll execute the NotFound handler if supplied. Otherwise, we'll write a 404 response and start unwinding the root middlware.
5.  Now that we have a target route, we can allocate the context tree of the target router.
6.  Start executing middleware on the nested middleware leading up to the final router/route.
7.  After all middleware is executed, we'll run another 'virtual' middleware that invokes the final handler corresponding to the target route.
8.  Unwind all middleware calls (if there's any code after next() in the middleware, obviously that's going to run at some point).

### Capturing path variables; regexp conditions
### 404 handlers
### 500 handlers
### Included middleware
We ship with three basic pieces of middleware: a logger, an exception printer, and a static file server. To use them:

```go
router := web.New(Context{})
router.Middleware(web.LoggerMiddleware).
       Middleware(web.ShowErrorsMiddleware).
       Middleware(web.StaticMiddleware("public")) // "public" is a directory to serve files from.
```

NOTE: You might not want to use web.ShowErrorsMiddleware in production. You can easily do something like this:
```go
router := web.New(Context{})
router.Middleware(web.LoggerMiddleware)
if MyEnvironment == "development" {
  router.Middleware(web.ShowErrorsMiddleware)
}
// ...
```

### Starting your server
Since web.Router implements http.Handler (eg, ServeHTTP(ResponseWriter, *Request)), you can easily plug it in to the standard Go http machinery:

```go
router := web.New(Context{})
// ... Add routes and such.
http.ListenAndServe("localhost:8080", router)
```

### Rendering responses

### Other Thoughts
* context can be the same between nestings


## FAQ

## Thanks

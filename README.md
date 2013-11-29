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
The first thing you need to do is make a new router. Routers serve ruquests and execute middleware.

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

What is that funny ```(*YourContext).Root``` notation?

```go
func (c *YourContext) Root(rw web.RequestWriter, req *web.Request) {
  if c.User != nil {
    fmt.Fprint(rw, "Hello,", c.User.Name)
  } else {
    fmt.Fprint(rw, "Hello, anonymous person")
  }
}
```

```go
func Root(c *YourContext, rw web.RequestWriter, req *web.Request) {}
```

```go
func Signin(rw web.RequestWriter, req *web.Request) {}
```


### Middleware
```go
router := web.New(YourContext{})
router.Middleware((*YourContext).UserRequired)
// add routes, more middleware
```

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

```go
func GenericMiddleware(rw web.ResponseWriter, r *web.Request, next web.NextMiddlewareFunc) {
  // ...
}

### Nested routers
### Request lifecycle
1.  Wrap the default Go http.ResponseWriter and http.Request in a web.ResponseWriter and web.Request, respectively (via structure embedding).
2.  Allocate a new root context. This context is passed into your root middleware.
3.  Execute middleware on the root router. We do this before we find a route!
4.  After all of the root router's middleware is executed, we'll run a 'virtual' routing middleware that determines the target route.
    *  If the there's no route found, we'll execute the NotFound handler if supplied. Otherwise, we'll write a 404 response and start unwinding the root middlware.
5.  Now that we have a target route, we'll start executing middleware on the nested middleware leading up to the final target router/route.
6.  After all middleware is executed, we'll run another 'virtual' middleware that invokes the final handler corresponding to the target route.
7.  Unwind all middleware calls (if there's any code after next() in the middleware, obviously that's going to run at some point).

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

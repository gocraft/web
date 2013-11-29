# Mars Web

Mars Web is a Go mux + middleware tool. We deal with casting and reflection so YOUR code can be statically typed. We play nicely with Go's built-in HTTP tools.

## Getting Started


## Features
* **Super Fast and Scalable**. Added latency is from 3-9μs per request. Routing performance is O(log(N)) in the number of routes.
* **Your own contexts**. Easily pass information between your middleware and handler with strong static typing.
* **Easy and Powerful routing**. Capture path variables. Validate path segments with regexps. Lovely API.
* **Middleware**. Middleware can express almost any web-layer feature. We make it easy.
* **Nested routers, contexts, and middleware**. Your app has an API, and admin area, and a logged out view. Each view needs different contexts and different middleware.
* **Embrace Go's net/http package**. Start your server with http.ListenAndServe(), and work directly with http.ResponseWriter and http.Request.
* **Minimal**. The core of Mars Web is lightweight and minimal. Add optional functionality with our built-in middleware, or write your own middleware.

## Performance

## Application Structure

### Making your router
```go
router := web.New(YourContext{})
```
### Your context
```go
type YourContext struct {
  User *User // Assumes you've defined a User type
}
```
### Routes and handlers

```go
router := web.New(YourContext{})
router.Get("/signin", (*YourContext).Signin)
router.Post("/sessions", (*YourContext).CreateSession)
router.Get("/", (*YourContext).Root)
```

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

# Mars Web

Mars Web is a Go mux + middleware tool. We deal with casting and reflection so YOUR code can be statically typed. We play nicely with Go's built-in HTTP tools.

## Getting Started


## Features
* **Super Fast**. 
* Your own contexts
* Easy and Powerful routing
* Middleware
* Nested routers, contexts, and middleware.
* Embrace Go's net/http package.
* Minimal

## Performance

## Structure

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

### Nested routers
### Request lifecycle
1.  Execute middleware on the root router. We do this before we find a route!
2.  Determine the target route on the target router.
    *  If the there's no route found, we'll execute the NotFound handler if supplied. Otherwise, we'll write a 404 response and start unwinding the root middlware.
3.  Other stuff

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

## FAQ

## Thanks

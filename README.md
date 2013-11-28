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
### Your context
### Routes and handlers
### Middleware
### Nested routers
### Request lifecycle
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
## FAQ

## Thanks

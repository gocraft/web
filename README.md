# Mars Web

Mars Web is a Go mux + middleware tool. We deal with casting and reflection so YOUR code can be statically typed. We play nicely with Go's built-in HTTP tools.

## Example

```go
type MyContext struct {
  User *User
}

func (c *MyContext) UserRequired(rw *web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc) {
  if c.User, ok := Db.Find(req.Get("user_id")); ok {
    next()
  } else {
    rw.WriteHeader(http.StatusUnauthorized)
  }
}

func (c *MyContext) ShowProfile(rw *web.ResponseWriter, req *web.Request) {
  fmt.Fprintf(rw, "Hi, %s", c.User.Name)
}

router := web.New(MyContext{})
router.Middleware((*MyContext).UserRequired)
router.Get("/profile", (*MyContext).ShowProfile)
http.ListenAndServe("localhost:8080", router)
```

Notice how the middleware and handlers are methods on our own custom structs. This lets us set variables in "before filters" and then read them in handlers, without any type assertions.

## Installation

```bash
go get github.com/cypriss/mars_web
```

## Thanks

This is the result of reading many other excellent Go routing libraries, including: pilu/traffic, robfig/revel, gorilla/mux, and more.
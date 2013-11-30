package main

import (
  "fmt"
  "net/http"
  "github.com/gocraft/web"
  "log"
)

type Context struct {
  RequestIdentifier string
}
type AdminContext struct {
  *Context
}

func (ctx *Context) SetRequestIdentifier(w web.ResponseWriter, r *web.Request, next web.NextMiddlewareFunc) {
  //fmt.Println("Setting request identifier")
  ctx.RequestIdentifier = "123"
  next(w, r)
}

func (ctx *Context) Signin(w web.ResponseWriter, r *web.Request) {
  fmt.Fprintf(w, "You got to signin")
}

func (ctx *AdminContext) UsersList(w web.ResponseWriter, r *web.Request) {
  fmt.Fprintln(w, "UsersList: ", w, r)
  fmt.Fprintln(w, "UsersList: request identifier: ", ctx.RequestIdentifier)
}

func (ctx *AdminContext) Exception(w web.ResponseWriter, r *web.Request) {
  var x,y int
  fmt.Println(x/y)
}

func (ctx *AdminContext) SuggestionView(w web.ResponseWriter, r *web.Request) {
  fmt.Fprintln(w, "SuggestionView: entered")
  fmt.Fprintln(w, "r = ", r.PathParams)
}

func main() {
  router := web.New(Context{})
  router.Middleware(web.LoggerMiddleware).
         Middleware(web.ShowErrorsMiddleware).
         Middleware(web.StaticMiddleware("public")).
         Middleware((*Context).SetRequestIdentifier)
  
  router.Get("/signin", (*Context).Signin)
  
  adminRouter := router.Subrouter(AdminContext{}, "/admin")
  
  adminRouter.Get("/users", (*AdminContext).UsersList)
  adminRouter.Get("/exception", (*AdminContext).Exception)
  adminRouter.Get("/forums/:forum_id:\\d+/suggestions/:suggestion_id:[a-z]+", (*AdminContext).SuggestionView)
  
  err := http.ListenAndServe(":8080", router)
  if err != nil {
    log.Fatal(err)
  }
}
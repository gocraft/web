package main

import (
  "fmt"
  "net/http"
  "web"
  "log"
)

type Context struct {
  RequestIdentifier string
}
type AdminContext struct {
  *Context
}

func (ctx *Context) SanitizeUtf8(w *web.ResponseWriter, r *web.Request, next web.NextMiddlewareFunc) {
  fmt.Println("Sanitizing utf8...")
  next()
}

func (ctx *Context) SetRequestIdentifier(w *web.ResponseWriter, r *web.Request, next web.NextMiddlewareFunc) {
  fmt.Println("Setting request identifier")
  ctx.RequestIdentifier = "123"
  next()
}

func (ctx *Context) Signin(w *web.ResponseWriter, r *web.Request) {
  fmt.Println("signin", w)
  //fmt.Fprintf(w, "Hi signin.\n")
  var x int
  var y int
  fmt.Println(x/y)
}

func (ctx *AdminContext) UsersList(w *web.ResponseWriter, r *web.Request) {
  fmt.Println("UsersList: ", w, r)
  fmt.Println("UsersList: request identifier: ", ctx.RequestIdentifier)
}

func (ctx *AdminContext) SuggestionView(w *web.ResponseWriter, r *web.Request) {
  fmt.Println("SuggestionView: entered")
  fmt.Println("r = ", r.UrlVariables)
}


func main() {
  router := web.New(Context{})
  router.AddMiddleware(web.LoggerMiddleware).
         AddMiddleware(web.ShowErrorsMiddleware).
         AddMiddleware((*Context).SanitizeUtf8).
         AddMiddleware((*Context).SetRequestIdentifier)
  
  router.Get("/signin", (*Context).Signin)
  
  adminRouter := router.NewSubrouter(AdminContext{})
  adminRouter.SetNamespace("/admin")   
  
  //adminRouter.Get("/users", (*AdminContext).UsersList)
  adminRouter.Get("/forums/:forum_id/suggestions/:suggestion_id", (*AdminContext).SuggestionView)
  
  fmt.Println("ok: ", adminRouter)
  
  
  err := http.ListenAndServe(":8080", router)
  if err != nil {
    log.Fatal(err)
  }
  
  // req, _ := http.NewRequest("GET", "/admin/users", nil)
  // router.ServeHTTP(nil, req)
  // 
  // req, _ = http.NewRequest("GET", "/admin/forums/3/suggestions/88-hithere?foo=bar", nil)
  // router.ServeHTTP(nil, req)
}
package web

import (
  . "launchpad.net/gocheck"
  "testing"
  "fmt"
  "net/http"
  "net/http/httptest"
)

//
// gocheck: hook into "go test"
//
func Test(t *testing.T) { TestingT(t) }
type ErrorTestSuite struct {
}
var _ = Suite(&ErrorTestSuite{})

func newTestRequest(method, path string) (*httptest.ResponseRecorder, *http.Request) {
  request, _ := http.NewRequest(method, path, nil)
  recorder := httptest.NewRecorder()

  return recorder, request
}

type Context struct {
  A int
}

type AdminContext struct {
  *Context
  B int
}

type ApiContext struct {
  *Context
  C int
}

type TicketsContext struct {
  *AdminContext
}

func (c *Context) SetA(w *ResponseWriter, r *Request, next NextMiddlewareFunc) {
  c.A = 1
  next()
}

func (c *AdminContext) SetB(w *ResponseWriter, r *Request, next NextMiddlewareFunc) {
  c.B = 10
  next()
}


func (c *ApiContext) SetC(w *ResponseWriter, r *Request, next NextMiddlewareFunc) {
  c.C = 100
  next()
}

func (c *TicketsContext) Index(w *ResponseWriter, r *Request) {
  fmt.Fprintf(w, "poopin")
}

func (s *ErrorTestSuite) TestNoErorHandler(c *C) {
  fmt.Println("here")
  
  router := New(Context{})
  router.Middleware((*Context).SetA)
  
  apiRouter := router.Subrouter(ApiContext{}, "/api")
  apiRouter.Middleware((*ApiContext).SetC)
  
  adminRouter := router.Subrouter(AdminContext{}, "/admin")
  adminRouter.Middleware((*AdminContext).SetB)
  
  ticketsRouter := adminRouter.Subrouter(TicketsContext{}, "/tickets")
  ticketsRouter.Get("/index", (*TicketsContext).Index)
  
  fmt.Println(ticketsRouter)
  
  rw, req := newTestRequest("GET", "/admin/tickets/index")
  router.ServeHTTP(rw, req)
  
  fmt.Printf(string(rw.Body.Bytes()))
}

// Things I want to test:
// - no handler {r, l}
// - handler on root {r, l}
// - handler on root AND api, exception on {r, api}
// - handler on api, exception on admin







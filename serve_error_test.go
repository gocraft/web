package web

import (
  . "launchpad.net/gocheck"
  "testing"
  "fmt"
  "net/http"
  "net/http/httptest"
  "strings"
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

type Context struct {}

type AdminContext struct {
  *Context
}

type ApiContext struct {
  *Context
}

type SiteContext struct {
  *Context
}

type TicketsContext struct {
  *AdminContext
}

func (c *Context) ErrorHandler(w *ResponseWriter, r *Request, err interface{}) {
  fmt.Fprintf(w, "My Error")
}

func (c *Context) Action(w *ResponseWriter, r *Request) {
  var x, y int
  fmt.Fprintln(w, x / y)
}

func (c *AdminContext) ErrorHandler(w *ResponseWriter, r *Request, err interface{}) {
  fmt.Fprintf(w, "Admin Error")
}

func (c *AdminContext) Action(w *ResponseWriter, r *Request) {
  var x, y int
  fmt.Fprintln(w, x / y)
}

func (c *ApiContext) ErrorHandler(w *ResponseWriter, r *Request, err interface{}) {
  fmt.Fprintf(w, "Api Error")
}

func (c *ApiContext) Action(w *ResponseWriter, r *Request) {
  var x, y int
  fmt.Fprintln(w, x / y)
}

func (c *TicketsContext) Action(w *ResponseWriter, r *Request) {
  var x, y int
  fmt.Fprintln(w, x / y)
}

func ErrorHandlerWithNoContext(w *ResponseWriter, r *Request, err interface{}) {
  fmt.Fprintf(w, "Contextless Error")
}

func (s *ErrorTestSuite) TestNoErorHandler(c *C) {
  router := New(Context{})
  router.Get("/action", (*Context).Action)
  
  admin := router.Subrouter(AdminContext{}, "/admin")
  admin.Get("/action", (*AdminContext).Action)
  
  rw, req := newTestRequest("GET", "/action")
  router.ServeHTTP(rw, req)
  c.Assert(strings.TrimSpace(string(rw.Body.Bytes())), Equals, "Application Error")
  c.Assert(rw.Code, Equals, http.StatusInternalServerError)
  
  rw, req = newTestRequest("GET", "/admin/action")
  router.ServeHTTP(rw, req)
  c.Assert(strings.TrimSpace(string(rw.Body.Bytes())), Equals, "Application Error")
  c.Assert(rw.Code, Equals, http.StatusInternalServerError)
}

func (s *ErrorTestSuite) TestHandlerOnRoot(c *C) {
  router := New(Context{})
  router.ErrorHandler((*Context).ErrorHandler)
  router.Get("/action", (*Context).Action)
  
  admin := router.Subrouter(AdminContext{}, "/admin")
  admin.Get("/action", (*AdminContext).Action)
  
  rw, req := newTestRequest("GET", "/action")
  router.ServeHTTP(rw, req)
  c.Assert(strings.TrimSpace(string(rw.Body.Bytes())), Equals, "My Error")
  c.Assert(rw.Code, Equals, http.StatusInternalServerError)
  
  rw, req = newTestRequest("GET", "/admin/action")
  router.ServeHTTP(rw, req)
  c.Assert(strings.TrimSpace(string(rw.Body.Bytes())), Equals, "My Error")
  c.Assert(rw.Code, Equals, http.StatusInternalServerError)
}

func (s *ErrorTestSuite) TestContextlessError(c *C) {
  router := New(Context{})
  router.ErrorHandler(ErrorHandlerWithNoContext)
  router.Get("/action", (*Context).Action)
  
  admin := router.Subrouter(AdminContext{}, "/admin")
  admin.Get("/action", (*AdminContext).Action)
  
  rw, req := newTestRequest("GET", "/action")
  router.ServeHTTP(rw, req)
  c.Assert(strings.TrimSpace(string(rw.Body.Bytes())), Equals, "Contextless Error")
  c.Assert(rw.Code, Equals, http.StatusInternalServerError)
  
  rw, req = newTestRequest("GET", "/admin/action")
  router.ServeHTTP(rw, req)
  c.Assert(strings.TrimSpace(string(rw.Body.Bytes())), Equals, "Contextless Error")
  c.Assert(rw.Code, Equals, http.StatusInternalServerError)
}

func (s *ErrorTestSuite) TestMultipleErrorHandlers(c *C) {
  router := New(Context{})
  router.ErrorHandler((*Context).ErrorHandler)
  router.Get("/action", (*Context).Action)
  
  admin := router.Subrouter(AdminContext{}, "/admin")
  admin.ErrorHandler((*AdminContext).ErrorHandler)
  admin.Get("/action", (*AdminContext).Action)
  
  rw, req := newTestRequest("GET", "/action")
  router.ServeHTTP(rw, req)
  c.Assert(strings.TrimSpace(string(rw.Body.Bytes())), Equals, "My Error")
  c.Assert(rw.Code, Equals, http.StatusInternalServerError)
  
  rw, req = newTestRequest("GET", "/admin/action")
  router.ServeHTTP(rw, req)
  c.Assert(strings.TrimSpace(string(rw.Body.Bytes())), Equals, "Admin Error")
  c.Assert(rw.Code, Equals, http.StatusInternalServerError)
}

// func (s *ErrorTestSuite) TestMultipleErrorHandlers2(c *C) {
//   router := New(Context{})
//   router.Get("/action", (*Context).Action)
//   
//   admin := router.Subrouter(AdminContext{}, "/admin")
//   admin.ErrorHandler((*AdminContext).ErrorHandler)
//   admin.Get("/action", (*AdminContext).Action)
//   
//   api := router.Subrouter(ApiContext{}, "/api")
//   api.ErrorHandler((*ApiContext).ErrorHandler)
//   api.Get("/action", (*ApiContext).Action)
//   
//   rw, req := newTestRequest("GET", "/action")
//   router.ServeHTTP(rw, req)
//   c.Assert(strings.TrimSpace(string(rw.Body.Bytes())), Equals, "Application Error")
//   c.Assert(rw.Code, Equals, http.StatusInternalServerError)
//   
//   rw, req = newTestRequest("GET", "/admin/action")
//   router.ServeHTTP(rw, req)
//   c.Assert(strings.TrimSpace(string(rw.Body.Bytes())), Equals, "Admin Error")
//   c.Assert(rw.Code, Equals, http.StatusInternalServerError)
//   
//   rw, req = newTestRequest("GET", "/admin/action")
//   router.ServeHTTP(rw, req)
//   c.Assert(strings.TrimSpace(string(rw.Body.Bytes())), Equals, "Api Error")
//   c.Assert(rw.Code, Equals, http.StatusInternalServerError)
// }



// Things I want to test:
// - handler on api, exception on admin
// - panic in middlware triggers handler of target







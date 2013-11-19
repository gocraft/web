package web_test

import (
  "github.com/cypriss/mars_web"
  . "launchpad.net/gocheck"
  "fmt"
  "net/http"
  "strings"
)

type ErrorTestSuite struct {}
var _ = Suite(&ErrorTestSuite{})

func ErrorHandlerWithNoContext(w web.ResponseWriter, r *web.Request, err interface{}) {
  fmt.Fprintf(w, "Contextless Error")
}

func (s *ErrorTestSuite) TestNoErorHandler(c *C) {
  router := web.New(Context{})
  router.Get("/action", (*Context).ErrorAction)
  
  admin := router.Subrouter(AdminContext{}, "/admin")
  admin.Get("/action", (*AdminContext).ErrorAction)
  
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
  router := web.New(Context{})
  router.ErrorHandler((*Context).ErrorHandler)
  router.Get("/action", (*Context).ErrorAction)
  
  admin := router.Subrouter(AdminContext{}, "/admin")
  admin.Get("/action", (*AdminContext).ErrorAction)
  
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
  router := web.New(Context{})
  router.ErrorHandler(ErrorHandlerWithNoContext)
  router.Get("/action", (*Context).ErrorAction)
  
  admin := router.Subrouter(AdminContext{}, "/admin")
  admin.Get("/action", (*AdminContext).ErrorAction)
  
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
  router := web.New(Context{})
  router.ErrorHandler((*Context).ErrorHandler)
  router.Get("/action", (*Context).ErrorAction)
  
  admin := router.Subrouter(AdminContext{}, "/admin")
  admin.ErrorHandler((*AdminContext).ErrorHandler)
  admin.Get("/action", (*AdminContext).ErrorAction)
  
  rw, req := newTestRequest("GET", "/action")
  router.ServeHTTP(rw, req)
  c.Assert(strings.TrimSpace(string(rw.Body.Bytes())), Equals, "My Error")
  c.Assert(rw.Code, Equals, http.StatusInternalServerError)
  
  rw, req = newTestRequest("GET", "/admin/action")
  router.ServeHTTP(rw, req)
  c.Assert(strings.TrimSpace(string(rw.Body.Bytes())), Equals, "Admin Error")
  c.Assert(rw.Code, Equals, http.StatusInternalServerError)
}

func (s *ErrorTestSuite) TestMultipleErrorHandlers2(c *C) {
  router := web.New(Context{})
  router.Get("/action", (*Context).ErrorAction)
  
  admin := router.Subrouter(AdminContext{}, "/admin")
  admin.ErrorHandler((*AdminContext).ErrorHandler)
  admin.Get("/action", (*AdminContext).ErrorAction)
  
  api := router.Subrouter(ApiContext{}, "/api")
  api.ErrorHandler((*ApiContext).ErrorHandler)
  api.Get("/action", (*ApiContext).ErrorAction)
  
  rw, req := newTestRequest("GET", "/action")
  router.ServeHTTP(rw, req)
  c.Assert(strings.TrimSpace(string(rw.Body.Bytes())), Equals, "Application Error")
  c.Assert(rw.Code, Equals, http.StatusInternalServerError)
  
  rw, req = newTestRequest("GET", "/admin/action")
  router.ServeHTTP(rw, req)
  c.Assert(strings.TrimSpace(string(rw.Body.Bytes())), Equals, "Admin Error")
  c.Assert(rw.Code, Equals, http.StatusInternalServerError)
  
  rw, req = newTestRequest("GET", "/api/action")
  router.ServeHTTP(rw, req)
  c.Assert(strings.TrimSpace(string(rw.Body.Bytes())), Equals, "Api Error")
  c.Assert(rw.Code, Equals, http.StatusInternalServerError)
}



// Things I want to test:
// - panic in middlware triggers handler of target
// - if contexts don't change between subrouters, then what.







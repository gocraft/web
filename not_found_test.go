package web_test

import (
  "github.com/cypriss/mars_web"
  . "launchpad.net/gocheck"
  "fmt"
  "net/http"
  "strings"
)

type NotFoundTestSuite struct {}
var _ = Suite(&NotFoundTestSuite{})

func (s *NotFoundTestSuite) TestNoHandler(c *C) {
  router := web.New(Context{})
  
  rw, req := newTestRequest("GET", "/this_path_doesnt_exist")
  router.ServeHTTP(rw, req)
  c.Assert(strings.TrimSpace(string(rw.Body.Bytes())), Equals, "Not Found")
  c.Assert(rw.Code, Equals, http.StatusNotFound)
}

func (s *NotFoundTestSuite) TestBadMethod(c *C) {
  router := web.New(Context{})
  
  rw, req := newTestRequest("POOP", "/this_path_doesnt_exist")
  router.ServeHTTP(rw, req)
  c.Assert(strings.TrimSpace(string(rw.Body.Bytes())), Equals, "Not Found")
  c.Assert(rw.Code, Equals, http.StatusNotFound)
}


func MyNotFoundHandler(rw web.ResponseWriter, r *web.Request) {
  fmt.Fprintf(rw, "My Not Found")
}

func (s *NotFoundTestSuite) TestWithHandler(c *C) {
  router := web.New(Context{})
  router.NotFoundHandler(MyNotFoundHandler)
  
  rw, req := newTestRequest("GET", "/this_path_doesnt_exist")
  router.ServeHTTP(rw, req)
  c.Assert(strings.TrimSpace(string(rw.Body.Bytes())), Equals, "My Not Found")
  c.Assert(rw.Code, Equals, http.StatusNotFound)
}

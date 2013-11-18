package web_test

import (
  "github.com/cypriss/mars_web"
  . "launchpad.net/gocheck"
  // "fmt"
  "net/http"
  "strings"
)

type NotFoundTestSuite struct {}
var _ = Suite(&ErrorTestSuite{})

func (s *ErrorTestSuite) TestNoHandler(c *C) {
  router := web.New(Context{})
  
  rw, req := newTestRequest("GET", "/this_path_doesnt_exist")
  router.ServeHTTP(rw, req)
  c.Assert(strings.TrimSpace(string(rw.Body.Bytes())), Equals, "Not Found")
  c.Assert(rw.Code, Equals, http.StatusNotFound)
}

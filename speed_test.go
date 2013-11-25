package web_test

import (
  "github.com/cypriss/mars_web"
  "testing"
  "fmt"
  "net/http"
  "net/http/httptest"
  "crypto/sha1"
  "io"
)

func testRequest(method, path string) (*httptest.ResponseRecorder, *http.Request) {
  request, _ := http.NewRequest(method, path, nil)
  recorder := httptest.NewRecorder()

  return recorder, request
}

type BenchContext struct {}

func (c *BenchContext) Action(w web.ResponseWriter, r *web.Request) {
  fmt.Fprintf(w, "hello")
}

// Simplest benchmark ever.
// One router, one route. No middleware. Just calling the action.
func BenchmarkSimple(b *testing.B) {
  router := web.New(BenchContext{})
  router.Get("/action", (*BenchContext).Action)
  
  rw, req := testRequest("GET", "/action")
  
  b.ResetTimer()
  for i := 0; i < b.N; i++ {
    router.ServeHTTP(rw, req)
  }
}

// Determine routing performance as a function of the # of routes.
// We're going to use JSON restful routes here:
// a given 'resource' will have index/show/create/update/delete:
// GET /resources
// GET /resources/:id
// POST /resources
// PUT /resources/:id
// DELETE /resources/:id
// We'll have 3 namespaces. Each namespace will have N resource.
const NUM_ROUTES = 50
func BenchmarkRouting(b *testing.B) {
  namespaces := []string{"admin", "api", "site"}
  resources := []string{}
  
  for i := 0; i < NUM_ROUTES; i += 1 {
    sha1 := sha1.New()
    io.WriteString(sha1, fmt.Sprintf("%d", i))
    strResource := fmt.Sprintf("%x", sha1.Sum(nil))
    resources = append(resources, strResource)
  }
  
  rootRouter := web.New(BenchContext{})
  
  for _, ns := range namespaces {
    subrouter := rootRouter.Subrouter(BenchContext{}, "/" + ns)
    
    for _, res := range resources {
      subrouter.Get("/" + res, (*BenchContext).Action)
      subrouter.Post("/" + res, (*BenchContext).Action)
      subrouter.Get("/" + res + "/:id", (*BenchContext).Action)
      subrouter.Put("/" + res + "/:id", (*BenchContext).Action)
      subrouter.Delete("/" + res + "/:id", (*BenchContext).Action)
    }
  }
  
  
  recorder := httptest.NewRecorder()
  requests := []*http.Request{}
  for _, ns := range namespaces {
    for _, res := range resources {
      req, _ := http.NewRequest("GET", "/" + ns + "/" + res, nil)
      requests = append(requests, req)
      req, _ = http.NewRequest("POST", "/" + ns + "/" + res, nil)
      requests = append(requests, req)
      req, _ = http.NewRequest("GET", "/" + ns + "/" + res + "/3937", nil)
      requests = append(requests, req)
      req, _ = http.NewRequest("PUT", "/" + ns + "/" + res + "/3937", nil)
      requests = append(requests, req)
      req, _ = http.NewRequest("DELETE", "/" + ns + "/" + res + "/3937", nil)
      requests = append(requests, req)
    }
  }
  
  reqId := 0
  
  b.ResetTimer()
  for i := 0; i < b.N; i++ {
    if reqId >= len(requests) {
      reqId = 0
    }
    req := requests[reqId]
    
    rootRouter.ServeHTTP(recorder, req)
    
    //if recorder.Code != 200 {
    //  panic("wat")
    //}
    
    reqId += 1
  }
  
}



package web_test

import (
  "github.com/cypriss/mars_web"
  "testing"
  "fmt"
  "strings"
  "net/http"
  "net/http/httptest"
)

//
// We're going to test everything from an integration perspective b/c I don't want to expose
// the tree.go guts.
//

type Ctx struct {}

type routeTest struct {
  route string
  get string
  vars map[string]string
}

// Converts the map into a consistent, string-comparable string (to compare with another map)
// Eg, stringifyMap({"foo": "bar"}) == stringifyMap({"foo": "bar"})
func stringifyMap(m map[string]string) string {
  if len(m) == 0 {
    return ""
  }
  
  return fmt.Sprint(m)  // NOTE: this seems to work. If you get re-ordering issues, then feel free to re-impl
}

func TestRoutes(t *testing.T) {
  router := web.New(Ctx{})
  
  table := []routeTest{
    {
      route: "/",
      get: "/",
      vars: nil,
    },
    {
      route: "/api/action",
      get: "/api/action",
      vars: nil,
    },
    {
      route: "/admin/action",
      get: "/admin/action",
      vars: nil,
    },
    {
      route: "/admin/action.json",
      get: "/admin/action.json",
      vars: nil,
    },
    {
      route: "/:api/action",
      get: "/poop/action",
      vars: map[string]string{"api": "poop"},
    },
    {
      route: "/api/:action",
      get: "/api/poop",
      vars: map[string]string{"action": "poop"},
    },
    {
      route: "/:seg1/:seg2/bob",
      get: "/a/b/bob",
      vars: map[string]string{"seg1": "a","seg2": "b"},
    },
    {
      route: "/:seg1/:seg2/ron",
      get: "/c/d/ron",
      vars: map[string]string{"seg1": "c", "seg2": "d"},
    },
    {
      route: "/:seg1/:seg2/:seg3",
      get: "/c/d/wat",
      vars: map[string]string{"seg1": "c", "seg2": "d", "seg3": "wat"},
    },
    {
      route: "/:seg1/:seg2/ron/apple",
      get: "/c/d/ron/apple",
      vars: map[string]string{"seg1": "c", "seg2": "d"},
    },
    {
      route: "/:seg1/:seg2/ron/:apple",
      get: "/c/d/ron/orange",
      vars: map[string]string{"seg1": "c", "seg2": "d", "apple": "orange"},
    },
  }
  
  // Create routes
  for _, rt := range table {
    // func: ensure closure is created per iteraction (it fails otherwise)
    func(exp string) {
      router.Get(rt.route, func(w web.ResponseWriter, r *web.Request) {
        w.Header().Set("X-VARS", stringifyMap(r.UrlVariables))
        fmt.Fprintf(w, exp)
      })
    }(rt.route)
  }
  
  // Execute them all:
  for _, rt := range table {
    fmt.Println("doing ", rt)
    recorder := httptest.NewRecorder()
    request, _ := http.NewRequest("GET", rt.get, nil)
    
    router.ServeHTTP(recorder, request)
    
    if recorder.Code != 200 {
      t.Error("Test:", rt, " Didn't get Code=200. Got Code=",recorder.Code)
    }
    body := strings.TrimSpace(string(recorder.Body.Bytes()))
    if body != rt.route {
      t.Error("Test:", rt, " Didn't get Body=", rt.route, ". Got Body=", body)
    }
    vars := recorder.Header().Get("X-VARS")
    if vars != stringifyMap(rt.vars) {
      t.Error("Test:", rt, " Didn't get Vars=", rt.vars, ". Got Vars=", vars)
    }
  }
}

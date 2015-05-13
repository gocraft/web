package web

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
)

//
// We're going to test everything from an integration perspective b/c I don't want to expose
// the tree.go guts.
//

type Ctx struct{}

type routeTest struct {
	route string
	get   string
	vars  map[string]string
}

// Converts the map into a consistent, string-comparable string (to compare with another map)
// Eg, stringifyMap({"foo": "bar"}) == stringifyMap({"foo": "bar"})
func stringifyMap(m map[string]string) string {
	if m == nil {
		return ""
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	keysLenMinusOne := len(keys) - 1

	var b bytes.Buffer

	b.WriteString("[")
	for i, k := range keys {
		b.WriteString(k)
		b.WriteRune(':')
		b.WriteString(m[k])

		if i != keysLenMinusOne {
			b.WriteRune(' ')
		}
	}
	b.WriteRune(']')

	return b.String()
}

func TestRoutes(t *testing.T) {
	router := New(Ctx{})

	table := []routeTest{
		{
			route: "/",
			get:   "/",
			vars:  nil,
		},
		{
			route: "/api/action",
			get:   "/api/action",
			vars:  nil,
		},
		{
			route: "/admin/action",
			get:   "/admin/action",
			vars:  nil,
		},
		{
			route: "/admin/action.json",
			get:   "/admin/action.json",
			vars:  nil,
		},
		{
			route: "/:api/action",
			get:   "/poop/action",
			vars:  map[string]string{"api": "poop"},
		},
		{
			route: "/api/:action",
			get:   "/api/poop",
			vars:  map[string]string{"action": "poop"},
		},
		{
			route: "/:seg1/:seg2/bob",
			get:   "/a/b/bob",
			vars:  map[string]string{"seg1": "a", "seg2": "b"},
		},
		{
			route: "/:seg1/:seg2/ron",
			get:   "/c/d/ron",
			vars:  map[string]string{"seg1": "c", "seg2": "d"},
		},
		{
			route: "/:seg1/:seg2/:seg3",
			get:   "/c/d/wat",
			vars:  map[string]string{"seg1": "c", "seg2": "d", "seg3": "wat"},
		},
		{
			route: "/:seg1/:seg2/ron/apple",
			get:   "/c/d/ron/apple",
			vars:  map[string]string{"seg1": "c", "seg2": "d"},
		},
		{
			route: "/:seg1/:seg2/ron/:apple",
			get:   "/c/d/ron/orange",
			vars:  map[string]string{"seg1": "c", "seg2": "d", "apple": "orange"},
		},
		{
			route: "/site2/:id:\\d+",
			get:   "/site2/123",
			vars:  map[string]string{"id": "123"},
		},
		{
			route: "/site2/:id:[a-z]+",
			get:   "/site2/abc",
			vars:  map[string]string{"id": "abc"},
		},
		{
			route: "/site2/:id:\\d[a-z]+",
			get:   "/site2/1abc",
			vars:  map[string]string{"id": "1abc"},
		},
		{
			route: "/site2/:id",
			get:   "/site2/1abc1",
			vars:  map[string]string{"id": "1abc1"},
		},
		{
			route: "/site2/:id:\\d+/other/:var:[A-Z]+",
			get:   "/site2/123/other/OK",
			vars:  map[string]string{"id": "123", "var": "OK"},
		},
		{
			route: "/site2/:id/:*",
			get:   "/site2/1abc1/foo/bar/baz/boo",
			vars:  map[string]string{"id": "1abc1", "*": "foo/bar/baz/boo"},
		},
		{
			route: "/site3/:id:\\d+/:*",
			get:   "/site3/123/foo/bar/baz/boo",
			vars:  map[string]string{"id": "123", "*": "foo/bar/baz/boo"},
		},
		{
			route: "/site3/:*",
			get:   "/site3/foo/bar/baz/boo",
			vars:  map[string]string{"*": "foo/bar/baz/boo"},
		},
	}

	// Create routes
	for _, rt := range table {
		//func: ensure closure is created per iteraction (it fails otherwise)
		func(exp string) {
			router.Get(rt.route, func(w ResponseWriter, r *Request) {
				w.Header().Set("X-VARS", stringifyMap(r.PathParams))
				fmt.Fprintf(w, exp)
			})
		}(rt.route)
	}

	// Execute them all:
	for _, rt := range table {
		recorder := httptest.NewRecorder()
		request, _ := http.NewRequest("GET", rt.get, nil)

		router.ServeHTTP(recorder, request)

		if recorder.Code != 200 {
			t.Error("Test:", rt, " Didn't get Code=200. Got Code=", recorder.Code)
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

func TestRoutesWithPrefix(t *testing.T) {
	router := NewWithPrefix(Ctx{}, "/v1")

	table := []routeTest{
		{
			route: "/",
			get:   "/v1/",
			vars:  nil,
		},
		{
			route: "/api/action",
			get:   "/v1/api/action",
			vars:  nil,
		},
		{
			route: "/admin/action",
			get:   "/v1/admin/action",
			vars:  nil,
		},
		{
			route: "/admin/action.json",
			get:   "/v1/admin/action.json",
			vars:  nil,
		},
		{
			route: "/:api/action",
			get:   "/v1/poop/action",
			vars:  map[string]string{"api": "poop"},
		},
		{
			route: "/api/:action",
			get:   "/v1/api/poop",
			vars:  map[string]string{"action": "poop"},
		},
		{
			route: "/:seg1/:seg2/bob",
			get:   "/v1/a/b/bob",
			vars:  map[string]string{"seg1": "a", "seg2": "b"},
		},
		{
			route: "/:seg1/:seg2/ron",
			get:   "/v1/c/d/ron",
			vars:  map[string]string{"seg1": "c", "seg2": "d"},
		},
		{
			route: "/:seg1/:seg2/:seg3",
			get:   "/v1/c/d/wat",
			vars:  map[string]string{"seg1": "c", "seg2": "d", "seg3": "wat"},
		},
		{
			route: "/:seg1/:seg2/ron/apple",
			get:   "/v1/c/d/ron/apple",
			vars:  map[string]string{"seg1": "c", "seg2": "d"},
		},
		{
			route: "/:seg1/:seg2/ron/:apple",
			get:   "/v1/c/d/ron/orange",
			vars:  map[string]string{"seg1": "c", "seg2": "d", "apple": "orange"},
		},
		{
			route: "/site2/:id:\\d+",
			get:   "/v1/site2/123",
			vars:  map[string]string{"id": "123"},
		},
		{
			route: "/site2/:id:[a-z]+",
			get:   "/v1/site2/abc",
			vars:  map[string]string{"id": "abc"},
		},
		{
			route: "/site2/:id:\\d[a-z]+",
			get:   "/v1/site2/1abc",
			vars:  map[string]string{"id": "1abc"},
		},
		{
			route: "/site2/:id",
			get:   "/v1/site2/1abc1",
			vars:  map[string]string{"id": "1abc1"},
		},
		{
			route: "/site2/:id:\\d+/other/:var:[A-Z]+",
			get:   "/v1/site2/123/other/OK",
			vars:  map[string]string{"id": "123", "var": "OK"},
		},
	}

	// Create routes
	for _, rt := range table {
		// func: ensure closure is created per iteraction (it fails otherwise)
		func(exp string) {
			router.Get(rt.route, func(w ResponseWriter, r *Request) {
				w.Header().Set("X-VARS", stringifyMap(r.PathParams))
				fmt.Fprintf(w, exp)
			})
		}(rt.route)
	}

	// Execute them all:
	for _, rt := range table {
		recorder := httptest.NewRecorder()
		request, _ := http.NewRequest("GET", rt.get, nil)

		router.ServeHTTP(recorder, request)

		if recorder.Code != 200 {
			t.Error("Test:", rt, " Didn't get Code=200. Got Code=", recorder.Code)
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

func TestRouteVerbs(t *testing.T) {
	router := New(Context{})
	router.Get("/a", func(w ResponseWriter, r *Request) {
		fmt.Fprintf(w, "GET")
	})
	router.Put("/a", func(w ResponseWriter, r *Request) {
		fmt.Fprintf(w, "PUT")
	})
	router.Post("/a", func(w ResponseWriter, r *Request) {
		fmt.Fprintf(w, "POST")
	})
	router.Delete("/a", func(w ResponseWriter, r *Request) {
		fmt.Fprintf(w, "DELETE")
	})
	router.Patch("/a", func(w ResponseWriter, r *Request) {
		fmt.Fprintf(w, "PATCH")
	})
	router.Head("/a", func(w ResponseWriter, r *Request) {
		fmt.Fprintf(w, "HEAD")
	})
	router.Options("/a", func(w ResponseWriter, r *Request) {
		fmt.Fprintf(w, "OPTIONS")
	})

	for _, method := range httpMethods {
		method := string(method)

		recorder := httptest.NewRecorder()
		request, _ := http.NewRequest(method, "/a", nil)

		router.ServeHTTP(recorder, request)

		if recorder.Code != 200 {
			t.Error("Test:", method, " Didn't get Code=200. Got Code=", recorder.Code)
		}

		body := strings.TrimSpace(string(recorder.Body.Bytes()))
		if body != method {
			t.Error("Test:", method, " Didn't get Body=", method, ". Got Body=", body)
		}
	}
}

func TestRouteHead(t *testing.T) {
	router := New(Context{})
	router.Get("/a", (*Context).A)

	rw, req := newTestRequest("GET", "/a")
	router.ServeHTTP(rw, req)
	assertResponse(t, rw, "context-A", 200)

	rw, req = newTestRequest("HEAD", "/a")
	router.ServeHTTP(rw, req)
	assertResponse(t, rw, "context-A", 200)
}

func TestIsRouted(t *testing.T) {
	router := New(Context{})
	router.Middleware(func(w ResponseWriter, r *Request, next NextMiddlewareFunc) {
		if r.IsRouted() {
			t.Error("Shouldn't be routed yet but was.")
		}
		if r.RoutePath() != "" {
			t.Error("Shouldn't have a route path yet.")
		}
		next(w, r)
		if !r.IsRouted() {
			t.Error("Should have been routed but wasn't.")
		}
	})
	subrouter := router.Subrouter(Context{}, "")
	subrouter.Middleware(func(w ResponseWriter, r *Request, next NextMiddlewareFunc) {
		if !r.IsRouted() {
			t.Error("Should have been routed but wasn't.")
		}
		next(w, r)
		if !r.IsRouted() {
			t.Error("Should have been routed but wasn't.")
		}
	})
	subrouter.Get("/a", func(w ResponseWriter, r *Request) {
		fmt.Fprintf(w, r.RoutePath())
	})

	rw, req := newTestRequest("GET", "/a")
	router.ServeHTTP(rw, req)
	assertResponse(t, rw, "/a", 200)
}

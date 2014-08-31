package web

import (
	"fmt"
	"net/http"
	"testing"
)

func MyNotFoundHandler(rw ResponseWriter, r *Request) {
	rw.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(rw, "My Not Found")
}

func (c *Context) HandlerWithContext(rw ResponseWriter, r *Request) {
	rw.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(rw, "My Not Found With Context")
}

func TestNoHandler(t *testing.T) {
	router := New(Context{})

	rw, req := newTestRequest("GET", "/this_path_doesnt_exist")
	router.ServeHTTP(rw, req)
	assertResponse(t, rw, "Not Found", http.StatusNotFound)
}

func TestBadMethod(t *testing.T) {
	router := New(Context{})

	rw, req := newTestRequest("POOP", "/this_path_doesnt_exist")
	router.ServeHTTP(rw, req)
	assertResponse(t, rw, "Not Found", http.StatusNotFound)
}

func TestWithHandler(t *testing.T) {
	router := New(Context{})
	router.NotFound(MyNotFoundHandler)

	rw, req := newTestRequest("GET", "/this_path_doesnt_exist")
	router.ServeHTTP(rw, req)
	assertResponse(t, rw, "My Not Found", http.StatusNotFound)
}

func TestWithRootContext(t *testing.T) {
	router := New(Context{})
	router.NotFound((*Context).HandlerWithContext)

	rw, req := newTestRequest("GET", "/this_path_doesnt_exist")
	router.ServeHTTP(rw, req)
	assertResponse(t, rw, "My Not Found With Context", http.StatusNotFound)
}

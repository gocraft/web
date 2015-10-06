package web

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestOptionsHandler(t *testing.T) {
	router := New(Context{})

	sub := router.Subrouter(Context{}, "/sub")
	sub.Middleware(AccessControlMiddleware)
	sub.Get("/action", (*Context).A)
	sub.Put("/action", (*Context).A)

	rw, req := newTestRequest("OPTIONS", "/sub/action")
	router.ServeHTTP(rw, req)
	assertResponse(t, rw, "", 200)
	assert.Equal(t, "GET, PUT", rw.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "*", rw.Header().Get("Access-Control-Allow-Origin"))

	rw, req = newTestRequest("GET", "/sub/action")
	router.ServeHTTP(rw, req)
	assert.Equal(t, "*", rw.Header().Get("Access-Control-Allow-Origin"))
}

func TestCustomOptionsHandler(t *testing.T) {
	router := New(Context{})
	router.OptionsHandler((*Context).OptionsHandler)

	sub := router.Subrouter(Context{}, "/sub")
	sub.Middleware(AccessControlMiddleware)
	sub.Get("/action", (*Context).A)
	sub.Put("/action", (*Context).A)

	rw, req := newTestRequest("OPTIONS", "/sub/action")
	router.ServeHTTP(rw, req)
	assertResponse(t, rw, "", 200)
	assert.Equal(t, "GET, PUT", rw.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "100", rw.Header().Get("Access-Control-Max-Age"))
}

func (c *Context) OptionsHandler(rw ResponseWriter, req *Request, methods []string) {
	rw.Header().Add("Access-Control-Allow-Methods", strings.Join(methods, ", "))
	rw.Header().Add("Access-Control-Max-Age", "100")
}

func AccessControlMiddleware(rw ResponseWriter, req *Request, next NextMiddlewareFunc) {
	rw.Header().Add("Access-Control-Allow-Origin", "*")
	next(rw, req)
}

package web

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func ErrorHandlerWithNoContext(w ResponseWriter, r *Request, err interface{}) {
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, "Contextless Error")
}

func TestNoErrorHandler(t *testing.T) {
	router := New(Context{})
	router.Get("/action", (*Context).ErrorAction)

	admin := router.Subrouter(AdminContext{}, "/admin")
	admin.Get("/action", (*AdminContext).ErrorAction)

	rw, req := newTestRequest("GET", "/action")
	router.ServeHTTP(rw, req)
	assertResponse(t, rw, "Application Errdor", http.StatusInternalServerError)

	rw, req = newTestRequest("GET", "/admin/action")
	router.ServeHTTP(rw, req)
	assertResponse(t, rw, "Application Error", http.StatusInternalServerError)
}

func TestHandlerOnRoot(t *testing.T) {
	router := New(Context{})
	router.Error((*Context).ErrorHandler)
	router.Get("/action", (*Context).ErrorAction)

	admin := router.Subrouter(AdminContext{}, "/admin")
	admin.Get("/action", (*AdminContext).ErrorAction)

	rw, req := newTestRequest("GET", "/action")
	router.ServeHTTP(rw, req)
	assertResponse(t, rw, "My Error", http.StatusInternalServerError)

	rw, req = newTestRequest("GET", "/admin/action")
	router.ServeHTTP(rw, req)
	assertResponse(t, rw, "My Error", http.StatusInternalServerError)
}

func TestContextlessError(t *testing.T) {
	router := New(Context{})
	router.Error(ErrorHandlerWithNoContext)
	router.Get("/action", (*Context).ErrorAction)

	admin := router.Subrouter(AdminContext{}, "/admin")
	admin.Get("/action", (*AdminContext).ErrorAction)

	rw, req := newTestRequest("GET", "/action")
	router.ServeHTTP(rw, req)
	assert.Equal(t, http.StatusInternalServerError, rw.Code)
	assertResponse(t, rw, "Contextless Error", http.StatusInternalServerError)

	rw, req = newTestRequest("GET", "/admin/action")
	router.ServeHTTP(rw, req)
	assertResponse(t, rw, "Contextless Error", http.StatusInternalServerError)
}

func TestMultipleErrorHandlers(t *testing.T) {
	router := New(Context{})
	router.Error((*Context).ErrorHandler)
	router.Get("/action", (*Context).ErrorAction)

	admin := router.Subrouter(AdminContext{}, "/admin")
	admin.Error((*AdminContext).ErrorHandler)
	admin.Get("/action", (*AdminContext).ErrorAction)

	rw, req := newTestRequest("GET", "/action")
	router.ServeHTTP(rw, req)
	assertResponse(t, rw, "My Error", http.StatusInternalServerError)

	rw, req = newTestRequest("GET", "/admin/action")
	router.ServeHTTP(rw, req)
	assertResponse(t, rw, "Admin Error", http.StatusInternalServerError)
}

func TestMultipleErrorHandlers2(t *testing.T) {
	router := New(Context{})
	router.Get("/action", (*Context).ErrorAction)

	admin := router.Subrouter(AdminContext{}, "/admin")
	admin.Error((*AdminContext).ErrorHandler)
	admin.Get("/action", (*AdminContext).ErrorAction)

	api := router.Subrouter(APIContext{}, "/api")
	api.Error((*APIContext).ErrorHandler)
	api.Get("/action", (*APIContext).ErrorAction)

	rw, req := newTestRequest("GET", "/action")
	router.ServeHTTP(rw, req)
	assertResponse(t, rw, "Application Error", http.StatusInternalServerError)

	rw, req = newTestRequest("GET", "/admin/action")
	router.ServeHTTP(rw, req)
	assertResponse(t, rw, "Admin Error", http.StatusInternalServerError)

	rw, req = newTestRequest("GET", "/api/action")
	router.ServeHTTP(rw, req)
	assertResponse(t, rw, "Api Error", http.StatusInternalServerError)
}

func TestRootMiddlewarePanic(t *testing.T) {
	router := New(Context{})
	router.Middleware((*Context).ErrorMiddleware)
	router.Error((*Context).ErrorHandler)
	admin := router.Subrouter(AdminContext{}, "/admin")
	admin.Error((*AdminContext).ErrorHandler)
	admin.Get("/action", (*AdminContext).ErrorAction)

	rw, req := newTestRequest("GET", "/admin/action")
	router.ServeHTTP(rw, req)
	assertResponse(t, rw, "My Error", 500)
}

func TestNonRootMiddlewarePanic(t *testing.T) {
	router := New(Context{})
	router.Error((*Context).ErrorHandler)
	admin := router.Subrouter(AdminContext{}, "/admin")
	admin.Middleware((*AdminContext).ErrorMiddleware)
	admin.Error((*AdminContext).ErrorHandler)
	admin.Get("/action", (*AdminContext).ErrorAction)

	rw, req := newTestRequest("GET", "/admin/action")
	router.ServeHTTP(rw, req)
	assertResponse(t, rw, "Admin Error", 500)
}

func TestConsistentContext(t *testing.T) {
	router := New(Context{})
	router.Error((*Context).ErrorHandler)
	admin := router.Subrouter(Context{}, "/admin")
	admin.Error((*Context).ErrorHandlerSecondary)
	admin.Get("/foo", (*Context).ErrorAction)

	rw, req := newTestRequest("GET", "/admin/foo")
	router.ServeHTTP(rw, req)
	assertResponse(t, rw, "My Secondary Error", 500)
}

package web

import (
	"strings"
	"testing"
)

func TestShowErrorsMiddleware(t *testing.T) {
	router := New(Context{})
	router.Middleware(ShowErrorsMiddleware)
	router.Get("/action", (*Context).A)
	router.Get("/boom", (*Context).ErrorAction)

	// Success:
	rw, req := newTestRequest("GET", "/action")
	router.ServeHTTP(rw, req)
	assertResponse(t, rw, "context-A", 200)

	// Boom:
	rw, req = newTestRequest("GET", "/boom")
	router.ServeHTTP(rw, req)
	if rw.Code != 500 {
		t.Errorf("Expected status code 500 but got %d", rw.Code)
	}

	body := strings.TrimSpace(string(rw.Body.Bytes()))
	if !strings.HasPrefix(body, "<html>") {
		t.Errorf("Expected an HTML page but got '%s'", body)
	}
}

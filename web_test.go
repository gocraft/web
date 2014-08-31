package web

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"
)

//
// This file will contain helpers and general things the rest of the suite needs
//

type nullPanicReporter struct{}

func (l nullPanicReporter) Panic(url string, err interface{}, stack string) {
	// no op
}
func init() {
	// This disables printing panics to stderr during testing, because that is very noisy,
	// and we purposefully test some panics.
	PanicHandler = nullPanicReporter{}
}

// Return's the caller's caller info.
func callerInfo() string {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		return ""
	}
	parts := strings.Split(file, "/")
	file = parts[len(parts)-1]
	return fmt.Sprintf("%s:%d", file, line)
}

// Make a testing request
func newTestRequest(method, path string) (*httptest.ResponseRecorder, *http.Request) {
	request, _ := http.NewRequest(method, path, nil)
	recorder := httptest.NewRecorder()

	return recorder, request
}

func assertResponse(t *testing.T, rr *httptest.ResponseRecorder, body string, code int) {
	if gotBody := strings.TrimSpace(string(rr.Body.Bytes())); body != gotBody {
		t.Errorf("assertResponse: expected body to be %s but got %s. (caller: %s)", body, gotBody, callerInfo())
	}
	if code != rr.Code {
		t.Errorf("assertResponse: expected code to be %d but got %d. (caller: %s)", code, rr.Code, callerInfo())
	}
}

//
// Some default contexts and possible error handlers / actions
//
type Context struct{}

type AdminContext struct {
	*Context
}

type APIContext struct {
	*Context
}

type SiteContext struct {
	*Context
}

type TicketsContext struct {
	*AdminContext
}

func (c *Context) ErrorMiddleware(w ResponseWriter, r *Request, next NextMiddlewareFunc) {
	var x, y int
	fmt.Fprintln(w, x/y)
}

func (c *Context) ErrorHandler(w ResponseWriter, r *Request, err interface{}) {
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, "My Error")
}

func (c *Context) ErrorHandlerSecondary(w ResponseWriter, r *Request, err interface{}) {
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, "My Secondary Error")
}

func (c *Context) ErrorAction(w ResponseWriter, r *Request) {
	var x, y int
	fmt.Fprintln(w, x/y)
}

func (c *AdminContext) ErrorMiddleware(w ResponseWriter, r *Request, next NextMiddlewareFunc) {
	var x, y int
	fmt.Fprintln(w, x/y)
}

func (c *AdminContext) ErrorHandler(w ResponseWriter, r *Request, err interface{}) {
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, "Admin Error")
}

func (c *AdminContext) ErrorAction(w ResponseWriter, r *Request) {
	var x, y int
	fmt.Fprintln(w, x/y)
}

func (c *APIContext) ErrorHandler(w ResponseWriter, r *Request, err interface{}) {
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, "Api Error")
}

func (c *APIContext) ErrorAction(w ResponseWriter, r *Request) {
	var x, y int
	fmt.Fprintln(w, x/y)
}

func (c *TicketsContext) ErrorAction(w ResponseWriter, r *Request) {
	var x, y int
	fmt.Fprintln(w, x/y)
}

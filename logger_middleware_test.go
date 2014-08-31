package web

import (
	"bytes"
	"log"
	"regexp"
	"testing"
)

func TestLoggerMiddleware(t *testing.T) {
	var buf bytes.Buffer
	Logger = log.New(&buf, "", 0)

	router := New(Context{})
	router.Middleware(LoggerMiddleware)
	router.Get("/action", (*Context).A)

	// Hit an action:
	rw, req := newTestRequest("GET", "/action")
	router.ServeHTTP(rw, req)
	assertResponse(t, rw, "context-A", 200)

	// Make sure our buf has something good:
	logRegexp := regexp.MustCompile("\\[\\d+ .{2}\\] 200 '/action'")
	if !logRegexp.MatchString(buf.String()) {
		t.Error("Got invalid log entry: ", buf.String())
	}

	// Do a 404:
	buf.Reset()
	rw, req = newTestRequest("GET", "/wat")
	router.ServeHTTP(rw, req)
	assertResponse(t, rw, "Not Found", 404)

	// Make sure our buf has something good:
	logRegexpNotFound := regexp.MustCompile("\\[\\d+ .{2}\\] 404 '/wat'")
	if !logRegexpNotFound.MatchString(buf.String()) {
		t.Error("Got invalid log entry: ", buf.String())
	}
}

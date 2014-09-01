package web

import (
	"net/http"
)

// ResponseWriter includes net/http's ResponseWriter and adds a StatusCode() method to obtain the written status code.
// A ResponseWriter is sent to handlers on each request.
type ResponseWriter interface {
	http.ResponseWriter

	// StatusCode returns the written status code, or 0 if none has been written yet.
	StatusCode() int
}

type appResponseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

// Don't need this yet because we get it for free:
func (w *appResponseWriter) Write(data []byte) (n int, err error) {
	if !w.written {
		w.statusCode = http.StatusOK
		w.written = true
	}
	return w.ResponseWriter.Write(data)
}

func (w *appResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.written = true
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *appResponseWriter) StatusCode() int {
	return w.statusCode
}

package web

import (
	"net/http"
)

type ResponseWriter interface {
	http.ResponseWriter
	StatusCode() int
}

type AppResponseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

// Don't need this yet because we get it for free:
func (w *AppResponseWriter) Write(data []byte) (n int, err error) {
	if !w.written {
		w.statusCode = http.StatusOK
		w.written = true
	}
	return w.ResponseWriter.Write(data)
}

func (w *AppResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.written = true
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *AppResponseWriter) StatusCode() int {
	return w.statusCode
}

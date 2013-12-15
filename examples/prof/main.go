package main

import (
	"github.com/gocraft/web"
	"fmt"
	"os"
	"log"
	"runtime/pprof"
	// "net/http/httptest"
	"net/http"
	"runtime"
)

//
// Null response writer
//
type NullWriter struct {}

func (w *NullWriter) Header() http.Header {
	return nil
}


func (w *NullWriter) Write(data []byte) (n int, err error) {
	return len(data), nil
}

func (w *NullWriter) WriteHeader(statusCode int) {}

//
// Simple app with middleware
//

type Context struct {
	
}

func (c *Context) Action(w web.ResponseWriter, r *web.Request) {
	fmt.Fprintf(w, "hello")
}

func (c *Context) Middleware(w web.ResponseWriter, r *web.Request, next web.NextMiddlewareFunc) {
	next(w, r)
}

func main() {
	runtime.MemProfileRate = 1
	
	router := web.New(Context{}).
		Middleware((*Context).Middleware).
		Middleware((*Context).Middleware).
		Middleware((*Context).Middleware).
		Middleware((*Context).Middleware).
		Middleware((*Context).Middleware).
		Middleware((*Context).Middleware).
		Get("/action", (*Context).Action)
	
	rw := &NullWriter{}
	req, _ := http.NewRequest("GET", "/action", nil)
	
	
	// pprof.StartCPUProfile(f)
	// defer pprof.StopCPUProfile()
	
	for i := 0; i < 1; i += 1 {
		router.ServeHTTP(rw, req)
	}
	
	f, err := os.Create("myprof.out")
	if err != nil {
	    log.Fatal(err)
	}
	
	pprof.WriteHeapProfile(f)
	f.Close()
}
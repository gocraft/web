package main

import (
	"github.com/gocraft/web"
	"fmt"
	"os"
	"log"
	"runtime/pprof"
	"net/http/httptest"
	"net/http"
)

type Context struct {
	
}

func (c *Context) Action(w web.ResponseWriter, r *web.Request) {
	fmt.Fprintf(w, "hello")
}

func (c *Context) Middleware(w web.ResponseWriter, r *web.Request, next web.NextMiddlewareFunc) {
	next(w, r)
}

func main() {
	f, err := os.Create("mycpuprof.out")
	if err != nil {
	    log.Fatal(err)
	}
	
	router := web.New(Context{}).
		Middleware((*Context).Middleware).
		Middleware((*Context).Middleware).
		Middleware((*Context).Middleware).
		Middleware((*Context).Middleware).
		Middleware((*Context).Middleware).
		Middleware((*Context).Middleware).
		Get("/action", (*Context).Action)
	
	rw := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/action", nil)
	
	
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()
	
	
	
	
	for i := 0; i < 1000000; i += 1 {
		router.ServeHTTP(rw, req)
	}
	
}
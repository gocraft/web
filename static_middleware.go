package web

import (
	"net/http"
	"path/filepath"
)

// StaticMiddleware("public") returns proper middleware
// NOTE: original impl is from github.com/codegangsta/martini
func StaticMiddleware(path string) func(ResponseWriter, *Request, NextMiddlewareFunc) {
	dir := http.Dir(path)
	return func(w ResponseWriter, req *Request, next NextMiddlewareFunc) {
		file := req.URL.Path
		f, err := dir.Open(file)
		if err != nil {
			next(w, req)
			return
		}
		defer f.Close()

		fi, err := f.Stat()
		if err != nil {
			next(w, req)
			return
		}

		// Try to serve index.html
		if fi.IsDir() {
			file = filepath.Join(file, "index.html")
			f, err = dir.Open(file)
			if err != nil {
				next(w, req)
				return
			}
			defer f.Close()

			fi, err = f.Stat()
			if err != nil || fi.IsDir() {
				next(w, req)
				return
			}
		}

		http.ServeContent(w, req.Request, file, fi.ModTime(), f)
	}
}

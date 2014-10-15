package web

import (
	"net/http"
	"path/filepath"
)

// StaticMiddleware is the same as StaticMiddlewareFromDir, but accepts a
// path string for backwards compatibility.
func StaticMiddleware(path string) func(ResponseWriter, *Request, NextMiddlewareFunc) {
	return StaticMiddlewareFromDir(http.Dir(path))
}

// StaticMiddlewareFromDir returns a middleware that serves static files from the specified http.FileSystem.
// This middleware is great for development because each file is read from disk each time and no
// special caching or cache headers are sent.
//
// If a path is requested which maps to a folder with an index.html folder on your filesystem,
// then that index.html file will be served.
func StaticMiddlewareFromDir(dir http.FileSystem) func(ResponseWriter, *Request, NextMiddlewareFunc) {
	return func(w ResponseWriter, req *Request, next NextMiddlewareFunc) {
		if req.Method != "GET" && req.Method != "HEAD" {
			next(w, req)
			return
		}

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

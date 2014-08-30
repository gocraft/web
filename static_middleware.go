package web

import (
	"net/http"
	"path/filepath"
	"strings"
)

// StaticMiddleware("public") returns proper middleware
// modified from original impl at github.com/codegangsta/martini
func StaticMiddleware(path string) func(ResponseWriter, *Request, NextMiddlewareFunc) {
	dir := http.Dir(path)
	return func(w ResponseWriter, req *Request, next NextMiddlewareFunc) {
		file := req.URL.Path
		var f http.File
		var err error
		// definitely serve index.html (or more like index.htm, index.php?)
		// for "/" or "/my_dir/"
		// @todo "/my_dir" won't be served
		if file[len(file)-1:] == "/" {
			file = filepath.Join(file, "index.html")
		}
		// if client accept gzip
		if strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {
			// search for corresponding gzip file
			f, _ = dir.Open(file + ".gz")
			if f != nil {
				// inform client gzip content encoding
				w.Header().Set("Content-Encoding", "gzip")
			}
		}
		// otherwise, neither client accepts gzip, nor server has it
		if f == nil {
			// fall back to uncompressed content
			f, err = dir.Open(file)
			if err != nil {
				next(w, req)
				return
			}
		}
		defer f.Close()
		fi, err := f.Stat()
		if err != nil {
			next(w, req)
			return
		}
		http.ServeContent(w, req.Request, file, fi.ModTime(), f)
	}
}

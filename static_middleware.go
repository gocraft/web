package web

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gocraft/web"
)

var Root string

// StaticOptions is a struct for specifying configuration options for the martini.Static middleware.
type StaticOptions struct {
	// Prefix is the optional prefix used to serve the static directory content
	Prefix string
	// SkipLogging will disable [Static] log messages when a static file is served.
	SkipLogging bool
	// IndexFile defines which file to serve as index if it exists.
	IndexFile string
	// Expires defines which user-defined function to use for producing a HTTP Expires Header
	// https://developers.google.com/speed/docs/insights/LeverageBrowserCaching
	Expires func() string
}

func init() {
	var err error
	Root, err = os.Getwd()
	if err != nil {
		panic(err)
	}
}

func prepareStaticOptions(options []StaticOptions) StaticOptions {
	var opt StaticOptions
	if len(options) > 0 {
		opt = options[0]
	}

	// Defaults
	if len(opt.IndexFile) == 0 {
		opt.IndexFile = "index.html"
	}
	// Normalize the prefix if provided
	if opt.Prefix != "" {
		// Ensure we have a leading '/'
		if opt.Prefix[0] != '/' {
			opt.Prefix = "/" + opt.Prefix
		}
		// Remove any trailing '/'
		opt.Prefix = strings.TrimRight(opt.Prefix, "/")
	}
	return opt
}

// Static returns a middleware handler that serves static files in the given directory.
func Static(directory string, staticOpt ...StaticOptions) func(web.ResponseWriter, *web.Request, web.NextMiddlewareFunc) {
	if !filepath.IsAbs(directory) {
		directory = filepath.Join(Root, directory)
	}
	dir := http.Dir(directory)
	opt := prepareStaticOptions(staticOpt)

	return func(w web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc) {
		if req.Method != "GET" && req.Method != "HEAD" {
			return
		}
		file := req.URL.Path
		// if we have a prefix, filter requests by stripping the prefix
		if opt.Prefix != "" {
			if !strings.HasPrefix(file, opt.Prefix) {
				next(w, req)
				return
			}
			file = file[len(opt.Prefix):]
			if file != "" && file[0] != '/' {
				next(w, req)
				return
			}
		}
		f, err := dir.Open(file)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		defer f.Close()

		fi, err := f.Stat()
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// try to serve index file
		if fi.IsDir() {
			// redirect if missing trailing slash
			if !strings.HasSuffix(req.URL.Path, "/") {
				http.Redirect(w, req.Request, req.URL.Path+"/", http.StatusFound)
				return
			}

			file = filepath.Join(file, opt.IndexFile)
			f, err = dir.Open(file)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			defer f.Close()

			fi, err = f.Stat()
			if err != nil || fi.IsDir() {
				w.WriteHeader(http.StatusNotFound)
				return
			}
		}

		if !opt.SkipLogging {
			log.Println("[Static] Serving " + file)
		}

		// Add an Expires header to the static content
		if opt.Expires != nil {
			w.Header().Set("Expires", opt.Expires())
		}

		http.ServeContent(w, req.Request, file, fi.ModTime(), f)
	}
}

package middleware

import (
	"fmt"
	"github.com/gocraft/web"
	"time"
)

// NoCache is a simple piece of middleware that sets a number of HTTP headers to prevent
// a route from being cached by an upstream proxy and/or client.
//
// As per http://wiki.nginx.org/HttpProxyModule - NoCache sets:
//		Expires: Thu, 01 Jan 1970 00:00:00 UTC
//      Cache-Control: no-cache, private, max-age=0
//		X-Accel-Expires: 0
func NoCache(rw web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc) {
	rw.Header().Set("Expires", fmt.Sprintf("%s", time.Unix(0, 0).Format(time.RFC1123)))
	rw.Header().Set("Cache-Control", "no-cache, private, max-age=0")
	rw.Header().Set("X-Accel-Expires", "0")
	next(rw, req)
}

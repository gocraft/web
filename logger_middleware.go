package web

import (
  "time"
  "fmt"
)

func LoggerMiddleware(rw *ResponseWriter, req *Request, next NextMiddlewareFunc) {
  startTime := time.Now()

  next()

  duration := time.Since(startTime).Nanoseconds()
  var durationUnits string
  switch {
  case duration > 1000000:
    durationUnits = "ms"
    duration /= 1000000
  case duration > 1000:
    durationUnits = "μs"
    duration /= 1000
  default:
    durationUnits = "ns"
  }

  fmt.Printf("[%d %s] '%s'\n", duration, durationUnits, req.URL.Path)
}

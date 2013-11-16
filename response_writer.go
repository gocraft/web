package web

import (
  "net/http"
)

type ResponseWriter struct {
  http.ResponseWriter
}

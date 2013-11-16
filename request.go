package web

import (
  "net/http"
  "reflect"
)

type Request struct {
  *http.Request
  
  UrlVariables map[string]string
  
  route *Route // The actual route that got invoked
  context reflect.Value // The target context corresponding to the route.
}
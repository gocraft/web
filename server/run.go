package main

import (
	"fmt"
	"github.com/ricardo-rossi/web"
	"net/http"
)

type context struct{}

func main() {
	fmt.Println("Starting server on port 3000...")
	router := web.New(context{})
	router.Get("/", func(rw web.ResponseWriter, req *web.Request) {
		fmt.Fprint(rw, "Hello Meetup!")
	})
	http.ListenAndServe("localhost:3000", router)
}

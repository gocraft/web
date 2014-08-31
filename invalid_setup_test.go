package web

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func (c *Context) InvalidHandler()                                     {}
func (c *Context) InvalidHandler2(w ResponseWriter, r *Request) string { return "" }
func (c *Context) InvalidHandler3(w ResponseWriter, r ResponseWriter)  {}

type invalidSubcontext struct{}

func (c *invalidSubcontext) Handler(w ResponseWriter, r *Request) {}

type invalidSubcontext2 struct {
	*invalidSubcontext
}

func TestInvalidContext(t *testing.T) {
	assert.Panics(t, func() {
		New(1)
	})

	assert.Panics(t, func() {
		router := New(Context{})
		router.Subrouter(invalidSubcontext{}, "")
	})

	assert.Panics(t, func() {
		router := New(Context{})
		router.Subrouter(invalidSubcontext2{}, "")
	})
}

func TestInvalidHandler(t *testing.T) {
	router := New(Context{})

	assert.Panics(t, func() {
		router.Get("/action", 1)
	})

	assert.Panics(t, func() {
		router.Get("/action", (*Context).InvalidHandler)
	})

	// Returns a string:
	assert.Panics(t, func() {
		router.Get("/action", (*Context).InvalidHandler2)
	})

	// Two writer inputs:
	assert.Panics(t, func() {
		router.Get("/action", (*Context).InvalidHandler3)
	})

	// Wrong context type:
	assert.Panics(t, func() {
		router.Get("/action", (*invalidSubcontext).Handler)
	})

	//
}

func TestInvalidMiddleware(t *testing.T) {
	router := New(Context{})

	assert.Panics(t, func() {
		router.Middleware((*Context).InvalidHandler)
	})
}

func TestInvalidNotFound(t *testing.T) {
	router := New(Context{})

	assert.Panics(t, func() {
		router.NotFound((*Context).InvalidHandler)
	})

	// Valid handler not on main router:
	subrouter := router.Subrouter(Context{}, "")
	assert.Panics(t, func() {
		subrouter.NotFound((*Context).A)
	})
}

func TestInvalidError(t *testing.T) {
	router := New(Context{})

	assert.Panics(t, func() {
		router.Error((*Context).InvalidHandler)
	})
}

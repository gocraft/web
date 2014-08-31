package web

import (
	"fmt"
	"testing"
)

func (c *Context) A(w ResponseWriter, r *Request) {
	fmt.Fprintf(w, "context-A")
}

func (c *Context) Z(w ResponseWriter, r *Request) {
	fmt.Fprintf(w, "context-Z")
}

func (c *AdminContext) B(w ResponseWriter, r *Request) {
	fmt.Fprintf(w, "admin-B")
}

func (c *APIContext) C(w ResponseWriter, r *Request) {
	fmt.Fprintf(w, "api-C")
}

func (c *TicketsContext) D(w ResponseWriter, r *Request) {
	fmt.Fprintf(w, "tickets-D")
}

func (c *Context) mwNoNext(w ResponseWriter, r *Request, next NextMiddlewareFunc) {
	fmt.Fprintf(w, "context-mw-NoNext ")
}

func (c *Context) mwAlpha(w ResponseWriter, r *Request, next NextMiddlewareFunc) {
	fmt.Fprintf(w, "context-mw-Alpha ")
	next(w, r)
}

func (c *Context) mwBeta(w ResponseWriter, r *Request, next NextMiddlewareFunc) {
	fmt.Fprintf(w, "context-mw-Beta ")
	next(w, r)
}

func (c *Context) mwGamma(w ResponseWriter, r *Request, next NextMiddlewareFunc) {
	fmt.Fprintf(w, "context-mw-Gamma ")
	next(w, r)
}

func (c *APIContext) mwDelta(w ResponseWriter, r *Request, next NextMiddlewareFunc) {
	fmt.Fprintf(w, "api-mw-Delta ")
	next(w, r)
}

func (c *AdminContext) mwEpsilon(w ResponseWriter, r *Request, next NextMiddlewareFunc) {
	fmt.Fprintf(w, "admin-mw-Epsilon ")
	next(w, r)
}

func (c *AdminContext) mwZeta(w ResponseWriter, r *Request, next NextMiddlewareFunc) {
	fmt.Fprintf(w, "admin-mw-Zeta ")
	next(w, r)
}

func (c *TicketsContext) mwEta(w ResponseWriter, r *Request, next NextMiddlewareFunc) {
	fmt.Fprintf(w, "tickets-mw-Eta ")
	next(w, r)
}

func mwGenricInterface(ctx interface{}, w ResponseWriter, r *Request, next NextMiddlewareFunc) {
	fmt.Fprintf(w, "context-mw-Interface ")
	next(w, r)
}

func TestFlatNoMiddleware(t *testing.T) {
	router := New(Context{})
	router.Get("/action", (*Context).A)
	router.Get("/action_z", (*Context).Z)

	rw, req := newTestRequest("GET", "/action")
	router.ServeHTTP(rw, req)
	assertResponse(t, rw, "context-A", 200)

	rw, req = newTestRequest("GET", "/action_z")
	router.ServeHTTP(rw, req)
	assertResponse(t, rw, "context-Z", 200)
}

func TestFlatOneMiddleware(t *testing.T) {
	router := New(Context{})
	router.Middleware((*Context).mwAlpha)
	router.Get("/action", (*Context).A)
	router.Get("/action_z", (*Context).Z)

	rw, req := newTestRequest("GET", "/action")
	router.ServeHTTP(rw, req)
	assertResponse(t, rw, "context-mw-Alpha context-A", 200)

	rw, req = newTestRequest("GET", "/action_z")
	router.ServeHTTP(rw, req)
	assertResponse(t, rw, "context-mw-Alpha context-Z", 200)
}

func TestFlatTwoMiddleware(t *testing.T) {
	router := New(Context{})
	router.Middleware((*Context).mwAlpha)
	router.Middleware((*Context).mwBeta)
	router.Get("/action", (*Context).A)
	router.Get("/action_z", (*Context).Z)

	rw, req := newTestRequest("GET", "/action")
	router.ServeHTTP(rw, req)
	assertResponse(t, rw, "context-mw-Alpha context-mw-Beta context-A", 200)

	rw, req = newTestRequest("GET", "/action_z")
	router.ServeHTTP(rw, req)
	assertResponse(t, rw, "context-mw-Alpha context-mw-Beta context-Z", 200)
}

func TestDualTree(t *testing.T) {
	router := New(Context{})
	router.Middleware((*Context).mwAlpha)
	router.Get("/action", (*Context).A)
	admin := router.Subrouter(AdminContext{}, "/admin")
	admin.Middleware((*AdminContext).mwEpsilon)
	admin.Get("/action", (*AdminContext).B)
	api := router.Subrouter(APIContext{}, "/api")
	api.Middleware((*APIContext).mwDelta)
	api.Get("/action", (*APIContext).C)

	rw, req := newTestRequest("GET", "/action")
	router.ServeHTTP(rw, req)
	assertResponse(t, rw, "context-mw-Alpha context-A", 200)

	rw, req = newTestRequest("GET", "/admin/action")
	router.ServeHTTP(rw, req)
	assertResponse(t, rw, "context-mw-Alpha admin-mw-Epsilon admin-B", 200)

	rw, req = newTestRequest("GET", "/api/action")
	router.ServeHTTP(rw, req)
	assertResponse(t, rw, "context-mw-Alpha api-mw-Delta api-C", 200)
}

func TestDualLeaningLeftTree(t *testing.T) {
	router := New(Context{})
	router.Get("/action", (*Context).A)
	admin := router.Subrouter(AdminContext{}, "/admin")
	admin.Get("/action", (*AdminContext).B)
	api := router.Subrouter(APIContext{}, "/api")
	api.Middleware((*APIContext).mwDelta)
	api.Get("/action", (*APIContext).C)

	rw, req := newTestRequest("GET", "/action")
	router.ServeHTTP(rw, req)
	assertResponse(t, rw, "context-A", 200)

	rw, req = newTestRequest("GET", "/admin/action")
	router.ServeHTTP(rw, req)
	assertResponse(t, rw, "admin-B", 200)

	rw, req = newTestRequest("GET", "/api/action")
	router.ServeHTTP(rw, req)
	assertResponse(t, rw, "api-mw-Delta api-C", 200)
}

func TestTicketsA(t *testing.T) {
	router := New(Context{})
	admin := router.Subrouter(AdminContext{}, "/admin")
	admin.Middleware((*AdminContext).mwEpsilon)
	tickets := admin.Subrouter(TicketsContext{}, "/tickets")
	tickets.Get("/action", (*TicketsContext).D)

	rw, req := newTestRequest("GET", "/admin/tickets/action")
	router.ServeHTTP(rw, req)
	assertResponse(t, rw, "admin-mw-Epsilon tickets-D", 200)
}

func TestTicketsB(t *testing.T) {
	router := New(Context{})
	admin := router.Subrouter(AdminContext{}, "/admin")
	tickets := admin.Subrouter(TicketsContext{}, "/tickets")
	tickets.Middleware((*TicketsContext).mwEta)
	tickets.Get("/action", (*TicketsContext).D)

	rw, req := newTestRequest("GET", "/admin/tickets/action")
	router.ServeHTTP(rw, req)
	assertResponse(t, rw, "tickets-mw-Eta tickets-D", 200)
}

func TestTicketsC(t *testing.T) {
	router := New(Context{})
	router.Middleware((*Context).mwAlpha)
	admin := router.Subrouter(AdminContext{}, "/admin")
	tickets := admin.Subrouter(TicketsContext{}, "/tickets")
	tickets.Get("/action", (*TicketsContext).D)

	rw, req := newTestRequest("GET", "/admin/tickets/action")
	router.ServeHTTP(rw, req)
	assertResponse(t, rw, "context-mw-Alpha tickets-D", 200)
}

func TestTicketsD(t *testing.T) {
	router := New(Context{})
	router.Middleware((*Context).mwAlpha)
	admin := router.Subrouter(AdminContext{}, "/admin")
	tickets := admin.Subrouter(TicketsContext{}, "/tickets")
	tickets.Middleware((*TicketsContext).mwEta)
	tickets.Get("/action", (*TicketsContext).D)

	rw, req := newTestRequest("GET", "/admin/tickets/action")
	router.ServeHTTP(rw, req)
	assertResponse(t, rw, "context-mw-Alpha tickets-mw-Eta tickets-D", 200)
}

func TestTicketsE(t *testing.T) {
	router := New(Context{})
	router.Middleware((*Context).mwAlpha)
	router.Middleware((*Context).mwBeta)
	router.Middleware((*Context).mwGamma)
	admin := router.Subrouter(AdminContext{}, "/admin")
	admin.Middleware((*AdminContext).mwEpsilon)
	admin.Middleware((*AdminContext).mwZeta)
	tickets := admin.Subrouter(TicketsContext{}, "/tickets")
	tickets.Middleware((*TicketsContext).mwEta)
	tickets.Get("/action", (*TicketsContext).D)

	rw, req := newTestRequest("GET", "/admin/tickets/action")
	router.ServeHTTP(rw, req)
	assertResponse(t, rw, "context-mw-Alpha context-mw-Beta context-mw-Gamma admin-mw-Epsilon admin-mw-Zeta tickets-mw-Eta tickets-D", 200)
}

func TestNoNext(t *testing.T) {
	router := New(Context{})
	router.Middleware((*Context).mwNoNext)
	router.Get("/action", (*Context).A)

	rw, req := newTestRequest("GET", "/action")
	router.ServeHTTP(rw, req)
	assertResponse(t, rw, "context-mw-NoNext", 200)
}

func TestSameContext(t *testing.T) {
	router := New(Context{})
	router.Middleware((*Context).mwAlpha).
		Middleware((*Context).mwBeta)
	admin := router.Subrouter(Context{}, "/admin")
	admin.Middleware((*Context).mwGamma)
	admin.Get("/foo", (*Context).A)

	rw, req := newTestRequest("GET", "/admin/foo")
	router.ServeHTTP(rw, req)
	assertResponse(t, rw, "context-mw-Alpha context-mw-Beta context-mw-Gamma context-A", 200)
}

func TestSameNamespace(t *testing.T) {
	router := New(Context{})
	admin := router.Subrouter(AdminContext{}, "/")
	admin.Get("/action", (*AdminContext).B)

	rw, req := newTestRequest("GET", "/action")
	router.ServeHTTP(rw, req)
	assertResponse(t, rw, "admin-B", 200)
}

func TestInterfaceMiddleware(t *testing.T) {
	router := New(Context{})
	router.Middleware(mwGenricInterface)
	router.Get("/action", (*Context).A)

	rw, req := newTestRequest("GET", "/action")
	router.ServeHTTP(rw, req)
	assertResponse(t, rw, "context-mw-Interface context-A", 200)
}

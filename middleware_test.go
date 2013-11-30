package web_test

import (
	"fmt"
	"github.com/gocraft/web"
	. "launchpad.net/gocheck"
	// "net/http"
	// "strings"
)

type MiddlewareTestSuite struct{}

var _ = Suite(&MiddlewareTestSuite{})

func (c *Context) A(w web.ResponseWriter, r *web.Request) {
	fmt.Fprintf(w, "context-A")
}

func (c *Context) Z(w web.ResponseWriter, r *web.Request) {
	fmt.Fprintf(w, "context-Z")
}

func (c *AdminContext) B(w web.ResponseWriter, r *web.Request) {
	fmt.Fprintf(w, "admin-B")
}

func (c *ApiContext) C(w web.ResponseWriter, r *web.Request) {
	fmt.Fprintf(w, "api-C")
}

func (c *TicketsContext) D(w web.ResponseWriter, r *web.Request) {
	fmt.Fprintf(w, "tickets-D")
}

func (c *Context) mwNoNext(w web.ResponseWriter, r *web.Request, next web.NextMiddlewareFunc) {
	fmt.Fprintf(w, "context-mw-NoNext ")
}

func (c *Context) mwAlpha(w web.ResponseWriter, r *web.Request, next web.NextMiddlewareFunc) {
	fmt.Fprintf(w, "context-mw-Alpha ")
	next(w, r)
}

func (c *Context) mwBeta(w web.ResponseWriter, r *web.Request, next web.NextMiddlewareFunc) {
	fmt.Fprintf(w, "context-mw-Beta ")
	next(w, r)
}

func (c *Context) mwGamma(w web.ResponseWriter, r *web.Request, next web.NextMiddlewareFunc) {
	fmt.Fprintf(w, "context-mw-Gamma ")
	next(w, r)
}

func (c *ApiContext) mwDelta(w web.ResponseWriter, r *web.Request, next web.NextMiddlewareFunc) {
	fmt.Fprintf(w, "api-mw-Delta ")
	next(w, r)
}

func (c *AdminContext) mwEpsilon(w web.ResponseWriter, r *web.Request, next web.NextMiddlewareFunc) {
	fmt.Fprintf(w, "admin-mw-Epsilon ")
	next(w, r)
}

func (c *AdminContext) mwZeta(w web.ResponseWriter, r *web.Request, next web.NextMiddlewareFunc) {
	fmt.Fprintf(w, "admin-mw-Zeta ")
	next(w, r)
}

func (c *TicketsContext) mwEta(w web.ResponseWriter, r *web.Request, next web.NextMiddlewareFunc) {
	fmt.Fprintf(w, "tickets-mw-Eta ")
	next(w, r)
}

func (s *MiddlewareTestSuite) TestFlatNoMiddleware(c *C) {
	router := web.New(Context{})
	router.Get("/action", (*Context).A)
	router.Get("/action_z", (*Context).Z)

	rw, req := newTestRequest("GET", "/action")
	router.ServeHTTP(rw, req)
	assertResponse(c, rw, "context-A", 200)

	rw, req = newTestRequest("GET", "/action_z")
	router.ServeHTTP(rw, req)
	assertResponse(c, rw, "context-Z", 200)
}

func (s *MiddlewareTestSuite) TestFlatOneMiddleware(c *C) {
	router := web.New(Context{})
	router.Middleware((*Context).mwAlpha)
	router.Get("/action", (*Context).A)
	router.Get("/action_z", (*Context).Z)

	rw, req := newTestRequest("GET", "/action")
	router.ServeHTTP(rw, req)
	assertResponse(c, rw, "context-mw-Alpha context-A", 200)

	rw, req = newTestRequest("GET", "/action_z")
	router.ServeHTTP(rw, req)
	assertResponse(c, rw, "context-mw-Alpha context-Z", 200)
}

func (s *MiddlewareTestSuite) TestFlatTwoMiddleware(c *C) {
	router := web.New(Context{})
	router.Middleware((*Context).mwAlpha)
	router.Middleware((*Context).mwBeta)
	router.Get("/action", (*Context).A)
	router.Get("/action_z", (*Context).Z)

	rw, req := newTestRequest("GET", "/action")
	router.ServeHTTP(rw, req)
	assertResponse(c, rw, "context-mw-Alpha context-mw-Beta context-A", 200)

	rw, req = newTestRequest("GET", "/action_z")
	router.ServeHTTP(rw, req)
	assertResponse(c, rw, "context-mw-Alpha context-mw-Beta context-Z", 200)
}

func (s *MiddlewareTestSuite) TestDualTree(c *C) {
	router := web.New(Context{})
	router.Middleware((*Context).mwAlpha)
	router.Get("/action", (*Context).A)
	admin := router.Subrouter(AdminContext{}, "/admin")
	admin.Middleware((*AdminContext).mwEpsilon)
	admin.Get("/action", (*AdminContext).B)
	api := router.Subrouter(ApiContext{}, "/api")
	api.Middleware((*ApiContext).mwDelta)
	api.Get("/action", (*ApiContext).C)

	rw, req := newTestRequest("GET", "/action")
	router.ServeHTTP(rw, req)
	assertResponse(c, rw, "context-mw-Alpha context-A", 200)

	rw, req = newTestRequest("GET", "/admin/action")
	router.ServeHTTP(rw, req)
	assertResponse(c, rw, "context-mw-Alpha admin-mw-Epsilon admin-B", 200)

	rw, req = newTestRequest("GET", "/api/action")
	router.ServeHTTP(rw, req)
	assertResponse(c, rw, "context-mw-Alpha api-mw-Delta api-C", 200)
}

func (s *MiddlewareTestSuite) TestDualLeaningLeftTree(c *C) {
	router := web.New(Context{})
	router.Get("/action", (*Context).A)
	admin := router.Subrouter(AdminContext{}, "/admin")
	admin.Get("/action", (*AdminContext).B)
	api := router.Subrouter(ApiContext{}, "/api")
	api.Middleware((*ApiContext).mwDelta)
	api.Get("/action", (*ApiContext).C)

	rw, req := newTestRequest("GET", "/action")
	router.ServeHTTP(rw, req)
	assertResponse(c, rw, "context-A", 200)

	rw, req = newTestRequest("GET", "/admin/action")
	router.ServeHTTP(rw, req)
	assertResponse(c, rw, "admin-B", 200)

	rw, req = newTestRequest("GET", "/api/action")
	router.ServeHTTP(rw, req)
	assertResponse(c, rw, "api-mw-Delta api-C", 200)
}

func (s *MiddlewareTestSuite) TestTicketsA(c *C) {
	router := web.New(Context{})
	admin := router.Subrouter(AdminContext{}, "/admin")
	admin.Middleware((*AdminContext).mwEpsilon)
	tickets := admin.Subrouter(TicketsContext{}, "/tickets")
	tickets.Get("/action", (*TicketsContext).D)

	rw, req := newTestRequest("GET", "/admin/tickets/action")
	router.ServeHTTP(rw, req)
	assertResponse(c, rw, "admin-mw-Epsilon tickets-D", 200)
}

func (s *MiddlewareTestSuite) TestTicketsB(c *C) {
	router := web.New(Context{})
	admin := router.Subrouter(AdminContext{}, "/admin")
	tickets := admin.Subrouter(TicketsContext{}, "/tickets")
	tickets.Middleware((*TicketsContext).mwEta)
	tickets.Get("/action", (*TicketsContext).D)

	rw, req := newTestRequest("GET", "/admin/tickets/action")
	router.ServeHTTP(rw, req)
	assertResponse(c, rw, "tickets-mw-Eta tickets-D", 200)
}

func (s *MiddlewareTestSuite) TestTicketsC(c *C) {
	router := web.New(Context{})
	router.Middleware((*Context).mwAlpha)
	admin := router.Subrouter(AdminContext{}, "/admin")
	tickets := admin.Subrouter(TicketsContext{}, "/tickets")
	tickets.Get("/action", (*TicketsContext).D)

	rw, req := newTestRequest("GET", "/admin/tickets/action")
	router.ServeHTTP(rw, req)
	assertResponse(c, rw, "context-mw-Alpha tickets-D", 200)
}

func (s *MiddlewareTestSuite) TestTicketsD(c *C) {
	router := web.New(Context{})
	router.Middleware((*Context).mwAlpha)
	admin := router.Subrouter(AdminContext{}, "/admin")
	tickets := admin.Subrouter(TicketsContext{}, "/tickets")
	tickets.Middleware((*TicketsContext).mwEta)
	tickets.Get("/action", (*TicketsContext).D)

	rw, req := newTestRequest("GET", "/admin/tickets/action")
	router.ServeHTTP(rw, req)
	assertResponse(c, rw, "context-mw-Alpha tickets-mw-Eta tickets-D", 200)
}

func (s *MiddlewareTestSuite) TestTicketsE(c *C) {
	router := web.New(Context{})
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
	assertResponse(c, rw, "context-mw-Alpha context-mw-Beta context-mw-Gamma admin-mw-Epsilon admin-mw-Zeta tickets-mw-Eta tickets-D", 200)
}

func (s *MiddlewareTestSuite) TestNoNext(c *C) {
	router := web.New(Context{})
	router.Middleware((*Context).mwNoNext)
	router.Get("/action", (*Context).A)

	rw, req := newTestRequest("GET", "/action")
	router.ServeHTTP(rw, req)
	assertResponse(c, rw, "context-mw-NoNext", 200)
}

func (s *MiddlewareTestSuite) TestSameContext(c *C) {
	router := web.New(Context{})
	router.Middleware((*Context).mwAlpha).
		Middleware((*Context).mwBeta)
	admin := router.Subrouter(Context{}, "/admin")
	admin.Middleware((*Context).mwGamma)
	admin.Get("/foo", (*Context).A)

	rw, req := newTestRequest("GET", "/admin/foo")
	router.ServeHTTP(rw, req)
	assertResponse(c, rw, "context-mw-Alpha context-mw-Beta context-mw-Gamma context-A", 200)
}

func (s *MiddlewareTestSuite) TestSameNamespace(c *C) {
	router := web.New(Context{})
	admin := router.Subrouter(AdminContext{}, "/")
	admin.Get("/action", (*AdminContext).B)

	rw, req := newTestRequest("GET", "/action")
	router.ServeHTTP(rw, req)
	assertResponse(c, rw, "admin-B", 200)
}

package http

import (
	"chatx-01-backend/internal/auth/usecase/authuc"
	"chatx-01-backend/internal/auth/usecase/useruc"
	"chatx-01-backend/internal/portal/auth"
	"chatx-01-backend/pkg/token"
	"net/http"
)

type ctrl struct {
	mux    *http.ServeMux
	prefix string

	authUsecase authuc.UseCase
	userUsecase useruc.UseCase

	authPr auth.Portal
}

func Register(
	mux *http.ServeMux,
	prefix string,
	authUsecase authuc.UseCase,
	userUsecase useruc.UseCase,
	tokenGenerator token.Generator,
	authPr auth.Portal,
) {
	c := &ctrl{
		mux:         mux,
		prefix:      prefix,
		authUsecase: authUsecase,
		userUsecase: userUsecase,
		authPr:      authPr,
	}

	c.registerHandlers()
}

// registerHandlers registers all handlers.
func (c *ctrl) registerHandlers() {
	// auth endpoints
	c.register(http.MethodPost, "/auth/login", http.HandlerFunc(c.login))
	c.register(http.MethodPost, "/auth/logout", http.HandlerFunc(c.logout), c.authPr.RequireAuth())

	// user endpoints
	c.register(http.MethodPost, "/users", http.HandlerFunc(c.createUser), c.authPr.RequireAdmin())
	c.register(http.MethodGet, "/users", http.HandlerFunc(c.getUsersList), c.authPr.RequireAdmin())
	c.register(http.MethodGet, "/users/{user_id}", http.HandlerFunc(c.getUser), c.authPr.RequireAdmin())
	c.register(http.MethodDelete, "/users/{user_id}", http.HandlerFunc(c.deleteUser), c.authPr.RequireAdmin())
	c.register(http.MethodGet, "/users/me", http.HandlerFunc(c.getMe), c.authPr.RequireAuth())
	c.register(http.MethodPut, "/users/me/password", http.HandlerFunc(c.changePassword), c.authPr.RequireAuth())
	c.register(http.MethodPut, "/users/me/image", http.HandlerFunc(c.changeImage), c.authPr.RequireAuth())
}

func (c *ctrl) register(
	method string,
	path string,
	handler http.Handler,
	middlewares ...func(http.Handler) http.Handler,
) {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}

	fullPath := c.prefix + path
	c.mux.Handle(method+" "+fullPath, handler)
}

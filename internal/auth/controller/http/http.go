package http

import (
	"chatx-01/internal/auth/usecase/authuc"
	"chatx-01/internal/auth/usecase/useruc"
	"net/http"
)

type ctrl struct {
	mux    *http.ServeMux
	prefix string

	authUsecase authuc.UseCase
	userUsecase useruc.UseCase
}

func New(mux *http.ServeMux, prefix string, authUsecase authuc.UseCase, userUsecase useruc.UseCase) *ctrl {
	c := &ctrl{
		mux:         mux,
		prefix:      prefix,
		authUsecase: authUsecase,
		userUsecase: userUsecase,
	}

	c.registerHandlers()
	return c
}

// registerHandlers registers all handlers.
func (c *ctrl) registerHandlers() {
	// Auth endpoints
	c.register(http.MethodPost, "/auth/login", http.HandlerFunc(c.login))
	c.register(http.MethodPost, "/auth/logout", http.HandlerFunc(c.logout))

	// User endpoints
	c.register(http.MethodGet, "/users/me", http.HandlerFunc(c.getMe))
	c.register(http.MethodPut, "/users/me/password", http.HandlerFunc(c.changePassword))
	c.register(http.MethodPut, "/users/me/image", http.HandlerFunc(c.changeImage))
	c.register(http.MethodPost, "/users", http.HandlerFunc(c.createUser))
	c.register(http.MethodDelete, "/users/{userId}", http.HandlerFunc(c.deleteUser))
	c.register(http.MethodGet, "/users/{userId}", http.HandlerFunc(c.getUser))
	c.register(http.MethodGet, "/users", http.HandlerFunc(c.getUsersList))
}

func (c *ctrl) register(method string, path string, handler http.Handler) {
	fullPath := c.prefix + path
	c.mux.Handle(method+" "+fullPath, handler)
}

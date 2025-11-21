package http

import (
	"chatx-01-backend/internal/auth/usecase/authuc"
	"chatx-01-backend/pkg/httptools"
	"net/http"
	"strings"
)

func (c *ctrl) login(w http.ResponseWriter, r *http.Request) {
	req, err := httptools.BindRequest[authuc.LoginReq](r)
	if err != nil {
		httptools.HandleError(w, err)
		return
	}

	resp, err := c.authUsecase.Login(r.Context(), req)
	if err != nil {
		httptools.HandleError(w, err)
		return
	}

	httptools.WriteResponse(http.StatusOK, w, resp)
}

func (c *ctrl) logout(w http.ResponseWriter, r *http.Request) {
	req, err := httptools.BindRequest[authuc.LogoutReq](r)
	if err != nil {
		httptools.HandleError(w, err)
		return
	}

	// Extract token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			req.AccessToken = parts[1]
		}
	}

	err = c.authUsecase.Logout(r.Context(), req)
	if err != nil {
		httptools.HandleError(w, err)
		return
	}

	httptools.WriteResponse(http.StatusNoContent, w, nil)
}

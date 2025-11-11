package http

import (
	"chatx-01/internal/auth/usecase/authuc"
	"chatx-01/pkg/httpx"
	"net/http"
)

func (c *ctrl) login(w http.ResponseWriter, r *http.Request) {
	req, err := httpx.BindRequest[authuc.LoginReq](r)
	if err != nil {
		httpx.HandleError(w, err)
		return
	}

	resp, err := c.authUsecase.Login(r.Context(), req)
	if err != nil {
		httpx.HandleError(w, err)
		return
	}

	httpx.WriteResponse(http.StatusOK, w, resp)
}

func (c *ctrl) logout(w http.ResponseWriter, r *http.Request) {
	req, err := httpx.BindRequest[authuc.LogoutReq](r)
	if err != nil {
		httpx.HandleError(w, err)
		return
	}

	err = c.authUsecase.Logout(r.Context(), req)
	if err != nil {
		httpx.HandleError(w, err)
		return
	}

	httpx.WriteResponse(http.StatusNoContent, w, nil)
}

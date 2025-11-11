package http

import (
	"chatx-01/internal/auth/usecase/useruc"
	"chatx-01/pkg/httpx"
	"net/http"
)

func (c *ctrl) getMe(w http.ResponseWriter, r *http.Request) {
	req, err := httpx.BindRequest[useruc.GetMeReq](r)
	if err != nil {
		httpx.HandleError(w, err)
		return
	}

	resp, err := c.userUsecase.GetMe(r.Context(), req)
	if err != nil {
		httpx.HandleError(w, err)
		return
	}

	httpx.WriteResponse(http.StatusOK, w, resp)
}

func (c *ctrl) changePassword(w http.ResponseWriter, r *http.Request) {
	req, err := httpx.BindRequest[useruc.ChangePasswordReq](r)
	if err != nil {
		httpx.HandleError(w, err)
		return
	}

	err = c.userUsecase.ChangePassword(r.Context(), req)
	if err != nil {
		httpx.HandleError(w, err)
		return
	}

	httpx.WriteResponse(http.StatusNoContent, w, nil)
}

func (c *ctrl) changeImage(w http.ResponseWriter, r *http.Request) {
	req, err := httpx.BindRequest[useruc.ChangeImageReq](r)
	if err != nil {
		httpx.HandleError(w, err)
		return
	}

	resp, err := c.userUsecase.ChangeImage(r.Context(), req)
	if err != nil {
		httpx.HandleError(w, err)
		return
	}

	httpx.WriteResponse(http.StatusOK, w, resp)
}

func (c *ctrl) createUser(w http.ResponseWriter, r *http.Request) {
	req, err := httpx.BindRequest[useruc.CreateUserReq](r)
	if err != nil {
		httpx.HandleError(w, err)
		return
	}

	resp, err := c.userUsecase.CreateUser(r.Context(), req)
	if err != nil {
		httpx.HandleError(w, err)
		return
	}

	httpx.WriteResponse(http.StatusCreated, w, resp)
}

func (c *ctrl) deleteUser(w http.ResponseWriter, r *http.Request) {
	req, err := httpx.BindRequest[useruc.DeleteUserReq](r)
	if err != nil {
		httpx.HandleError(w, err)
		return
	}

	err = c.userUsecase.DeleteUser(r.Context(), req)
	if err != nil {
		httpx.HandleError(w, err)
		return
	}

	httpx.WriteResponse(http.StatusNoContent, w, nil)
}

func (c *ctrl) getUser(w http.ResponseWriter, r *http.Request) {
	req, err := httpx.BindRequest[useruc.GetUserReq](r)
	if err != nil {
		httpx.HandleError(w, err)
		return
	}

	resp, err := c.userUsecase.GetUser(r.Context(), req)
	if err != nil {
		httpx.HandleError(w, err)
		return
	}

	httpx.WriteResponse(http.StatusOK, w, resp)
}

func (c *ctrl) getUsersList(w http.ResponseWriter, r *http.Request) {
	req, err := httpx.BindRequest[useruc.GetUsersListReq](r)
	if err != nil {
		httpx.HandleError(w, err)
		return
	}

	resp, err := c.userUsecase.GetUsersList(r.Context(), req)
	if err != nil {
		httpx.HandleError(w, err)
		return
	}

	httpx.WriteResponse(http.StatusOK, w, resp)
}

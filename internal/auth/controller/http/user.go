package http

import (
	"chatx-01-backend/internal/auth/usecase/useruc"
	"chatx-01-backend/pkg/httptools"
	"net/http"
)

func (c *ctrl) getMe(w http.ResponseWriter, r *http.Request) {
	req, err := httptools.BindRequest[useruc.GetMeReq](r)
	if err != nil {
		httptools.HandleError(w, err)
		return
	}

	resp, err := c.userUsecase.GetMe(r.Context(), req)
	if err != nil {
		httptools.HandleError(w, err)
		return
	}

	httptools.WriteResponse(http.StatusOK, w, resp)
}

func (c *ctrl) changePassword(w http.ResponseWriter, r *http.Request) {
	req, err := httptools.BindRequest[useruc.ChangePasswordReq](r)
	if err != nil {
		httptools.HandleError(w, err)
		return
	}

	err = c.userUsecase.ChangePassword(r.Context(), req)
	if err != nil {
		httptools.HandleError(w, err)
		return
	}

	httptools.WriteResponse(http.StatusNoContent, w, nil)
}

func (c *ctrl) changeImage(w http.ResponseWriter, r *http.Request) {
	req, err := httptools.BindRequest[useruc.ChangeImageReq](r)
	if err != nil {
		httptools.HandleError(w, err)
		return
	}

	resp, err := c.userUsecase.ChangeImage(r.Context(), req)
	if err != nil {
		httptools.HandleError(w, err)
		return
	}

	httptools.WriteResponse(http.StatusOK, w, resp)
}

func (c *ctrl) createUser(w http.ResponseWriter, r *http.Request) {
	req, err := httptools.BindRequest[useruc.CreateUserReq](r)
	if err != nil {
		httptools.HandleError(w, err)
		return
	}

	resp, err := c.userUsecase.CreateUser(r.Context(), req)
	if err != nil {
		httptools.HandleError(w, err)
		return
	}

	httptools.WriteResponse(http.StatusCreated, w, resp)
}

func (c *ctrl) deleteUser(w http.ResponseWriter, r *http.Request) {
	req, err := httptools.BindRequest[useruc.DeleteUserReq](r)
	if err != nil {
		httptools.HandleError(w, err)
		return
	}

	err = c.userUsecase.DeleteUser(r.Context(), req)
	if err != nil {
		httptools.HandleError(w, err)
		return
	}

	httptools.WriteResponse(http.StatusNoContent, w, nil)
}

func (c *ctrl) getUser(w http.ResponseWriter, r *http.Request) {
	req, err := httptools.BindRequest[useruc.GetUserReq](r)
	if err != nil {
		httptools.HandleError(w, err)
		return
	}

	resp, err := c.userUsecase.GetUser(r.Context(), req)
	if err != nil {
		httptools.HandleError(w, err)
		return
	}

	httptools.WriteResponse(http.StatusOK, w, resp)
}

func (c *ctrl) getUsersList(w http.ResponseWriter, r *http.Request) {
	req, err := httptools.BindRequest[useruc.GetUsersListReq](r)
	if err != nil {
		httptools.HandleError(w, err)
		return
	}

	resp, err := c.userUsecase.GetUsersList(r.Context(), req)
	if err != nil {
		httptools.HandleError(w, err)
		return
	}

	httptools.WriteResponse(http.StatusOK, w, resp)
}

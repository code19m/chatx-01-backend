package http

import (
	"chatx-01-backend/internal/auth/usecase/useruc"
	"chatx-01-backend/pkg/httptools"
	"io"
	"net/http"
	"strings"
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

func (c *ctrl) uploadImage(w http.ResponseWriter, r *http.Request) {
	const maxFileSize = 10 << 20 // 10 MB

	r.Body = http.MaxBytesReader(w, r.Body, maxFileSize)
	if err := r.ParseMultipartForm(maxFileSize); err != nil {
		httptools.HandleError(w, err)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		httptools.HandleError(w, err)
		return
	}
	defer file.Close()

	fileData, err := io.ReadAll(file)
	if err != nil {
		httptools.HandleError(w, err)
		return
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	req := useruc.UploadImageReq{
		File:        fileData,
		FileName:    header.Filename,
		ContentType: contentType,
		Size:        int64(len(fileData)),
	}

	if err := req.Validate(); err != nil {
		httptools.HandleError(w, err)
		return
	}

	resp, err := c.userUsecase.UploadImage(r.Context(), req)
	if err != nil {
		httptools.HandleError(w, err)
		return
	}

	httptools.WriteResponse(http.StatusOK, w, resp)
}

func (c *ctrl) downloadImage(w http.ResponseWriter, r *http.Request) {
	imagePath := strings.TrimPrefix(r.PathValue("image_path"), "/")

	req := useruc.DownloadImageReq{
		ImagePath: imagePath,
	}

	if err := req.Validate(); err != nil {
		httptools.HandleError(w, err)
		return
	}

	resp, err := c.userUsecase.DownloadImage(r.Context(), req)
	if err != nil {
		httptools.HandleError(w, err)
		return
	}

	w.Header().Set("Content-Type", resp.ContentType)
	w.Header().Set("Content-Disposition", "inline; filename=\""+resp.FileName+"\"")
	w.WriteHeader(http.StatusOK)
	w.Write(resp.File)
}

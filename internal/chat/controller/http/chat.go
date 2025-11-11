package http

import (
	"chatx-01/internal/chat/usecase/chatuc"
	"chatx-01/pkg/httpx"
	"net/http"
)

func (c *ctrl) getDMsList(w http.ResponseWriter, r *http.Request) {
	req, err := httpx.BindRequest[chatuc.GetDMsListReq](r)
	if err != nil {
		httpx.HandleError(w, err)
		return
	}

	resp, err := c.chatUsecase.GetDMsList(r.Context(), req)
	if err != nil {
		httpx.HandleError(w, err)
		return
	}

	httpx.WriteResponse(http.StatusOK, w, resp)
}

func (c *ctrl) getGroupsList(w http.ResponseWriter, r *http.Request) {
	req, err := httpx.BindRequest[chatuc.GetGroupsListReq](r)
	if err != nil {
		httpx.HandleError(w, err)
		return
	}

	resp, err := c.chatUsecase.GetGroupsList(r.Context(), req)
	if err != nil {
		httpx.HandleError(w, err)
		return
	}

	httpx.WriteResponse(http.StatusOK, w, resp)
}

func (c *ctrl) getChat(w http.ResponseWriter, r *http.Request) {
	req, err := httpx.BindRequest[chatuc.GetChatReq](r)
	if err != nil {
		httpx.HandleError(w, err)
		return
	}

	resp, err := c.chatUsecase.GetChat(r.Context(), req)
	if err != nil {
		httpx.HandleError(w, err)
		return
	}

	httpx.WriteResponse(http.StatusOK, w, resp)
}

func (c *ctrl) createDM(w http.ResponseWriter, r *http.Request) {
	req, err := httpx.BindRequest[chatuc.CreateDMReq](r)
	if err != nil {
		httpx.HandleError(w, err)
		return
	}

	resp, err := c.chatUsecase.CreateDM(r.Context(), req)
	if err != nil {
		httpx.HandleError(w, err)
		return
	}

	httpx.WriteResponse(http.StatusCreated, w, resp)
}

func (c *ctrl) createGroup(w http.ResponseWriter, r *http.Request) {
	req, err := httpx.BindRequest[chatuc.CreateGroupReq](r)
	if err != nil {
		httpx.HandleError(w, err)
		return
	}

	resp, err := c.chatUsecase.CreateGroup(r.Context(), req)
	if err != nil {
		httpx.HandleError(w, err)
		return
	}

	httpx.WriteResponse(http.StatusCreated, w, resp)
}

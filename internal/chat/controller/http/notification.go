package http

import (
	"chatx-01/internal/chat/usecase/notificationuc"
	"chatx-01/pkg/httpx"
	"net/http"
)

func (c *ctrl) getUnreadMessagesCount(w http.ResponseWriter, r *http.Request) {
	req, err := httpx.BindRequest[notificationuc.GetUnreadMessagesCountReq](r)
	if err != nil {
		httpx.HandleError(w, err)
		return
	}

	resp, err := c.notificationUsecase.GetUnreadMessagesCount(r.Context(), req)
	if err != nil {
		httpx.HandleError(w, err)
		return
	}

	httpx.WriteResponse(http.StatusOK, w, resp)
}

func (c *ctrl) getUnreadMessagesCountByChat(w http.ResponseWriter, r *http.Request) {
	req, err := httpx.BindRequest[notificationuc.GetUnreadMessagesCountByChatReq](r)
	if err != nil {
		httpx.HandleError(w, err)
		return
	}

	resp, err := c.notificationUsecase.GetUnreadMessagesCountByChat(r.Context(), req)
	if err != nil {
		httpx.HandleError(w, err)
		return
	}

	httpx.WriteResponse(http.StatusOK, w, resp)
}

func (c *ctrl) markMessagesAsRead(w http.ResponseWriter, r *http.Request) {
	req, err := httpx.BindRequest[notificationuc.MarkMessagesAsReadReq](r)
	if err != nil {
		httpx.HandleError(w, err)
		return
	}

	err = c.notificationUsecase.MarkMessagesAsRead(r.Context(), req)
	if err != nil {
		httpx.HandleError(w, err)
		return
	}

	httpx.WriteResponse(http.StatusNoContent, w, nil)
}

func (c *ctrl) getOnlineStatusByUsers(w http.ResponseWriter, r *http.Request) {
	req, err := httpx.BindRequest[notificationuc.GetOnlineStatusByUsersReq](r)
	if err != nil {
		httpx.HandleError(w, err)
		return
	}

	resp, err := c.notificationUsecase.GetOnlineStatusByUsers(r.Context(), req)
	if err != nil {
		httpx.HandleError(w, err)
		return
	}

	httpx.WriteResponse(http.StatusOK, w, resp)
}

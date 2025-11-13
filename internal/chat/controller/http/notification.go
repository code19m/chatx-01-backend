package http

import (
	"chatx-01/internal/chat/usecase/notificationuc"
	"chatx-01/pkg/httptools"
	"net/http"
)

func (c *ctrl) getUnreadMessagesCount(w http.ResponseWriter, r *http.Request) {
	req, err := httptools.BindRequest[notificationuc.GetUnreadMessagesCountReq](r)
	if err != nil {
		httptools.HandleError(w, err)
		return
	}

	resp, err := c.notificationUsecase.GetUnreadMessagesCount(r.Context(), req)
	if err != nil {
		httptools.HandleError(w, err)
		return
	}

	httptools.WriteResponse(http.StatusOK, w, resp)
}

func (c *ctrl) getUnreadMessagesCountByChat(w http.ResponseWriter, r *http.Request) {
	req, err := httptools.BindRequest[notificationuc.GetUnreadMessagesCountByChatReq](r)
	if err != nil {
		httptools.HandleError(w, err)
		return
	}

	resp, err := c.notificationUsecase.GetUnreadMessagesCountByChat(r.Context(), req)
	if err != nil {
		httptools.HandleError(w, err)
		return
	}

	httptools.WriteResponse(http.StatusOK, w, resp)
}

func (c *ctrl) markMessagesAsRead(w http.ResponseWriter, r *http.Request) {
	req, err := httptools.BindRequest[notificationuc.MarkMessagesAsReadReq](r)
	if err != nil {
		httptools.HandleError(w, err)
		return
	}

	err = c.notificationUsecase.MarkMessagesAsRead(r.Context(), req)
	if err != nil {
		httptools.HandleError(w, err)
		return
	}

	httptools.WriteResponse(http.StatusNoContent, w, nil)
}

func (c *ctrl) getOnlineStatusByUsers(w http.ResponseWriter, r *http.Request) {
	req, err := httptools.BindRequest[notificationuc.GetOnlineStatusByUsersReq](r)
	if err != nil {
		httptools.HandleError(w, err)
		return
	}

	resp, err := c.notificationUsecase.GetOnlineStatusByUsers(r.Context(), req)
	if err != nil {
		httptools.HandleError(w, err)
		return
	}

	httptools.WriteResponse(http.StatusOK, w, resp)
}

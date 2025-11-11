package http

import (
	"chatx-01/internal/chat/usecase/messageuc"
	"chatx-01/pkg/httpx"
	"net/http"
)

func (c *ctrl) getMessagesList(w http.ResponseWriter, r *http.Request) {
	req, err := httpx.BindRequest[messageuc.GetMessagesListReq](r)
	if err != nil {
		httpx.HandleError(w, err)
		return
	}

	resp, err := c.messageUsecase.GetMessagesList(r.Context(), req)
	if err != nil {
		httpx.HandleError(w, err)
		return
	}

	httpx.WriteResponse(http.StatusOK, w, resp)
}

func (c *ctrl) sendMessage(w http.ResponseWriter, r *http.Request) {
	req, err := httpx.BindRequest[messageuc.SendMessageReq](r)
	if err != nil {
		httpx.HandleError(w, err)
		return
	}

	resp, err := c.messageUsecase.SendMessage(r.Context(), req)
	if err != nil {
		httpx.HandleError(w, err)
		return
	}

	httpx.WriteResponse(http.StatusCreated, w, resp)
}

func (c *ctrl) editMessage(w http.ResponseWriter, r *http.Request) {
	req, err := httpx.BindRequest[messageuc.EditMessageReq](r)
	if err != nil {
		httpx.HandleError(w, err)
		return
	}

	err = c.messageUsecase.EditMessage(r.Context(), req)
	if err != nil {
		httpx.HandleError(w, err)
		return
	}

	httpx.WriteResponse(http.StatusNoContent, w, nil)
}

func (c *ctrl) deleteMessage(w http.ResponseWriter, r *http.Request) {
	req, err := httpx.BindRequest[messageuc.DeleteMessageReq](r)
	if err != nil {
		httpx.HandleError(w, err)
		return
	}

	err = c.messageUsecase.DeleteMessage(r.Context(), req)
	if err != nil {
		httpx.HandleError(w, err)
		return
	}

	httpx.WriteResponse(http.StatusNoContent, w, nil)
}

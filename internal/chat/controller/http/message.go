package http

import (
	"chatx-01/internal/chat/usecase/messageuc"
	"chatx-01/pkg/httptools"
	"net/http"
)

func (c *ctrl) getMessagesList(w http.ResponseWriter, r *http.Request) {
	req, err := httptools.BindRequest[messageuc.GetMessagesListReq](r)
	if err != nil {
		httptools.HandleError(w, err)
		return
	}

	resp, err := c.messageUsecase.GetMessagesList(r.Context(), req)
	if err != nil {
		httptools.HandleError(w, err)
		return
	}

	httptools.WriteResponse(http.StatusOK, w, resp)
}

func (c *ctrl) sendMessage(w http.ResponseWriter, r *http.Request) {
	req, err := httptools.BindRequest[messageuc.SendMessageReq](r)
	if err != nil {
		httptools.HandleError(w, err)
		return
	}

	resp, err := c.messageUsecase.SendMessage(r.Context(), req)
	if err != nil {
		httptools.HandleError(w, err)
		return
	}

	httptools.WriteResponse(http.StatusCreated, w, resp)
}

func (c *ctrl) editMessage(w http.ResponseWriter, r *http.Request) {
	req, err := httptools.BindRequest[messageuc.EditMessageReq](r)
	if err != nil {
		httptools.HandleError(w, err)
		return
	}

	err = c.messageUsecase.EditMessage(r.Context(), req)
	if err != nil {
		httptools.HandleError(w, err)
		return
	}

	httptools.WriteResponse(http.StatusNoContent, w, nil)
}

func (c *ctrl) deleteMessage(w http.ResponseWriter, r *http.Request) {
	req, err := httptools.BindRequest[messageuc.DeleteMessageReq](r)
	if err != nil {
		httptools.HandleError(w, err)
		return
	}

	err = c.messageUsecase.DeleteMessage(r.Context(), req)
	if err != nil {
		httptools.HandleError(w, err)
		return
	}

	httptools.WriteResponse(http.StatusNoContent, w, nil)
}

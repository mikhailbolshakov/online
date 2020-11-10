package server

import (
	"chats/models"
	"chats/sentry"
	"chats/infrastructure"
	"encoding/json"
)

type Router struct {
	routes map[string]func(h *Hub, c *Client, clientRequest []byte)
}

func NewRouter() *Router {
	return &Router{
		routes: make(map[string]func(h *Hub, c *Client, clientRequest []byte)),
	}
}

func (r *Router) Handle(
	route string,
	handle func(
		h *Hub,
		c *Client,
		clientRequst []byte,
	),
) {
	r.routes[route] = handle
}

func (r *Router) Dispatch(h *Hub, c *Client, clientRequest []byte) {
	request := &models.WSChatRequest{}
	err := json.Unmarshal(clientRequest, &request)
	if err != nil {
		infrastructure.SetError(&sentry.SystemError{
			Error:   err,
			Message: infrastructure.UnmarshallingError,
			Code:    infrastructure.UnmarshallingErrorCode,
		})

		return
	}

	if handler, ok := r.routes[request.Type]; !ok {
		infrastructure.SetError(&sentry.SystemError{
			Error:   nil,
			Message: WsEventTypeNotExists,
			Code:    WsEventTypeNotExistsCode,
		})

		return
	} else {
		handler(h, c, clientRequest)
	}
}

func SetRouter() *Router {
	router := NewRouter()
	event := NewEvent()
	router.Handle(EventMessage, event.EventMessage)
	router.Handle(EventMessageStatus, event.EventMessageStatus)
	router.Handle(EventOpponentStatus, event.EventOpponentStatus)
	router.Handle(EventJoin, event.EventJoin)
	router.Handle(EventTyping, event.EventTyping)
	router.Handle(EventEcho, event.EventEcho)

	return router
}

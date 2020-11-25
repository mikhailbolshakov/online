package server

import (
	"chats/app"
	"chats/system"
	"encoding/json"
)

type Router struct {
	routes map[string]func(h *Hub, c *Session, clientRequest []byte)
}

func NewRouter() *Router {
	return &Router{
		routes: make(map[string]func(h *Hub, c *Session, clientRequest []byte)),
	}
}

func (r *Router) Handle(
	route string,
	handle func(
		h *Hub,
		c *Session,
		clientRequst []byte,
	),
) {
	r.routes[route] = handle
}

func (r *Router) Dispatch(h *Hub, c *Session, clientRequest []byte) {
	request := &WSChatRequest{}
	err := json.Unmarshal(clientRequest, &request)
	if err != nil {
		app.E().SetError(&system.Error{
			Error:   err,
			Message: system.GetError(system.UnmarshallingErrorCode),
			Code:    system.UnmarshallingErrorCode,
		})

		return
	}

	if handler, ok := r.routes[request.Type]; !ok {
		app.E().SetError(system.SysErr(nil, system.WsEventTypeNotExistsCode, nil))
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

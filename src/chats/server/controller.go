package server

import (
	"chats/sdk"
	"chats/system"
	"encoding/json"
	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
	"net/http"
	"strings"
	"sync"
	"time"
)

type WsServer struct {
	apiTopic            string
	shutdownSleep       time.Duration
	port                string
	hub                 *Hub
	server              *http.Server
	logs                *system.Logs
	actualAccounts      map[uuid.UUID]time.Time
	actualAccountsMutex sync.Mutex
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (ws *WsServer) ApiConsumer() {
	apiConsumerErrorChan := make(chan *sdk.Error, 2048)
	go ws.hub.app.Sdk.
		Subject(ws.apiTopic).
		ApiConsumer(ws.GetHandler, apiConsumerErrorChan)

	for {
		err := <-apiConsumerErrorChan
		system.ErrHandler.SetError(&system.Error{
			Error:   err.Error,
			Message: err.Message,
			Code:    err.Code,
			Data:    err.Data,
		})
	}
}

func (ws *WsServer) GetHandler(request []byte) ([]byte, *sdk.Error) {
	data, err := ws.Router(request)
	if err != nil {
		return nil, &sdk.Error{
			Error:   err.Error,
			Message: err.Message,
			Code:    err.Code,
			Data:    err.Data,
		}
	} else {
		return data, nil
	}
}

func (ws *WsServer) Router(request []byte) ([]byte, *system.Error) {

	clientRequest := &sdk.ApiRequest{}
	err := json.Unmarshal(request, &clientRequest)
	if err != nil {
		return nil, system.UnmarshalRequestError1201(err, request)
	}

	switch strings.ToUpper(clientRequest.Method) {
	case http.MethodGet:
		switch clientRequest.Path {

		case "/chats/messages/update",
			"/chats/chat/recent":
			return ws.getChatRecent(request)

		case "/chats/messages/history",
			"/chats/chat/history":
			return ws.getChatHistory(request)

		}
	case http.MethodPost:
		switch clientRequest.Path {

		case "/chats/message":
			return ws.setChatMessage(request)

		case "/ws/client/message":
			return ws.sendClientMessage(request)

		}

	case http.MethodPut:
		switch clientRequest.Path {

		case "/chats/account/subscribe":
			return ws.changeChatAccountSubscribe(request)

		}
	}

	return nil, &system.Error{
		Error:   nil,
		Message: sdk.GetError(1203),
		Code:    1203,
		Data:    request,
	}
}


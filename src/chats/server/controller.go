package server

import (
	"chats/database"
	"chats/sentry"
	"chats/infrastructure"
	"chats/sdk"
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/mkevac/gopinba"
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
	logs                *infrastructure.Logs
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
		infrastructure.SetError(&sentry.SystemError{
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

func (ws *WsServer) Router(request []byte) ([]byte, *sentry.SystemError) {

	clientRequest := &sdk.ApiRequest{}
	err := json.Unmarshal(request, &clientRequest)
	if err != nil {
		return nil, infrastructure.UnmarshalRequestError1201(err, request)
	}

	if ws.hub.app.DB.Pinba != nil {
		tags := map[string]string{
			"group":  "sdk",
			"topic":  ws.apiTopic,
			"method": clientRequest.Method,
			"path":   clientRequest.Path,
		}
		timer := gopinba.TimerStart(tags)
		defer func(timer *gopinba.Timer, db *database.Storage) {
			timer.Stop()
			db.Pinba.SendRequest(&gopinba.Request{
				Tags:        timer.Tags,
				RequestTime: timer.GetDuration(),
			})
		}(timer, ws.hub.app.DB)
	}

	switch strings.ToUpper(clientRequest.Method) {
	case http.MethodGet:
		switch clientRequest.Path {
		case "/chats/chats":
			return ws.getChatChats(request)

		case "/chats/chat":
			return ws.getChatById(request)

		case "/order/chats":
			return ws.ChatByOrder(request)

		case "/chats/info":
			return ws.getChatsInfo(request)

		case "/chats/messages/update",
			"/chats/chat/recent":
			return ws.getChatRecent(request)

		case "/chats/messages/history",
			"/chats/chat/history":
			return ws.getChatHistory(request)

		case "/chats/last":
			return ws.getLastChat(request)

		}
	case http.MethodPost:
		switch clientRequest.Path {

		case "/account":
			return ws.createAccount(request)

		case "/chats/new":
			return ws.setChatNew(request)

		case "/chats/new/subscribe":
			return ws.setChatNewSubscribe(request)

		case "/chats/account/subscribe":
			return ws.setChatAccountSubscribe(request)

		case "/chats/account/unsubscribe":
			return ws.setChatUserUnsubscribe(request)

		case "/chats/message":
			return ws.setChatMessage(request)

		case "/chats/status":
			return ws.setChatStatus(request)

		case "/ws/client/message":
			return ws.sendClientMessage(request)

		case "/ws/client/consultation/update":
			return ws.clientConsultationUpdate(request)
		}

	case http.MethodPut:
		switch clientRequest.Path {

		case "/chats/account/subscribe":
			return ws.changeChatUserSubscribe(request)

		case "/account":
			return ws.updateAccount(request)

		}
	}

	return nil, &sentry.SystemError{
		Error:   nil,
		Message: sdk.GetError(1203),
		Code:    1203,
		Data:    request,
	}
}


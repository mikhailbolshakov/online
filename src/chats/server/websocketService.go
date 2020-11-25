package server

import (
	"chats/app"
	a "chats/repository/account"
	rr "chats/repository/room"
	"chats/system"
	"encoding/json"
	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
	"net/http"
)

type WebSocketService struct {
	ws *WsServer
}

func (s *WebSocketService) setRouting(router *mux.Router) {

	router.HandleFunc("/ws/", func(w http.ResponseWriter, r *http.Request) {
		app.L().Debug("request:" + r.Host + r.URL.EscapedPath())
		s.AccountConnect(w, r)
	})

}

func createResponse(response *WSChatErrorResponse) []byte {
	result, err := json.Marshal(response)

	if err != nil {
		app.E().SetError(system.SysErr(err, system.WsCreateClientResponseCode, nil))
	}

	return result
}

func (s *WebSocketService) AccountConnect(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	//	upgrade websocket connection
	conn, err := s.ws.httpServer.wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		app.E().SetError(system.SysErr(err, system.WsUpgradeProblemCode, nil))
		return
	}

	//	userToken
	token := r.URL.Query().Get("token")
	app.L().Debugf("Session with token %s is connecting \n", token)
	if token == "" {
		response := &WSChatErrorResponse{
			Error: WSChatErrorErrorResponse{
				Message: system.WsEmptyToken,
				Code:    system.WsEmptyTokenCode,
			},
		}
		w.Write(createResponse(response))
		app.E().SetError(system.SysErr(err, system.WsEmptyTokenCode, []byte("token: " + token)))
		return
	}

	// get registered account by the ID (token) passed from the client
	// currently we assume that account Id is passed
	// if token comes we need to verify it with the external system
	accRep := a.CreateRepository(app.GetDB())
	account, sysErr := accRep.GetAccount(uuid.FromStringOrNil(token), "")
	app.L().Debugf("Account found by token: %v \n", account)
	if sysErr != nil || account.Id == uuid.Nil {
		response := &WSChatErrorResponse{
			Error: WSChatErrorErrorResponse{
				Message: system.WsUserIdentification,
				Code:    system.WsUserIdentificationCode,
			},
		}
		w.Write(createResponse(response))
		app.E().SetError(system.SysErr(err, system.WsUserIdentificationCode, []byte("token: " + token)))
		return
	}

	// initialize account WS session
	session := InitSession(s.ws.hub, conn)

	// try to find existent rooms with the account subscribed
	// if no rooms, empty map is retrieved
	rooms := make(map[uuid.UUID]*Room)
	roomRep := rr.CreateRepository(app.GetDB())
	subscribers := roomRep.GetAccountSubscribers(account.Id)
	app.L().Debugf("Subscribers found: %v \n", subscribers)
	for roomId, _ := range subscribers {
		room := s.ws.hub.LoadRoomIfNotExists(roomId)
		app.L().Debugf("Room loaded and added to session: %v \n", room)
		room.AddSession(session.sessionId)
		rooms[roomId] = room
	}
	session.account = ConvertAccountFromModel(account)
	session.rooms = rooms
	session.subscribers = subscribers

	s.ws.hub.registerChan <- session
	go session.Write()
	go session.Read()

}

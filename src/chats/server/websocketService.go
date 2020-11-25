package server

import (
	"chats/models"
	"chats/system"
	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
	"log"
	"net/http"
)

type WebSocketService struct {
	ws *WsServer
}

func (s *WebSocketService) setRouting(router *mux.Router) {

	router.HandleFunc("/online/ws", func(w http.ResponseWriter, r *http.Request) {
		log.Println("request:" + r.Host + r.URL.EscapedPath())
		s.AccountConnect(w, r)
	})

}

func (s *WebSocketService) AccountConnect(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	//	upgrade websocket connection
	conn, err := s.ws.httpServer.wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		system.ErrHandler.SetError(&system.Error{
			Error:   err,
			Message: WsUpgradeProblem,
			Code:    WsUpgradeProblemCode,
		})

		return
	}

	//	userToken
	token := r.URL.Query().Get("token")
	log.Printf("Session with token %s is connecting \n", token)
	if token == "" {
		response := &models.WSChatErrorResponse{
			Error: models.WSChatErrorErrorResponse{
				Message: WsEmptyToken,
				Code:    WsEmptyTokenCode,
			},
		}
		w.Write(createResponse(response))
		system.ErrHandler.SetError(&system.Error{
			Error:   nil,
			Message: WsEmptyToken,
			Code:    WsEmptyTokenCode,
			Data:    []byte("token: " + token),
		})

		return
	}

	// get registered account by the ID (token) passed from the client
	// currently we assume that account Id is passed
	// if token comes we need to verify it with the external system
	account, sysErr := s.ws.hub.app.DB.GetAccount(uuid.FromStringOrNil(token), "")
	log.Printf("Account found by token: %v \n", account)
	if sysErr != nil || account.Id == uuid.Nil {
		response := &models.WSChatErrorResponse{
			Error: models.WSChatErrorErrorResponse{
				Message: WsUserIdentification,
				Code:    WsUserIdentificationCode,
			},
		}
		w.Write(createResponse(response))
		system.ErrHandler.SetError(&system.Error{
			Error:   nil,
			Message: WsUserIdentification,
			Code:    WsUserIdentificationCode,
			Data:    []byte("token: " + token),
		})

		return
	}

	// initialize account WS session
	session := InitSession(s.ws.hub, conn)

	// try to find existent rooms with the account subscribed
	// if no rooms, empty map is retrieved
	rooms := make(map[uuid.UUID]*Room)
	subscribers := s.ws.hub.app.DB.GetAccountSubscribers(account.Id)
	log.Printf("Subscribers found: %v \n", subscribers)
	for roomId, _ := range subscribers {
		room := s.ws.hub.LoadRoomIfNotExists(roomId)
		log.Printf("Room loaded and added to session: %v \n", room)
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

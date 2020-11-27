package server

import (
	"chats/app"
	"chats/system"
	r "chats/repository/room"
	"encoding/json"
	uuid "github.com/satori/go.uuid"
	"sync"
	"time"
)

type Hub struct {
	sessions        map[uuid.UUID]*Session
	accountSessions map[uuid.UUID]*Session
	accounts        map[uuid.UUID]bool
	rooms           map[uuid.UUID]*Room
	roomMutex       sync.Mutex
	registerChan    chan *Session
	unregisterChan  chan *Session
	messageChan     chan *RoomMessage
	router          *Router
}

func NewHub() *Hub {

	return &Hub{
		sessions:        make(map[uuid.UUID]*Session),
		accountSessions: make(map[uuid.UUID]*Session),
		accounts:        make(map[uuid.UUID]bool),
		rooms:           make(map[uuid.UUID]*Room),
		registerChan:    make(chan *Session),
		unregisterChan:  make(chan *Session),
		messageChan:     make(chan *RoomMessage),
		router:          SetRouter(),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case session := <-h.registerChan:
			h.sessions[session.sessionId] = session
			h.accountSessions[session.account.Id] = session
			h.accounts[session.account.Id] = true
			app.L().Debug(">>> session register:", session.account.Id) //	TODO
			h.checkConnectionStatus(session.account.Id, true)
		case session := <-h.unregisterChan:
			go h.clientConnectionChange(session)
			h.onSessionDisconnect(session)
			app.L().Debug(">>> session unregister:", session.account.Id) //	TODO
			h.checkConnectionStatus(session.account.Id, false)
		case message := <-h.messageChan:
			app.L().Debugf("Sending message to internal topic. roomId: %s, accountId: %s", message.RoomId, message.AccountId)

			{ //	Scaling
				answer, err := json.Marshal(message)
				if err != nil {
					app.L().Debug("Не удалось сформировать ответ клиенту")
					app.L().Debug(err)
				} else {
					go app.Instance.Inf.Nats.Subject(app.Instance.Inf.Nats.InsideTopic()).Publish(answer) //	return sentry
				}
			}
		}
	}
}

func (h *Hub) onSessionDisconnect(session *Session) {

	if _, ok := h.sessions[session.sessionId]; ok {

		app.L().Debugf("Session cleanup %s", session.sessionId)

		// set account status offline
		err := wsServer.setAccountOffline(session.account.Id)
		if err != nil {
			app.E().SetError(err)
		}

		delete(h.sessions, session.sessionId)
		h.removeSessionFromRooms(session)
		close(session.sendChan)

		if !h.accountHasSessions(session.account) {
			delete(h.accounts, session.account.Id)
			delete(h.accountSessions, session.account.Id)
		}

	}
}

func (h *Hub) accountHasSessions(account *Account) bool {
	for _, s := range h.sessions {
		if s.account.Id == account.Id {
			return true
		}
	}

	return false
}

func (h *Hub) removeSessionFromRooms(session *Session) {
	for _, room := range session.rooms {
		room.removeSession(session)
	}
}

func (h *Hub) removeAllSessions() {
	for _, session := range h.sessions {
		h.onSessionDisconnect(session)
	}
}

func (h *Hub) LoadRoomIfNotExists(roomId uuid.UUID) *Room {

	rep := r.CreateRepository(app.Instance.Inf.DB)
	subscribers := rep.GetRoomAccountSubscribers(roomId)

	h.roomMutex.Lock()
	defer h.roomMutex.Unlock()

	if _, ok := h.rooms[roomId]; !ok {
		room := InitRoom(roomId, subscribers)
		h.rooms[roomId] = room
		app.L().Debugf("New room is loaded. room_id %s \n", roomId)
		return room
	} else {
		h.rooms[roomId].subscribers = subscribers
		app.L().Debugf("Existent room is found. room_id %s \n", roomId)
	}

	return h.rooms[roomId]
}

func (h *Hub) onMessage(message []byte, client *Session) {
	h.router.Dispatch(h, client, message)
}

func (h *Hub) SendMessageToRoom(message *RoomMessage) {
	h.messageChan <- message
}

func (h *Hub) sendMessage(session *Session, message []byte) {
	session.send(message)
}

func (h *Hub) checkConnectionStatus(userId uuid.UUID, online bool) {

	// TODO: send message to BUS

	/*
	if !online {
		if _, ok := h.accountSessions[userId]; !ok {
			go h.inf.Nats.ChangeConnectionStatus(userId, online)
		}
	} else {
		go h.inf.Nats.ChangeConnectionStatus(userId, online)
	}
	 */
}

func (h *Hub) clientConnectionChange(session *Session) {
	time.Sleep(10 * time.Second)

	//	todo вынести в отдельный метод
	cronGetUserStatusRequestMessage := &CronGetUserStatusRequest{
		Type: MessageTypeGetUserStatus,
		Data: CronGetAccountStatusRequestData{
			AccountId: session.account.Id,
		},
	}

	request, err := json.Marshal(cronGetUserStatusRequestMessage)
	if err != nil {
		app.E().SetError(system.MarshalError1011(err, nil))
		return
	}

	response, cronRequestError := app.GetNats().
		Subject(app.GetNats().CronTopic()).
		Request(request)

	if cronRequestError != nil {
		app.E().SetError(&system.Error{
			Error:   cronRequestError.Error,
			Message: system.GetError(system.CronResponseErrorCode),
			Code:    system.CronResponseErrorCode,
			Data:    append(request, response...),
		})
		return
	}

	cronGetUserResponse := &CronGetAccountStatusResponse{}
	err = json.Unmarshal(response, cronGetUserResponse)
	if err != nil {
		app.E().SetError(system.MarshalError1011(err, response))
		return
	}

	if !cronGetUserResponse.Online {
		//if _, ok := h.accountSessions[session.account.Id]; !ok {	//	plan B
		rep := r.CreateRepository(app.GetDB())
		opponentId := rep.LastOpponentId(session.account.Id)

		if opponentId != uuid.Nil {
			data := new(WSAccountStatusModel)
			data.AccountId = session.account.Id

			roomMessage := &RoomMessage{
				AccountId: opponentId,
				Message: &WSChatResponse{
					Type: EventClientConnectionError,
					Data: &data,
				},
			}

			go h.SendMessageToRoom(roomMessage)
		}

		/*for chatId, item := range session.subscribers {
			if item.Role == database.UserTypeClient {
				data := new(models.WSAccountStatusModel)
				data.AccountId = session.account.Id

				roomMessage := &RoomMessage{
					RoomId: chatId,
					Message: &models.WSChatResponse{
						Type: EventClientConnectionError,
						Data: &data,
					},
				}

				go h.SendMessageToRoom(roomMessage)
			}
		}*/
	}
}

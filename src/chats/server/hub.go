package server

import (
	"chats/application"
	"chats/system"
	"chats/models"
	"chats/sdk"
	"encoding/json"
	uuid "github.com/satori/go.uuid"
	"log"
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
	app             *application.Application
}

func NewHub(app *application.Application) *Hub {

	return &Hub{
		sessions:        make(map[uuid.UUID]*Session),
		accountSessions: make(map[uuid.UUID]*Session),
		accounts:        make(map[uuid.UUID]bool),
		rooms:           make(map[uuid.UUID]*Room),
		registerChan:    make(chan *Session),
		unregisterChan:  make(chan *Session),
		messageChan:     make(chan *RoomMessage),
		router:          SetRouter(),
		app:             app,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case session := <-h.registerChan:
			h.sessions[session.sessionId] = session
			h.accountSessions[session.account.Id] = session
			h.accounts[session.account.Id] = true
			log.Println(">>> session register:", session.account.Id) //	TODO
			h.checkConnectionStatus(session.account.Id, true)
		case session := <-h.unregisterChan:
			go h.clientConnectionChange(session)
			h.onSessionDisconnect(session)
			log.Println(">>> session unregister:", session.account.Id) //	TODO
			h.checkConnectionStatus(session.account.Id, false)
		case message := <-h.messageChan:
			log.Printf("Sending message to internal topic: %v \n", message.Message)

			{ //	Scaling
				answer, err := json.Marshal(message)
				if err != nil {
					log.Println("Не удалось сформировать ответ клиенту")
					log.Println(err)
				} else {
					go h.app.Sdk.Subject(system.InsideTopic()).Publish(answer) //	return sentry
				}
			}
		}
	}
}

func (h *Hub) onSessionDisconnect(session *Session) {
	if _, ok := h.sessions[session.sessionId]; ok {
		delete(h.sessions, session.sessionId)
		h.removeSessionFromRooms(session)
		close(session.sendChan)

		if !h.accountHasSessions(session.account) {
			delete(h.accounts, session.account.Id)
			delete(h.accountSessions, session.account.Id)
		}
	}
}

func (h *Hub) accountHasSessions(account *sdk.Account) bool {
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

	subscribers := h.app.DB.GetRoomAccountSubscribers(roomId)

	h.roomMutex.Lock()
	defer h.roomMutex.Unlock()

	if _, ok := h.rooms[roomId]; !ok {
		room := InitRoom(roomId, subscribers)
		h.rooms[roomId] = room
		log.Printf("New room is loaded. room_id %s \n", roomId)
		return room
	} else {
		h.rooms[roomId].subscribers = subscribers
		log.Printf("Existent room is found. room_id %s \n", roomId)
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
			go h.app.Sdk.ChangeConnectionStatus(userId, online)
		}
	} else {
		go h.app.Sdk.ChangeConnectionStatus(userId, online)
	}
	 */
}

func (h *Hub) clientConnectionChange(session *Session) {
	time.Sleep(10 * time.Second)

	//	todo вынести в отдельный метод
	cronGetUserStatusRequestMessage := &models.CronGetUserStatusRequest{
		Type: MessageTypeGetUserStatus,
		Data: models.CronGetAccountStatusRequestData{
			AccountId: session.account.Id,
		},
	}

	request, err := json.Marshal(cronGetUserStatusRequestMessage)
	if err != nil {
		system.ErrHandler.SetError(system.MarshalError1011(err, nil))
		return
	}

	response, cronRequestError := h.app.Sdk.
		Subject(system.CronTopic()).
		Request(request)

	if cronRequestError != nil {
		system.ErrHandler.SetError(&system.Error{
			Error:   cronRequestError.Error,
			Message: system.CronResponseError,
			Code:    system.CronResponseErrorCode,
			Data:    append(request, response...),
		})
		return
	}

	cronGetUserResponse := &models.CronGetAccountStatusResponse{}
	err = json.Unmarshal(response, cronGetUserResponse)
	if err != nil {
		system.ErrHandler.SetError(system.MarshalError1011(err, response))
		return
	}

	if !cronGetUserResponse.Online {
		//if _, ok := h.accountSessions[session.account.Id]; !ok {	//	plan B
		opponentId := h.app.DB.LastOpponentId(session.account.Id)

		if opponentId != uuid.Nil {
			data := new(models.AccountStatusModel)
			data.AccountId = session.account.Id

			roomMessage := &RoomMessage{
				AccountId: opponentId,
				Message: &models.WSChatResponse{
					Type: EventClientConnectionError,
					Data: &data,
				},
			}

			go h.SendMessageToRoom(roomMessage)
		}

		/*for chatId, item := range session.subscribers {
			if item.Role == database.UserTypeClient {
				data := new(models.AccountStatusModel)
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

package server

import (
	"chats/application"
	"chats/models"
	"chats/sentry"
	"chats/infrastructure"
	"encoding/json"
	"chats/sdk"
	uuid "github.com/satori/go.uuid"
	"log"
	"sync"
	"time"
)

type Hub struct {
	clients        map[string]*Client
	clientsId      map[uuid.UUID]*Client
	accounts       map[uuid.UUID]bool
	rooms          map[uuid.UUID]*Room
	roomMutex      sync.Mutex
	registerChan   chan *Client
	unregisterChan chan *Client
	messageChan    chan *RoomMessage
	router         *Router
	app            *application.Application
}

func NewHub(app *application.Application) *Hub {

	return &Hub{
		clients:        make(map[string]*Client),
		clientsId:      make(map[uuid.UUID]*Client),
		accounts:       make(map[uuid.UUID]bool),
		rooms:          make(map[uuid.UUID]*Room),
		registerChan:   make(chan *Client),
		unregisterChan: make(chan *Client),
		messageChan:    make(chan *RoomMessage),
		router:         SetRouter(),
		app:            app,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.registerChan:
			h.clients[client.uniqId] = client
			h.clientsId[client.account.Id] = client
			h.accounts[client.account.Id] = true
			log.Println(">>> client register:", client.account.Id) //	TODO
			h.checkConnectionStatus(client.account.Id, true)
		case client := <-h.unregisterChan:
			go h.clientConnectionChange(client)
			h.onClientDisconnect(client)
			log.Println(">>> client unregister:", client.account.Id) //	TODO
			h.checkConnectionStatus(client.account.Id, false)
		case message := <-h.messageChan:
			log.Println(">>> message to client:", message.Message) //	TODO

			{ //	Scaling
				answer, err := json.Marshal(message)
				if err != nil {
					log.Println("Не удалось сформировать ответ клиенту")
					log.Println(err)
				} else {
					go h.app.Sdk.Subject(infrastructure.InsideTopic()).Publish(answer) //	return sentry
				}
			}
		}
	}
}

func (h *Hub) onClientDisconnect(client *Client) {
	if _, ok := h.clients[client.uniqId]; ok {
		delete(h.clients, client.uniqId)
		h.removeClientFromRooms(client)
		close(client.sendChan)

		if !h.userHasClients(client.account) {
			delete(h.accounts, client.account.Id)
			delete(h.clientsId, client.account.Id)
		}
	}
}

func (h *Hub) userHasClients(user *sdk.Account) bool {
	for _, otherClient := range h.clients {
		if otherClient.account.Id == user.Id {
			return true
		}
	}

	return false
}

func (h *Hub) removeClientFromRooms(client *Client) {
	for _, room := range client.rooms {
		room.removeClient(client)
	}
}

func (h *Hub) removeAllClients() {
	for _, client := range h.clients {
		h.onClientDisconnect(client)
	}
}

func (h *Hub) CreateRoomIfNotExists(chatId uuid.UUID) *Room {
	h.roomMutex.Lock()
	defer h.roomMutex.Unlock()

	subscribers := h.app.DB.ChatSubscribes(chatId)
	if _, ok := h.rooms[chatId]; !ok {
		room := CreateRoom(chatId, subscribers)
		h.rooms[chatId] = room

		return room
	} else {
		h.rooms[chatId].subscribers = subscribers
	}

	return h.rooms[chatId]
}

func (h *Hub) onClientMessage(message []byte, client *Client) {
	h.router.Dispatch(h, client, message)
}

func (h *Hub) SendMessageToRoom(message *RoomMessage) {
	h.messageChan <- message
}

func (h *Hub) sendMessageToClient(client *Client, message []byte) {
	client.send(message)
}

func (h *Hub) checkConnectionStatus(userId uuid.UUID, online bool) {

	// TODO: send message to BUS

	/*
	if !online {
		if _, ok := h.clientsId[userId]; !ok {
			go h.app.Sdk.ChangeConnectionStatus(userId, online)
		}
	} else {
		go h.app.Sdk.ChangeConnectionStatus(userId, online)
	}
	 */
}

func (h *Hub) clientConnectionChange(c *Client) {
	time.Sleep(10 * time.Second)

	//	todo вынести в отдельный метод
	cronGetUserStatusRequestMessage := &models.CronGetUserStatusRequest{
		Type: MessageTypeGetUserStatus,
		Data: models.CronGetAccountStatusRequestData{
			AccountId: c.account.Id,
		},
	}

	request, err := json.Marshal(cronGetUserStatusRequestMessage)
	if err != nil {
		infrastructure.SetError(infrastructure.MarshalError1011(err, nil))
		return
	}

	response, cronRequestError := h.app.Sdk.
		Subject(infrastructure.CronTopic()).
		Request(request)

	if cronRequestError != nil {
		infrastructure.SetError(&sentry.SystemError{
			Error:   cronRequestError.Error,
			Message: infrastructure.CronResponseError,
			Code:    infrastructure.CronResponseErrorCode,
			Data:    append(request, response...),
		})
		return
	}

	cronGetUserResponse := &models.CronGetAccountStatusResponse{}
	err = json.Unmarshal(response, cronGetUserResponse)
	if err != nil {
		infrastructure.SetError(infrastructure.MarshalError1011(err, response))
		return
	}

	if !cronGetUserResponse.Online {
		//if _, ok := h.clientsId[c.account.Id]; !ok {	//	plan B
		opponentId := h.app.DB.LastOpponentId(c.account.Id)

		if opponentId != uuid.Nil {
			data := new(models.AccountStatusModel)
			data.AccountId = c.account.Id

			roomMessage := &RoomMessage{
				AccountId: opponentId,
				Message: &models.WSChatResponse{
					Type: EventClientConnectionError,
					Data: &data,
				},
			}

			go h.SendMessageToRoom(roomMessage)
		}

		/*for chatId, item := range c.subscribes {
			if item.Role == database.UserTypeClient {
				data := new(models.AccountStatusModel)
				data.AccountId = c.account.Id

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

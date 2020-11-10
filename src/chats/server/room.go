package server

import (
	"chats/database"
	"chats/models"
	uuid "github.com/satori/go.uuid"
	"sync"
)

type Room struct {
	clients     map[string]bool
	subscribers []database.SubscribeAccountModel
	clientMutex sync.Mutex
	roomId      uuid.UUID
}

type RoomMessage struct {
	SendPush  bool
	AccountId uuid.UUID
	RoomId    uuid.UUID
	Message   *models.WSChatResponse
}

func CreateRoom(roomId uuid.UUID, subscribers []database.SubscribeAccountModel) *Room {
	return &Room{
		clients:     make(map[string]bool),
		subscribers: subscribers,
		roomId:      roomId,
	}
}

func (r *Room) removeClient(client *Client) {
	r.clientMutex.Lock()
	defer r.clientMutex.Unlock()

	if _, ok := r.clients[client.uniqId]; ok {
		delete(r.clients, client.uniqId)
	}
}

func (r *Room) AddClient(clientUniqId string) {
	r.clientMutex.Lock()
	defer r.clientMutex.Unlock()

	r.clients[clientUniqId] = true
}

func (r *Room) getRoomClientIds() []string {
	r.clientMutex.Lock()
	defer r.clientMutex.Unlock()

	var uniqueIds []string

	for uniqueId, _ := range r.clients {
		uniqueIds = append(uniqueIds, uniqueId)
	}

	return uniqueIds
}

func (r *Room) UpdateSubscribers(subscribers []database.SubscribeAccountModel) {
	r.clientMutex.Lock()
	defer r.clientMutex.Unlock()
	r.subscribers = subscribers
}

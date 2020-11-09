package server

import (
	"chats/database"
	"chats/models"
	"sync"
)

type Room struct {
	clients     map[string]bool
	subscribers []database.SubscribeUserModel
	clientMutex sync.Mutex
	roomId      uint
}

type RoomMessage struct {
	SendPush bool
	UserId   uint
	RoomId   uint
	Message  *models.WSChatResponse
}

func CreateRoom(roomId uint, subscribers []database.SubscribeUserModel) *Room {
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

func (r *Room) UpdateSubscribers(subscribers []database.SubscribeUserModel) {
	r.clientMutex.Lock()
	defer r.clientMutex.Unlock()
	r.subscribers = subscribers
}

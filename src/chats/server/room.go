package server

import (
	"chats/models"
	uuid "github.com/satori/go.uuid"
	"sync"
)

type Room struct {
	wsSessions  map[string]bool
	subscribers []models.AccountSubscriber
	mutex       sync.Mutex
	roomId      uuid.UUID
}

type RoomMessage struct {
	SendPush  bool
	AccountId uuid.UUID
	RoomId    uuid.UUID
	Message   *models.WSChatResponse
}

func InitRoom(roomId uuid.UUID, subscribers []models.AccountSubscriber) *Room {
	return &Room{
		wsSessions:  make(map[string]bool),
		subscribers: subscribers,
		roomId:      roomId,
	}
}

func (r *Room) removeSession(session *Session) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, ok := r.wsSessions[session.sessionId]; ok {
		delete(r.wsSessions, session.sessionId)
	}
}

func (r *Room) AddSession(sessionId string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.wsSessions[sessionId] = true
}

func (r *Room) getRoomSessionIds() []string {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	var sessionIds []string

	for sessionId, _ := range r.wsSessions {
		sessionIds = append(sessionIds, sessionId)
	}

	return sessionIds
}

func (r *Room) UpdateSubscribers(subscribers []models.AccountSubscriber) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.subscribers = subscribers
}

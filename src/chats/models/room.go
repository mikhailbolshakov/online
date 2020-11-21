package models

import (
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	"time"
)


type RoomSubscriber struct {
	Id            uuid.UUID
	RoomId        uuid.UUID  `gorm:"column:room_id"`
	AccountId     uuid.UUID  `gorm:"column:account_id"`
	Role          string     `gorm:"column:role"`
	SystemAccount uint8      `gorm:"column:system_account"`
	UnsubscribeAt *time.Time `gorm:"column:unsubscribe_at"`
	BaseModel
}

type Room struct {
	Id          uuid.UUID
	ReferenceId string `gorm:"column:reference_id"`
	Hash string `gorm:"column:hash"`
	Chat uint8 `gorm:"column:chat"`
	Audio uint8 `gorm:"column:audio"`
	Video uint8 `gorm:"column:video"`
	ClosedAt *time.Time `gorm:"column:closed_at"`
	Subscribers []RoomSubscriber
	BaseModel
}

type GetRoomCriteria struct {
	AccountId uuid.UUID
	ExternalAccountId string
	ReferenceId string
	RoomId uuid.UUID
	WithClosed bool
	WithSubscribers bool
}

type AccountSubscriber struct {
	RoomId uuid.UUID
	AccountId uuid.UUID
	SubscriberId uuid.UUID
	Role string
}

type ChatMessage struct {
	Id              uuid.UUID
	ClientMessageId string    `gorm:"column:client_message_id"`
	ChatId          uuid.UUID `gorm:"column:chat_id"`
	SubscribeId     uuid.UUID `gorm:"column:subscribe_id"`
	Type            string    `gorm:"column:type"`
	Message         string    `gorm:"column:message"`
	FileId          string    `gorm:"column:file_id"`
	Params          string    `gorm:"column:params"`
	BaseModel
}

type ChatMessageStatus struct {
	Id          uuid.UUID
	MessageId   uuid.UUID `gorm:"column:message_id"`
	SubscribeId uuid.UUID `gorm:"column:subscribe_id"`
	Status      string    `gorm:"column:status" sql:"not null;type:ENUM('recd', 'read');default:'recd';"`
	BaseModel
}

type ChatMessageParams struct {
	gorm.Model
	MessageId uuid.UUID
	Key       string
	Value     string
}

type ChatMessageHistory struct {
	AccountId   uuid.UUID `json:"account_id"`
	ChatId      uuid.UUID `json:"chat_id"`
	MessageId   uuid.UUID `json:"message_id"`
	NewMessages bool      `json:"new_messages"`
	UserType    string    `json:"user_type"`
	Count       int64     `json:"count"`
	Admin       bool      `json:"admin"`
	Search      string    `json:"search"`
	Date        string    `json:"date"`
	OnlyOneChat bool      `json:"only_one_chat"`
}

type FirstMessage struct {
	Id uuid.UUID `json:"id"`
}

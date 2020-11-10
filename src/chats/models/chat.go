package models

import (
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
)

type Chat struct {
	Id          uuid.UUID
	ReferenceId string `gorm:"column:reference_id"`
	Status      string `gorm:"column:status"`
	BaseModel
}

type ChatSubscribe struct {
	Id        uuid.UUID
	ChatId    uuid.UUID `gorm:"column:chat_id"`
	Active    uint8     `gorm:"column:active"`
	AccountId uuid.UUID `gorm:"column:account_id"`
	Role      string    `gorm:"column:role"`
	BaseModel
}

type ChatMessage struct {
	Id              uuid.UUID
	ClientMessageId string    `gorm:"column:client_message_id"`
	ChatId          uuid.UUID `gorm:"column:chat_id"`
	SubscribeId     uuid.UUID `gorm:"column:subscribe_id"`
	Type            string    `gorm:"column:type"`
	Message         string    `gorm:"column:message"`
	FileId          string    `gorm:"column:file_id"`
	BaseModel
}

type ChatMessageStatus struct {
	gorm.Model
	MessageId   uuid.UUID `gorm:"column:message_id"`
	SubscribeId uuid.UUID `gorm:"column:subscribe_id"`
	Status      string    `gorm:"column:status" sql:"not null;type:ENUM('recd', 'read');default:'recd';"`
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
	Count       uint16    `json:"count"`
	Admin       bool      `json:"admin"`
	Search      string    `json:"search"`
	Date        string    `json:"date"`
	OnlyOneChat bool      `json:"only_one_chat"`
}

type FirstMessage struct {
	Id uuid.UUID `json:"id"`
}

package models

import "github.com/jinzhu/gorm"

type Chat struct {
	gorm.Model
	OrderId uint   `gorm:"column:order_id"`
	Status  string `gorm:"column:status" sql:"not null;type:ENUM('opened', 'closed');default:'opened';"`
}

type ChatSubscribe struct {
	gorm.Model
	ChatId   uint   `gorm:"column:chat_id"`
	Active   uint8  `gorm:"column:active"`
	UserId   uint   `gorm:"column:user_id"`
	UserType string `gorm:"column:user_type" sql:"not null;type:ENUM('client', 'operator', 'doctor');default:'client';"`
}

type ChatMessage struct {
	gorm.Model
	ClientMessageId string `gorm:"column:client_message_id"`
	ChatId          uint   `gorm:"column:chat_id"`
	SubscribeId     uint   `gorm:"column:subscribe_id"`
	Type            string `gorm:"column:type"`
	Message         string `gorm:"column:message"`
	FileId          string `gorm:"column:file_id"`
}

type ChatMessageStatus struct {
	gorm.Model
	MessageId   uint   `gorm:"column:message_id"`
	SubscribeId uint   `gorm:"column:subscribe_id"`
	Status      string `gorm:"column:status" sql:"not null;type:ENUM('recd', 'read');default:'recd';"`
}

type ChatMessageParams struct {
	gorm.Model
	MessageId uint
	Key       string
	Value     string
}

type ChatMessageHistory struct {
	UserId      uint   `json:"user_id"`
	ChatId      uint   `json:"chat_id"`
	MessageId   uint   `json:"message_id"`
	NewMessages bool   `json:"new_messages"`
	UserType    string `json:"user_type"`
	Count       uint16 `json:"count"`
	Admin       bool   `json:"admin"`
	Search      string `json:"search"`
	Date        string `json:"date"`
	OnlyOneChat bool   `json:"only_one_chat"`
}

type FirstMessage struct {
	Id uint `json:"id"`
}

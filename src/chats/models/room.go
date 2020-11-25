package models

import (
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
	ReferenceId string     `gorm:"column:reference_id"`
	Hash        string     `gorm:"column:hash"`
	Chat        uint8      `gorm:"column:chat"`
	Audio       uint8      `gorm:"column:audio"`
	Video       uint8      `gorm:"column:video"`
	ClosedAt    *time.Time `gorm:"column:closed_at"`
	Subscribers []RoomSubscriber
	BaseModel
}

type GetRoomCriteria struct {
	AccountId         uuid.UUID
	ExternalAccountId string
	ReferenceId       string
	RoomId            uuid.UUID
	WithClosed        bool
	WithSubscribers   bool
}

type AccountSubscriber struct {
	RoomId       uuid.UUID
	AccountId    uuid.UUID
	SubscriberId uuid.UUID
	Role         string
}

type ChatOpponent struct {
	SubscriberId uuid.UUID
	AccountId    uuid.UUID
}

type ChatMessage struct {
	Id              uuid.UUID
	ClientMessageId string    `gorm:"column:client_message_id"`
	RoomId          uuid.UUID `gorm:"column:room_id"`
	SubscribeId     uuid.UUID `gorm:"column:subscribe_id"`
	AccountId       uuid.UUID `gorm:"column:account_id"`
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
	AccountId   uuid.UUID `gorm:"column:account_id"`
	Status      string    `gorm:"column:status" sql:"not null;type:ENUM('recd', 'read');default:'recd';"`
	BaseModel
}

type ChatMessageHistory struct {
	AccountId   uuid.UUID `json:"account_id"`
	RoomId      uuid.UUID `json:"room_id"`
	MessageId   uuid.UUID `json:"message_id"`
	NewMessages bool      `json:"new_messages"`
	Role        string    `json:"role"`
	Count       int64     `json:"count"`
	Admin       bool      `json:"admin"`
	Search      string    `json:"search"`
	Date        string    `json:"date"`
	OnlyOneChat bool      `json:"only_one_chat"`
}

type FirstMessage struct {
	Id uuid.UUID `json:"id"`
}

type SortRequest struct {
	Field string `json:"field"`
	// ask | desc
	Direction string `json:"direction"`
}

type PagingRequest struct {
	Size   int
	Index  int
	SortBy []SortRequest
}

type PagingResponse struct {
	Total int
	Index int
}

type GetMessageHistoryCriteria struct {
	AccountId         uuid.UUID
	AccountExternalId string
	ReferenceId       string
	RoomId            uuid.UUID
	Statuses          map[uuid.UUID]string
	CreatedBefore     *time.Time
	CreatedAfter      *time.Time
	WithStatuses      bool
	SentOnly          bool
	ReceivedOnly      bool
	WithAccounts      bool
}

type MessageStatus struct {
	AccountId  uuid.UUID
	Status     string
	StatusDate time.Time
}

type MessageAccount struct {
	Id         uuid.UUID
	Type       string `gorm:"column:account_type"`
	Status     string `gorm:"column:status"`
	Account    string `gorm:"column:account"`
	ExternalId string `gorm:"column:external_id"`
	FirstName  string `gorm:"column:first_name"`
	MiddleName string `gorm:"column:middle_name"`
	LastName   string `gorm:"column:last_name"`
	Email      string `gorm:"column:email"`
	Phone      string `gorm:"column:phone"`
	AvatarUrl  string `gorm:"column:avatar_url"`
}

type MessageHistoryItem struct {
	Id               uuid.UUID
	ClientMessageId  string
	ReferenceId      string
	RoomId           uuid.UUID
	Type             string
	Message          string
	FileId           string
	Params           map[string]string
	SenderAccountId  uuid.UUID
	SenderExternalId string
	Statuses         []MessageStatus
}

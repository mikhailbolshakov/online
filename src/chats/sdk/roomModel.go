package sdk

import (
	uuid "github.com/satori/go.uuid"
	"time"
)

type Room struct {
	Id          uuid.UUID `json:"id"`
	ReferenceId string    `json:"referenceId"`
	Hash        string    `json:"hash"`
	Chat        bool      `json:"chat"`
	Video       bool      `json:"video"`
	Audio       bool      `json:"audio"`
}

type SubscriberRequest struct {
	Account *AccountIdRequest `json:"account"`
	Role    string            `json:"role"`
}

type RoomRequest struct {
	ReferenceId string              `json:"referenceId"`
	Chat        bool                `json:"chat"`
	Video       bool                `json:"video"`
	Audio       bool                `json:"audio"`
	Subscribers []SubscriberRequest `json:"subscribers"`
}

type CreateRoomRequest struct {
	Room *RoomRequest
}

type RoomResponse struct {
	Id   uuid.UUID `json:"id"`
	Hash string    `json:"hash"`
}

type CreateRoomResponse struct {
	Result *RoomResponse
	Errors []ErrorResponse
}

type GetSubscriberResponse struct {
	Id            uuid.UUID
	AccountId     uuid.UUID
	Role          string
	UnSubscribeAt *time.Time
}

type GetRoomResponse struct {
	Id          uuid.UUID
	Hash        string
	ReferenceId string
	Chat        bool
	Video       bool
	Audio       bool
	ClosedAt    *time.Time
	Subscribers []GetSubscriberResponse
}

type GetRoomsByCriteriaRequest struct {
	ReferenceId     string
	AccountId       *AccountIdRequest
	RoomId          uuid.UUID
	WithClosed      bool
	WithSubscribers bool
}

type GetRoomsByCriteriaResponse struct {
	Rooms  []GetRoomResponse
	Errors []ErrorResponse
}

type RoomSubscribeRequest struct {
	RoomId      uuid.UUID           `json:"roomId"`
	ReferenceId string              `json:"referenceId"`
	Subscribers []SubscriberRequest `json:"subscribers"`
}

type RoomSubscribeResponse struct {
	Errors []ErrorResponse
}

type RoomUnsubscribeRequest struct {
	RoomId      uuid.UUID        `json:"roomId"`
	ReferenceId string           `json:"referenceId"`
	AccountId   AccountIdRequest `json:"accountId"`
}

type RoomUnsubscribeResponse struct {
	Errors []ErrorResponse
}

type RoomMessageAccountSubscribeRequest struct {
	AccountId uuid.UUID `json:"accountId"`
	RoomId    uuid.UUID `json:"roomId"`
	Role      string    `json:"role"`
}

type RoomMessageAccountUnsubscribeRequest struct {
	AccountId uuid.UUID `json:"accountId"`
	RoomId    uuid.UUID `json:"roomId"`
}

type CloseRoomRequest struct {
	RoomId      uuid.UUID `json:"roomId"`
	ReferenceId string    `json:"referenceId"`
}

type CloseRoomResponse struct {
	Errors []ErrorResponse
}

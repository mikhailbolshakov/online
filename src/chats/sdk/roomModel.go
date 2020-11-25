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
	Room *RoomRequest `json:"room"`
}

type RoomResponse struct {
	Id   uuid.UUID `json:"id"`
	Hash string    `json:"hash"`
}

type CreateRoomResponse struct {
	Result *RoomResponse   `json:"result"`
	Errors []ErrorResponse `json:"errors"`
}

type GetSubscriberResponse struct {
	Id            uuid.UUID  `json:"id"`
	AccountId     uuid.UUID  `json:"accountId"`
	Role          string     `json:"role"`
	UnSubscribeAt *time.Time `json:"unsubscribeAt"`
}

type GetRoomResponse struct {
	Id          uuid.UUID               `json:"id"`
	Hash        string                  `json:"hash"`
	ReferenceId string                  `json:"referenceId"`
	Chat        bool                    `json:"chat"`
	Video       bool                    `json:"video"`
	Audio       bool                    `json:"audio"`
	ClosedAt    *time.Time              `json:"closedAt"`
	Subscribers []GetSubscriberResponse `json:"subscribers"`
}

type GetRoomsByCriteriaRequest struct {
	ReferenceId     string            `json:"referenceId"`
	AccountId       *AccountIdRequest `json:"accountId"`
	RoomId          uuid.UUID         `json:"roomId"`
	WithClosed      bool              `json:"withClosed"`
	WithSubscribers bool              `json:"withSubscribers"`
}

type GetRoomsByCriteriaResponse struct {
	Rooms  []GetRoomResponse `json:"rooms"`
	Errors []ErrorResponse   `json:"errors"`
}

type RoomSubscribeRequest struct {
	RoomId      uuid.UUID           `json:"roomId"`
	ReferenceId string              `json:"referenceId"`
	Subscribers []SubscriberRequest `json:"subscribers"`
}

type RoomSubscribeResponse struct {
	Errors []ErrorResponse `json:"errors"`
}

type RoomUnsubscribeRequest struct {
	RoomId      uuid.UUID        `json:"roomId"`
	ReferenceId string           `json:"referenceId"`
	AccountId   AccountIdRequest `json:"accountId"`
}

type RoomUnsubscribeResponse struct {
	Errors []ErrorResponse `json:"errors"`
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
	Errors []ErrorResponse `json:"errors"`
}

type SortRequest struct {
	Field string `json:"field"`
	// ask | desc
	Direction string `json:"direction"`
}

type PagingRequest struct {
	Size   int           `json:"pageSize"`
	Index  int           `json:"pageIndex"`
	SortBy []SortRequest `json:"sortBy"`
}

type PagingResponse struct {
	Total int `json:"pages"`
	Index int `json:"index"`
}

type GetMessageHistoryCriteria struct {
	// messages for all rooms on which given account has been subscribed (even though it had been unsubscribed then)
	AccountId AccountIdRequest `json:"accountId"`
	// messages for all rooms related to given reference
	ReferenceId string `json:"referenceId"`
	// messages for the given room
	RoomId uuid.UUID `json:"roomId"`
	// messages where given opponent has given status map[accountId]status
	Statuses map[uuid.UUID]string `json:"accountStatuses"`
	// messages created before time
	CreatedBefore *time.Time `json:"createdBefore"`
	// messages created after time
	CreatedAfter *time.Time `json:"createdAfter"`
	// add statuses to response (empty otherwise)
	WithStatuses bool `json:"withStatuses"`
	// add accounts info to response
	WithAccounts bool `json:"withAccounts"`
	// if true retrieves only messages sent by the given account (works together with AccountId)
	SentOnly bool `json:"sentOnly"`
	// if true retrieves only messages received by the given account (works together with AccountId)
	// mutual exclusion with SentOnly (SentOnly == true && ReceivedOnly == true -> invalid criteria combination)
	ReceivedOnly bool `json:"receivedOnly"`
}

type GetMessageHistoryRequest struct {
	PagingRequest *PagingRequest             `json:"pagingRequest"`
	Criteria      *GetMessageHistoryCriteria `json:"criteria"`
}

type MessageAccount struct {
	Id         uuid.UUID `json:"id"`
	Type       string    `json:"type"`
	Status     string    `json:"status"`
	Account    string    `json:"account"`
	ExternalId string    `json:"externalId"`
	FirstName  string    `json:"firstName"`
	MiddleName string    `json:"middleName"`
	LastName   string    `json:"lastName"`
	Email      string    `json:"email"`
	Phone      string    `json:"phone"`
	AvatarUrl  string    `json:"avatarUrl"`
}

type GetMessageHistoryResponse struct {
	Messages []MessageHistoryItem `json:"messages"`
	Accounts []MessageAccount     `json:"accounts"`
	Paging   *PagingResponse      `json:"paging"`
	Errors   []ErrorResponse      `json:"errors"`
}

type MessageStatus struct {
	AccountId  uuid.UUID
	Status     string
	StatusDate time.Time
}

type MessageHistoryItem struct {
	Id              uuid.UUID         `json:"id"`
	ClientMessageId string            `json:"clientMessageId"`
	ReferenceId     string            `json:"referenceId"`
	RoomId          uuid.UUID         `json:"roomId"`
	Type            string            `json:"type"`
	Message         string            `json:"message"`
	FileId          string            `json:"fileId"`
	Params          map[string]string `json:"params"`
	// Account Id of the account who has sent this message
	SenderAccountId uuid.UUID `json:"senderAccountId"`
	// External Id of the account who has sent this message
	SenderExternalId string `json:"senderExternalId"`
	// Message statuses for all room's accounts map[accountId]status
	Statuses []MessageStatus `json:"statuses"`
}

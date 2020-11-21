package models

import (
	"chats/sdk"
	uuid "github.com/satori/go.uuid"
)


// TODO: remove relations to sdk

//	request default
type WSChatRequest struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

//	response 200
type WSChatResponse struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type WSChatPushResponse struct {
	WSChatResponse
	Recipient string `json:"recipient"`
}

//	response 500
type WSChatErrorResponse struct {
	Error WSChatErrorErrorResponse `json:"sentry"`
}
type WSChatErrorErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

//	message request
type WSChatMessagesRequest struct {
	AccountId uuid.UUID                 `json:"account_id"`
	Type      string                    `json:"type"`
	Data      WSChatMessagesDataRequest `json:"data"`
}

type WSChatMessagesDataRequest struct {
	Messages []WSChatMessageDataRequest `json:"messages"`
}

type WSChatMessageDataRequest struct {
	ClientMessageId string            `json:"clientMessageId"`
	RoomId          uuid.UUID         `json:"roomId"`
	Type            string            `json:"type"`
	Text            string            `json:"text"`
	Params          map[string]string `json:"params"`
}

//	message response
type WSChatMessagesDataResponse struct {
	Messages []interface{} `json:"messages"`
	Accounts []sdk.Account `json:"accounts"`
}
type WSChatMessagesDataMessageResponse struct {
	Id              uuid.UUID         `json:"id"`
	ClientMessageId string            `json:"clientMessageId"`
	InsertDate      string            `json:"insertDate"`
	ChatId          uuid.UUID         `json:"chatId"`
	AccountId       uuid.UUID         `json:"accountId"`
	Sender          string            `json:"sender"`
	Status          string            `json:"status"`
	Type            string            `json:"type"`
	Text            string            `json:"text"`
	Params          map[string]string `json:"params"`
}
type WSChatMessagesDataMessageFileResponse struct {
	WSChatMessagesDataMessageResponse
	File sdk.FileModel `json:"file"`
}

//	messageStatus request
type WSChatMessageStatusRequest struct {
	Type string                         `json:"type"`
	Data WSChatMessageStatusDataRequest `json:"data"`
}
type WSChatMessageStatusDataRequest struct {
	Status    string    `json:"status"`
	ChatId    uuid.UUID `json:"chatId"`
	MessageId uuid.UUID `json:"messageId"`
}

//	messageStatus response
type WSChatMessageStatusResponse struct {
	Type string                          `json:"type"`
	Data WSChatMessageStatusDataResponse `json:"data"`
}
type WSChatMessageStatusDataResponse struct {
	Status    string    `json:"status"`
	ChatId    uuid.UUID `json:"chatId"`
	MessageId uuid.UUID `json:"messageId"`
}

//	opponentStatus request
type WSChatOpponentStatusRequest struct {
	Type string                          `json:"type"`
	Data WSChatOpponentStatusDataRequest `json:"data"`
}
type WSChatOpponentStatusDataRequest struct {
	RoomId uuid.UUID `json:"roomId"`
}

//	opponentStatus response
type WSChatOpponentStatusResponse struct {
	Type string                           `json:"type"`
	Data WSChatOpponentStatusDataResponse `json:"data"`
}
type WSChatOpponentStatusDataResponse struct {
	RoomId   uuid.UUID            `json:"roomId"`
	Accounts []AccountStatusModel `json:"accounts"`
}

//	join request without response
type WSChatJoinRequest struct {
	Type string                `json:"type"`
	Data WSChatJoinDataRequest `json:"data"`
}

type WSChatJoinDataRequest struct {
	ConsultationId string `json:"consultationId"`
}

//	typing request
type WSChatTypingRequest struct {
	Type string                  `json:"type"`
	Data WSChatTypingDataRequest `json:"data"`
}
type WSChatTypingDataRequest struct {
	RoomId uuid.UUID `json:"roomId"`
	Status string    `json:"status"`
}

//	typing request
type WSChatEchoRequest struct {
	Type string   `json:"type"`
	Data struct{} `json:"data"`
}

//	typing request
type WSChatEchoResponse struct {
}

//	typing response
type WSChatTypingDataResponse struct {
	AccountId uuid.UUID `json:"accountId"`
	Message   string    `json:"message"`
	Status    string    `json:"status"`
}

//	anyMessageToClient from nats [response only]
type WSMessageToMobileClientResponse struct {
	Type string                              `json:"type"`
	Data WSMessageToMobileClientDataResponse `json:"data"`
}
type WSMessageToMobileClientDataResponse struct {
	AccountId uuid.UUID         `json:"accountId"`
	Type      string            `json:"type"`
	Data      map[string]string `json:"data"`
}

//	system user
type WSSystemUserRequest struct {
	SendPush  bool      `json:"send_push"`
	AccountId uuid.UUID `json:"account_id"`
	RoomId    uint      `json:"room_id"`
}

//	system subscribe user
type WSSystemUserSubscribeRequest struct {
	WSSystemUserRequest
	Message WSSystemUserSubscribeRequestMessage `json:"message"`
}
type WSSystemUserSubscribeRequestMessage struct {
	Type string                                 `json:"type"`
	Data sdk.RoomMessageAccountSubscribeRequest `json:"data"`
}

//system unsubscribe user
type WSSystemUserUnsubscribeRequest struct {
	WSSystemUserRequest
	Message WSSystemUserUnsubscribeRequestMessage `json:"message"`
}

type WSSystemUserUnsubscribeRequestMessage struct {
	Type string                                `json:"type"`
	Data sdk.RoomMessageAccountUnsubscribeRequest `json:"data"`
}

//	other models

type ExpandedAccountModel struct {
	sdk.Account
	Role string `json:"role"`
}

type AccountStatusModel struct {
	AccountId uuid.UUID `json:"accountId"`
	Status    string    `json:"status"`
}

type CronMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type CronSendOnlineUsers struct {
	Type string                     `json:"type"`
	Data CronSendOnlineAccountsData `json:"data"`
}

type CronSendOnlineAccountsData struct {
	Accounts []uuid.UUID `json:"accounts"`
}

type CronGetUserStatusRequest struct {
	Type string                          `json:"type"`
	Data CronGetAccountStatusRequestData `json:"data"`
}

type CronGetAccountStatusRequestData struct {
	AccountId uuid.UUID `json:"account_id"`
}

type CronGetAccountStatusResponse struct {
	Online bool `json:"online"`
}

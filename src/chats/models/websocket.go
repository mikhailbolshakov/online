package models

import "gitlab.medzdrav.ru/health-service/go-sdk"

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
	UserId uint                      `json:"user_id"`
	Type   string                    `json:"type"`
	Data   WSChatMessagesDataRequest `json:"data"`
}

type WSChatMessagesDataRequest struct {
	Messages []WSChatMessageDataRequest `json:"messages"`
}

type WSChatMessageDataRequest struct {
	ClientMessageId string            `json:"clientMessageId"`
	ChatId          uint              `json:"chatId"`
	Type            string            `json:"type"`
	Text            string            `json:"text"`
	Params          map[string]string `json:"params"`
}

//	message response
type WSChatMessagesDataResponse struct {
	Messages []interface{}   `json:"messages"`
	Users    []sdk.UserModel `json:"users"`
}
type WSChatMessagesDataMessageResponse struct {
	Id              uint              `json:"id"`
	ClientMessageId string            `json:"clientMessageId"`
	InsertDate      string            `json:"insertDate"`
	ChatId          uint              `json:"chatId"`
	UserId          uint              `json:"userId"`
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
	Status    string `json:"status"`
	ChatId    uint   `json:"chatId"`
	MessageId uint   `json:"messageId"`
}

//	messageStatus response
type WSChatMessageStatusResponse struct {
	Type string                          `json:"type"`
	Data WSChatMessageStatusDataResponse `json:"data"`
}
type WSChatMessageStatusDataResponse struct {
	Status    string `json:"status"`
	ChatId    uint   `json:"chatId"`
	MessageId uint   `json:"messageId"`
}

//	opponentStatus request
type WSChatOpponentStatusRequest struct {
	Type string                          `json:"type"`
	Data WSChatOpponentStatusDataRequest `json:"data"`
}
type WSChatOpponentStatusDataRequest struct {
	ChatId uint `json:"chatId"`
}

//	opponentStatus response
type WSChatOpponentStatusResponse struct {
	Type string                           `json:"type"`
	Data WSChatOpponentStatusDataResponse `json:"data"`
}
type WSChatOpponentStatusDataResponse struct {
	ChatId uint              `json:"chatId"`
	Users  []UserStatusModel `json:"users"`
}

//	join request without response
type WSChatJoinRequest struct {
	Type string                `json:"type"`
	Data WSChatJoinDataRequest `json:"data"`
}

type WSChatJoinDataRequest struct {
	ConsultationId uint `json:"consultationId"`
}

//	typing request
type WSChatTypingRequest struct {
	Type string                  `json:"type"`
	Data WSChatTypingDataRequest `json:"data"`
}
type WSChatTypingDataRequest struct {
	ChatId uint   `json:"chatId"`
	Status string `json:"status"`
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
	UserId  uint   `json:"userId"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

//	anyMessageToClient from nats [response only]
type WSMessageToMobileClientResponse struct {
	Type string                              `json:"type"`
	Data WSMessageToMobileClientDataResponse `json:"data"`
}
type WSMessageToMobileClientDataResponse struct {
	UserId uint              `json:"userId"`
	Type   string            `json:"type"`
	Data   map[string]string `json:"data"`
}

//	system user
type WSSystemUserRequest struct {
	SendPush bool `json:"send_push"`
	UserId   uint `json:"user_id"`
	RoomId   uint `json:"room_id"`
}

//	system subscribe user
type WSSystemUserSubscribeRequest struct {
	WSSystemUserRequest
	Message WSSystemUserSubscribeRequestMessage `json:"message"`
}
type WSSystemUserSubscribeRequestMessage struct {
	Type string                           `json:"type"`
	Data sdk.ChatUserSubscribeRequestBody `json:"data"`
}

//	Deprecated: struct is deprecated
type WSSystemUserSubscribeRequestMessageData struct {
	UserId   uint   `json:"user_id"`
	ChatId   uint   `json:"chat_id"`
	UserType string `json:"user_type"`
}

//system unsubscribe user
type WSSystemUserUnsubscribeRequest struct {
	WSSystemUserRequest
	Message WSSystemUserUnsubscribeRequestMessage `json:"message"`
}

type WSSystemUserUnsubscribeRequestMessage struct {
	Type string                             `json:"type"`
	Data sdk.ChatUserUnsubscribeRequestBody `json:"data"`
}

//	other models

type ExpandedUserModel struct {
	sdk.UserModel
	UserType string `json:"user_type"`
}

type UserStatusModel struct {
	UserId uint   `json:"userId"`
	Status string `json:"status"`
}

type CronMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type CronSendOnlineUsers struct {
	Type string                  `json:"type"`
	Data CronSendOnlineUsersData `json:"data"`
}

type CronSendOnlineUsersData struct {
	Users []uint `json:"users"`
}

type CronGetUserStatusRequest struct {
	Type string                       `json:"type"`
	Data CronGetUserStatusRequestData `json:"data"`
}

type CronGetUserStatusRequestData struct {
	UserId uint `json:"user_id"`
}

type CronGetUserStatusResponse struct {
	Online bool `json:"online"`
}

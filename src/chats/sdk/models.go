package sdk

import (
	uuid "github.com/satori/go.uuid"
	"time"
)

//	request
type ApiRequest struct {
	Method string      `json:"method"`
	Path   string      `json:"path"`
	Body   interface{} `json:"body"`
}
type ApiRequestModel struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

// api sentry response
type ApiErrorResponse struct {
	Error ApiErrorResponseError `json:"sentry"`
}
type ApiErrorResponseError struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// cron sentry response
type CronErrorResponse struct {
	Error CronErrorResponseError
}
type CronErrorResponseError ApiErrorResponseError

//	user
type ApiAccountResponse struct {
	Data ApiAccountDataAccountResponse `json:"data"`
}
type ApiUserDataResponse struct {
	User ApiAccountDataAccountResponse `json:"user"`
}
type ApiAccountDataAccountResponse struct {
	Id         uuid.UUID `json:"id"`
	FirstName  string    `json:"first_name"`
	MiddleName string    `json:"middle_name"`
	LastName   string    `json:"last_name"`
	Photo      string    `json:"photo"`
}

//	storage
type ApiFileRequest struct {
	FileId    string    `json:"file_id"`
	ChatId    uuid.UUID `json:"chat_id"`
	AccountId uuid.UUID `json:"account_id"`
}

type ApiFileResponse struct {
	Data FileModel `json:"data"`
}

type ApiFileResponseData struct {
	Id string `json:"id"`
}

type ChatNewResponse struct {
	Data ChatNewResponseData `json:"data"`
}
type ChatNewResponseData struct {
	ChatId uuid.UUID `json:"chat_id"`
}

//	Подписать пользователя [POST] /chats/user/subscribe
type ChatAccountSubscribeRequest struct {
	ApiRequestModel
	Body RoomMessageAccountSubscribeRequest `json:"body"`
}


type ChatAccountSubscribeResponse struct {
	Data ChatAccountSubscribeResponseData `json:"data"`
}
type ChatAccountSubscribeResponseData struct {
	Result bool `json:"result"`
}

//	Изменение userId у подписчика на чат [PUT] /chats/user/subscribe
type ChatAccountSubscribeChangeRequest struct {
	ApiRequestModel
	Body ChatAccountSubscribeChangeRequestBody `json:"body"`
}
type ChatAccountSubscribeChangeRequestBody struct {
	ChatId       uuid.UUID `json:"chat_id"`
	OldAccountId uuid.UUID `json:"old_account_id"`
	NewAccountId uuid.UUID `json:"new_account_id"`
}

type ChatAccountSubscribeChangeResponse struct {
	Data BoolResponseData `json:"data"`
}

//	Добавить сообщение [POST] /chats/message
type ChatMessageRequest struct {
	ApiRequestModel
	Body ChatMessageRequestBody `json:"body"`
}
type ChatMessageRequestBody struct {
	ChatId          uuid.UUID         `json:"chat_id"`
	AccountId       uuid.UUID         `json:"account_id"`
	Message         string            `json:"message"`
	ClientMessageId string            `json:"client_message_id"`
	Type            string            `json:"type"`
	Params          map[string]string `json:"params"`
	FileId          string            `json:"file_id"`
}

type ChatMessageResponse struct {
	Data ChatMessageResponseData `json:"data"`
}
type ChatMessageResponseData struct {
	Id              uuid.UUID         `json:"id"`
	ClientMessageId string            `json:"client_message_id"`
	InsertDate      string            `json:"insert_date"`
	ChatId          uuid.UUID         `json:"chat_id"`
	AccountId       uuid.UUID         `json:"account_id"`
	Sender          string            `json:"sender"`
	Status          string            `json:"status"`
	Type            string            `json:"type"`
	Text            string            `json:"text"`
	File            interface{}       `json:"file"`
	Params          map[string]string `json:"params"`
}

type ChatStatusResponse struct {
	Data ChatStatusDataResponse `json:"data"`
}
type ChatStatusDataResponse struct {
	Result bool `json:"result"`
}

//	Получить список чатов пользователя [GET] /chats/chats
type ChatListRequest struct {
	ApiRequestModel
	Body ChatListRequestBody `json:"body"`
}
type ChatListRequestBody struct {
	Account AccountIdRequest `json:"account"`
	Count   int              `json:"count"`
}

type ChatListResponse struct {
	Data []ChatListResponseDataItem `json:"data"`
}

type ChatLastResponse struct {
	Data ChatListResponseDataItem `json:"data"`
}

type ChatsLastRequestBody struct {
	AccountId uuid.UUID `json:"account_id"`
}

//	Получить информацию о последнем чате
type ChatsLastRequest struct {
	ApiRequestModel
	Body ChatsLastRequestBody `json:"body"`
}

type ChatListResponseDataItem struct {
	ChatListDataItem
	Opponent interface{} `json:"opponent"`
}

type ChatListDataItem struct {
	Id             uuid.UUID `json:"id"`
	Status         string    `json:"status"`
	ReferenceId    string    `json:"reference_id"`
	InsertDate     string    `json:"insert_date"`
	LastUpdateDate string    `json:"last_update_date"`
}

//	Получить информацию о списке чатов [GET] /chats/info
type ChatsInfoRequest struct {
	ApiRequestModel
	Body ChatsInfoRequestBody `json:"body"`
}
type ChatsInfoRequestBody struct {
	Account AccountIdRequest `json:"account"`
	ChatsId []uuid.UUID      `json:"chats_id"`
}

type ChatsInfoResponse struct {
	Data []ChatListResponseDataItem `json:"data"`
}

//	Получить список сообщений и информацию о чате [GET] /chats/messages
type ChatMessagesRequest struct {
	ApiRequestModel
	Body ChatMessagesRequestBody `json:"body"`
}
type ChatMessagesRequestBody struct {
	ChatId      uuid.UUID `json:"chat_id"`      // Обязательное поле
	MessageId   uuid.UUID `json:"message_id"`   // id последнего сообщения (не обязательное)
	NewMessages bool      `json:"new_messages"` // true - получить сообщения от MessageId, false - до MessageId
	Role        string    `json:"role"`         // роль юзера
	Search      string    `json:"search"`
	Date        string    `json:"date"`
	Count       int64    `json:"count"` // количество сообщений (не обязательное)
	Admin       bool      `json:"admin"`
	OnlyOneChat bool      `json:"only_one_chat"`
}

type ChatMessagesResponse struct {
	Data []ChatMessagesResponseDataItem `json:"data"`
}
type ChatMessagesResponseDataItem struct {
	ChatMessagesResponseDataItemTmp
	InsertDate string `json:"insert_date"`
}

type ChatMessagesResponseDataItemDb struct {
	ChatMessagesResponseDataItemTmp
	InsertDate time.Time `json:"insert_date"`
}

type ChatMessagesResponseDataItemTmp struct {
	Id              uuid.UUID         `json:"id"`
	ClientMessageId string            `json:"client_message_id"`
	ChatId          uuid.UUID         `json:"chat_id"`
	Status          string            `json:"status"`
	Type            string            `json:"type"`
	Text            string            `json:"text"`
	FileId          string            `json:"file_id"`
	Params          map[string]string `json:"params"`
	Sender          string            `json:"sender"`
	AccountId       uuid.UUID         `json:"account_id"`
}

//	Список сообщений для клиента [GET] /chats/chat/recent
type ChatMessagesRecentRequest struct {
	ApiRequestModel
	Body ChatMessagesRecentRequestBody `json:"body"`
}
type ChatMessagesRecentRequestBody struct {
	AccountId uuid.UUID `json:"account_id"`
	ChatMessagesRequestBody
}

type ChatMessagesRecentResponse struct {
	Data ChatMessagesRecentResponseData `json:"data"`
}
type ChatMessagesRecentResponseData struct {
	Messages []ChatMessagesResponseDataItem `json:"messages"`
	Accounts []Account                      `json:"accounts"`
}

//	Список истории сообщений для клиента [GET] /chats/chat/hictory (по сути тож самое что и /chats/chat/recent)
type ChatMessagesHistoryRequest struct {
	ApiRequestModel
	Body ChatMessagesRecentRequestBody `json:"body"`
}
type ChatMessagesHistoryResponse struct {
	Data ChatMessagesRecentResponseData `json:"data"`
}

//	Сообщение в веб-сокет для клиента
type MessageToMobileClientRequest struct {
	ApiRequestModel
	Body MessageToMobileClientRequestBody `json:"body"`
}
type MessageToMobileClientRequestBody struct {
	AccountId uuid.UUID         `json:"account_id"`
	Type      string            `json:"type"`
	Data      map[string]string `json:"data"`
}

type MessageToMobileClientResponse struct {
	Data MessageToMobileClientResponseData `json:"data"`
}
type MessageToMobileClientResponseData struct {
	Result bool `json:"result"`
}

type BoolResponseData struct {
	Result bool `json:"result"`
}

type FileModel struct {
	Id         string `json:"id"`
	Title      string `json:"title"`
	Url        string `json:"url"`
	Thumbnail  string `json:"thumbnail"`
	MimeType   string `json:"mimeType"`
	InsertDate string `json:"insertDate"`
	Size       int64  `json:"size"`
}


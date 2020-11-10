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
	Id         uuid.UUID   `json:"id"`
	FirstName  string `json:"first_name"`
	MiddleName string `json:"middle_name"`
	LastName   string `json:"last_name"`
	Photo      string `json:"photo"`
}

//	personal
type ApiPersonalResponse struct {
	Data ApiPersonalResponseData `json:"data"`
}
type ApiPersonalResponseData struct {
	AccountId  uuid.UUID `json:"account_id"`
	FirstName  string    `json:"first_name"`
	MiddleName string    `json:"middle_name"`
	LastName   string    `json:"last_name"`
	Photo      string    `json:"photo"`
}

//	doctors
type ApiDoctorSpecializationRespones struct {
	Data ApiDoctorSpecializationResponesData `json:"data"`
}
type ApiDoctorSpecializationResponesData struct {
	Id    uuid.UUID   `json:"id"`
	Title string `json:"title"`
}

//	push

// Deprecated: ApiUserPushTokenResponse is deprecated.
type ApiUserPushTokenResponse struct {
	Data ApiUserPushTokenResponseData `json:"data"`
}

// Deprecated: ApiUserPushTokenResponseData is deprecated.
type ApiUserPushTokenResponseData struct {
	Ios     string `json:"ios"`
	Android string `json:"android"`
}

// Deprecated: ApiUserPushResponse is deprecated.
type ApiUserPushResponse struct {
	Type      string      `json:"type"`
	Recipient string      `json:"recipient"`
	Data      interface{} `json:"data"`
}

type RecipientWithOrder struct {
	AccountId      uuid.UUID `json:"user_id"`
	ConsultationId string      `json:"consultation_id"`
}

type ApiUserPushRequest struct {
	Message ApiUserPushRequestMessage `json:"message"`
	//	uint / []uint userId
	Recipient interface{} `json:"recipient"`
}
type ApiUserPushRequestMessage struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

//	storage
type ApiFileRequest struct {
	FileId    string `json:"file_id"`
	ChatId    uuid.UUID   `json:"chat_id"`
	AccountId uuid.UUID   `json:"account_id"`
}
type ApiFileResponse struct {
	Data FileModel `json:"data"`
}
type ApiFileResponseData struct {
	Id string `json:"id"`
}

//	telemed
type ApiTelemedRequest struct {
	AccountId uuid.UUID `json:"user_id"`
	Online    bool      `json:"online"`
}

//	Создать чат [POST] /chats/new
type ChatNewRequest struct {
	ApiRequestModel
	Body ChatNewRequestBody `json:"body"`
}
type ChatNewRequestBody struct {
	ReferenceId string `json:"reference_id"`
}

//	Создать чат и пописать [POST] /chats/new/subscribe
type ChatNewSubscribeRequest struct {
	ApiRequestModel
	Body ChatNewSubscribeRequestBody `json:"body"`
}
type ChatNewSubscribeRequestBody struct {
	ReferenceId string          `json:"reference_id"`
	Account     *AccountRequest `json:"account"`
	UserType    string          `json:"user_type"`
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
	Body ChatAccountSubscribeRequestBody `json:"body"`
}
type ChatAccountSubscribeRequestBody struct {
	Account *AccountRequest `json:"account"`
	ChatId  uuid.UUID       `json:"chat_id"`
}

type ChatAccountSubscribeResponse struct {
	Data ChatAccountSubscribeResponseData `json:"data"`
}
type ChatAccountSubscribeResponseData struct {
	Result bool `json:"result"`
}

//	Отписать пользователя [POST] /chats/user/unsubscribe
type ChatUserUnsubscribeRequest struct {
	ApiRequestModel
	Body ChatAccountUnsubscribeRequestBody `json:"body"`
}
type ChatAccountUnsubscribeRequestBody struct {
	AccountId uuid.UUID `json:"account_id"`
	ChatId    uuid.UUID `json:"chat_id"`
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

//	Изменить статус чата на "закрыт"
type ChatStatusRequest struct {
	ApiRequestModel
	Body ChatStatusBodyRequest `json:"body"`
}

type ChatStatusBodyRequest struct {
	ChatId uuid.UUID `json:"chat_id"`
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
	AccountId uuid.UUID `json:"account_id"`
	Count     uint16    `json:"count"`
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

//	Получить информацию о чате [GET] /chats/chat
type ChatInfoRequest struct {
	ApiRequestModel
	Body ChatInfoRequestBody `json:"body"`
}
type ChatInfoRequestBody struct {
	ChatId    uuid.UUID `json:"chat_id"`
	AccountId uuid.UUID `json:"account_id"`
}

//	Получить последний чат по заявке для пользователя [GET] /order/chat
type RefereneChatRequest struct {
	ApiRequestModel
	Body []ReferenceChatRequestBodyItem `json:"body"`
}
type ReferenceChatRequestBodyItem struct {
	ReferenceId uuid.UUID `json:"reference_id"`
	OpponentId  uuid.UUID `json:"opponent_id"`
}

type ChatInfoResponse struct {
	Data ChatListResponseDataItem `json:"data"`
}

//	Получить информацию о списке чатов [GET] /chats/info
type ChatsInfoRequest struct {
	ApiRequestModel
	Body ChatsInfoRequestBody `json:"body"`
}
type ChatsInfoRequestBody struct {
	AccountId uuid.UUID   `json:"account_id"`
	ChatsId   []uuid.UUID `json:"chats_id"`
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
	Role        string    `json:"role"`    // роль юзера
	Search      string    `json:"search"`
	Date        string    `json:"date"`
	Count       uint16    `json:"count"` // количество сообщений (не обязательное)
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
	Accounts []AccountModel                 `json:"accounts"`
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

type ClientConsultationUpdateRequest struct {
	ApiRequestModel
	Body ClientConsultationUpdateRequestBody `json:"body"`
}
type ClientConsultationUpdateRequestBody struct {
	AccountId uuid.UUID                               `json:"account_id"`
	Data      ClientConsultationUpdateRequestBodyData `json:"data"`
}
type ClientConsultationUpdateRequestBodyData struct {
	Active         bool `json:"active"`
	ConsultationId uint `json:"consultationId"`
}

//	other models
type AccountModel struct {
	Id         uuid.UUID   `json:"id"`
	FirstName  string `json:"firstName"`
	LastName   string `json:"lastName"`
	MiddleName string `json:"middleName"`
	Photo      string `json:"photo"`
}

type UserTokenModel struct {
	AccessToken string `json:"access_token"`
}

type AccountIdModel struct {
	Id uuid.UUID `json:"id"`
}

type BoolResponseData struct {
	Result bool `json:"result"`
}

type ApiPersonalDataResponse struct {
	AccountId  uuid.UUID `json:"account_id"`
	FirstName  string    `json:"first_name"`
	MiddleName string    `json:"middle_name"`
	LastName   string    `json:"last_name"`
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

type ConsultationResponseResponse struct {
	Data ConsultationResponseResponseData `json:"data"`
}
type ConsultationResponseResponseData struct {
	PatientId uuid.UUID `json:"patient_id"`
	AccountId uuid.UUID `json:"account_id"`
}

type AccountRequest struct {
	AccountId  uuid.UUID   `json:"account_id"`
	ExternalId string `json:"external_id"`
}

type CreateAccountRequest struct {
	ApiRequestModel
	Body CreateAccountRequestBody `json:"body"`
}

type CreateAccountRequestBody struct {
	Account    string `json:"account"`
	Type       string `json:"type"`
	ExternalId string `json:"externalId"`
	FirstName  string `json:"firstName"`
	MiddleName string `json:"middleName"`
	LastName   string `json:"lastName"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	AvatarUrl  string `json:"avatarUrl"`
}

type CreateAccountResponse struct {
	AccountId uuid.UUID `json:"account_id"`
}
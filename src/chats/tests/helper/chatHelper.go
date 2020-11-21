package helper

import (
	"chats/database"
	"chats/models"
	"chats/sdk"
	"chats/server"
	"encoding/json"
	"errors"
	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
)

type WSChatResponse struct {
	Type string      `json:"type"`
	Data WSChatMessagesDataResponse `json:"data"`
}
type WSChatMessagesDataResponse struct {
	Messages []models.WSChatMessagesDataMessageResponse `json:"messages"`
	Accounts []sdk.Account `json:"accounts"`
}

func NewChatAndSubscribe(sdkService *sdk.Sdk, accountId uuid.UUID, externalId string, referenceId string, role string) (*sdk.ChatNewResponse, error) {
	return nil, nil
	//newChatSubscribeRq := sdk.ChatNewSubscribeRequest{
	//	ApiRequestModel: sdk.ApiRequestModel{
	//		Method: "POST",
	//		Path:   "/chats/new/subscribe",
	//	},
	//	Body:            sdk.ChatNewSubscribeRequestBody{
	//		ReferenceId: referenceId,
	//		Account:     &sdk.AccountIdRequest{
	//			AccountId:  accountId,
	//			ExternalId: externalId,
	//		},
	//		Role:        role,
	//	},
	//}
	//
	//dataRq, err := json.Marshal(newChatSubscribeRq)
	//if err != nil {
	//	return nil, err
	//}
	//
	//dataRs, e := sdkService.
	//	Subject(topic).
	//	Request(dataRq)
	//if e != nil {
	//	return nil, errors.New(e.Message)
	//}
	//
	//newChatRs := &sdk.ChatNewResponse{Data: sdk.ChatNewResponseData{}}
	//err = json.Unmarshal(dataRs, newChatRs)
	//if err != nil {
	//	return nil, err
	//}
	//
	//return newChatRs, nil

}

func ChatSubscribe(sdkService *sdk.Sdk, chatId uuid.UUID, accountId uuid.UUID, externalId string, role string) (*sdk.ChatAccountSubscribeResponse, error) {

	subscribeRq := &sdk.ChatAccountSubscribeRequest{
		ApiRequestModel: sdk.ApiRequestModel{
			Method: "POST",
			Path:   "/chats/account/subscribe",
		},
		Body:            sdk.RoomMessageAccountSubscribeRequest{
			AccountId: accountId,
			RoomId: chatId,
			Role:   role,
		},
	}

	subscribeDataRq, err := json.Marshal(subscribeRq)
	if err != nil {
		return nil, err
	}

	subscribeDataRs, e := sdkService.
		Subject(topic).
		Request(subscribeDataRq)
	if e != nil {
		return nil, errors.New(e.Message)
	}

	subscribeRs := &sdk.ChatAccountSubscribeResponse{Data: sdk.ChatAccountSubscribeResponseData{}}
	err = json.Unmarshal(subscribeDataRs, subscribeRs)
	if err != nil {
		return nil, err
	}

	return subscribeRs, nil

}

func SendMessage(socket *websocket.Conn, accountId uuid.UUID, messageType string, message *models.WSChatMessageDataRequest) error {

	if message.ClientMessageId == "" {
		clientMessageId, _ := uuid.NewV4()
		message.ClientMessageId = clientMessageId.String()
	}

	msgRq := &models.WSChatMessagesRequest{
		AccountId: accountId,
		Type:      messageType,
		Data: models.WSChatMessagesDataRequest{
			Messages: []models.WSChatMessageDataRequest{ *message },
		},
	}

	request, err := json.Marshal(msgRq)
	if err != nil {
		return err
	}

	err = socket.WriteMessage(websocket.TextMessage, request)
	if err != nil {
		return err
	}
	return nil
}

func SendReadStatus(socket *websocket.Conn, chatId uuid.UUID, messageId uuid.UUID) error {

	msgRq := &models.WSChatMessageStatusRequest{
		Type:      server.EventMessageStatus,
		Data: models.WSChatMessageStatusDataRequest{
			Status:    database.MessageStatusRead,
			ChatId:    chatId,
			MessageId: messageId,
		},
	}

	request, err := json.Marshal(msgRq)
	if err != nil {
		return err
	}

	err = socket.WriteMessage(websocket.TextMessage, request)
	if err != nil {
		return err
	}
	return nil
}

func GetChatInfo(sdkService *sdk.Sdk, chatId uuid.UUID, accountId uuid.UUID, externalId string) (*sdk.ChatsInfoResponse, error) {

	chatInfoRq := &sdk.ChatsInfoRequest{
		ApiRequestModel: sdk.ApiRequestModel{
			Method: "GET",
			Path:   "/chats/info",
		},
		Body:            sdk.ChatsInfoRequestBody{
			Account: sdk.AccountIdRequest{
				AccountId:  accountId,
				ExternalId: externalId,
			},
			ChatsId: []uuid.UUID{chatId},
		},
	}

	subscribeDataRq, err := json.Marshal(chatInfoRq)
	if err != nil {
		return nil, err
	}

	chatInfoDataRs, e := sdkService.
		Subject(topic).
		Request(subscribeDataRq)
	if e != nil {
		return nil, errors.New(e.Message)
	}

	chatInfo := &sdk.ChatsInfoResponse{}
	err = json.Unmarshal(chatInfoDataRs, chatInfo)
	if err != nil {
		return nil, err
	}

	return chatInfo, nil
}

func GetChatsByAccount(sdkService *sdk.Sdk, accountId uuid.UUID, externalId string) (*sdk.ChatListResponse, error) {

	chatsRq := &sdk.ChatListRequest{
		ApiRequestModel: sdk.ApiRequestModel{
			Method: "GET",
			Path:   "/chats/chats",
		},
		Body: sdk.ChatListRequestBody{
			Account: sdk.AccountIdRequest{
				AccountId:  accountId,
				ExternalId: externalId,
			},
			Count:   1,
		},
	}

	chatsDataRq, err := json.Marshal(chatsRq)
	if err != nil {
		return nil, err
	}

	chatInfoDataRs, e := sdkService.
		Subject(topic).
		Request(chatsDataRq)
	if e != nil {
		return nil, errors.New(e.Message)
	}

	chatInfo := &sdk.ChatListResponse{}
	err = json.Unmarshal(chatInfoDataRs, chatInfo)
	if err != nil {
		return nil, err
	}

	return chatInfo, nil
}
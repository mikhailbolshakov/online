package server

import (
	"chats/database"
	"chats/system"
	"chats/models"
	"chats/sdk"
	"encoding/json"
	uuid "github.com/satori/go.uuid"
	"time"
)

const (
	ChatStatusOpened = "opened"
	ChatStatusClosed = "closed"
)

const (
	SubscribeActive   = 1
	SubscribeDeactive = 0

	UserTypeClient   = "client"
	UserTypeDoctor   = "doctor"
	UserTypeOperator = "operator"
	UserTypeBot      = "bot"
)


func (ws *WsServer) getChatHistory(params []byte) ([]byte, *system.Error) {
	data := &sdk.ChatMessagesHistoryRequest{}
	err := json.Unmarshal(params, data)
	if err != nil {
		return nil, system.UnmarshalRequestError1201(err, params)
	}

	if data.Body.ChatId == uuid.Nil {
		return nil, &system.Error{
			Error:   err,
			Message: sdk.GetError(1402),
			Code:    1402,
			Data:    params,
		}
	} else if data.Body.AccountId == uuid.Nil {
		return nil, &system.Error{
			Error:   err,
			Message: sdk.GetError(1403),
			Code:    1403,
			Data:    params,
		}
	}

	hp := &models.ChatMessageHistory{
		AccountId:   data.Body.AccountId,
		ChatId:      data.Body.ChatId,
		MessageId:   data.Body.MessageId,
		NewMessages: data.Body.NewMessages,
		UserType:    data.Body.Role,
		Count:       data.Body.Count,
		Admin:       data.Body.Admin,
		Search:      data.Body.Search,
		Date:        data.Body.Date,
		OnlyOneChat: data.Body.OnlyOneChat,
	}

	ChatMessagesResponse, err := ws.hub.app.DB.GetMessagesHistory(hp)
	if err != nil {
		return nil, &system.Error{
			Error:   err,
			Message: err.Error(),
			Code:    database.MysqlErrorCode,
			Data:    params,
		}
	}

	ChatAccountsResponse := []sdk.Account{}
	if len(ChatMessagesResponse) == 0 {
		ChatMessagesResponse = []sdk.ChatMessagesResponseDataItem{}
	}

	var chatIds []uuid.UUID

	for _, item := range ChatMessagesResponse {
		chatIds = append(chatIds, item.ChatId)
	}

	opponents, sentryErr := ws.hub.app.DB.GetChatOpponents(chatIds, ws.hub.app.Sdk)
	if sentryErr != nil {
		return nil, sentryErr
	}

	for _, item := range opponents {
		ChatAccountsResponse = append(ChatAccountsResponse, *ConvertAccountFromExpandedAccountModel(&item))
	}

	result, err := json.Marshal(sdk.ChatMessagesHistoryResponse{
		Data: sdk.ChatMessagesRecentResponseData{
			Messages: ChatMessagesResponse,
			Accounts: ChatAccountsResponse,
		},
	})
	if err != nil {
		return nil, system.MarshalError1011(err, params)
	}

	return result, nil
}

func (ws *WsServer) getChatRecent(params []byte) ([]byte, *system.Error) {
	data := &sdk.ChatMessagesRecentRequest{}
	err := json.Unmarshal(params, data)
	if err != nil {
		return nil, system.UnmarshalRequestError1201(err, params)
	}

	body := data.Body

	if body.ChatId == uuid.Nil {
		return nil, &system.Error{
			Error:   nil,
			Message: sdk.GetError(1402),
			Code:    1402,
			Data:    params,
		}
	} else if body.AccountId == uuid.Nil {
		return nil, &system.Error{
			Error:   nil,
			Message: sdk.GetError(1403),
			Code:    1403,
			Data:    params,
		}
	}

	rp := &models.ChatMessageHistory{
		AccountId:   body.AccountId,
		ChatId:      body.ChatId,
		MessageId:   body.MessageId,
		NewMessages: body.NewMessages,
		UserType:    body.Role,
		Admin:       body.Admin,
		Count:       body.Count,
		Search:      data.Body.Search,
		Date:        data.Body.Date,
	}

	ChatMessagesResponse, sentryErr := ws.hub.app.DB.GetMessagesRecent(rp)
	if sentryErr != nil {
		return nil, sentryErr
	}

	ChatAccountsResponse := []sdk.Account{}
	if len(ChatMessagesResponse) == 0 {
		ChatMessagesResponse = []sdk.ChatMessagesResponseDataItem{}
	}

	opponents, sentryErr := ws.hub.app.DB.GetChatOpponents(append([]uuid.UUID{}, rp.ChatId), ws.hub.app.Sdk)
	if sentryErr != nil {
		return nil, sentryErr
	}

	for _, item := range opponents {
		ChatAccountsResponse = append(ChatAccountsResponse, *ConvertAccountFromExpandedAccountModel(&item))
	}

	result, err := json.Marshal(sdk.ChatMessagesRecentResponse{
		Data: sdk.ChatMessagesRecentResponseData{
			Messages: ChatMessagesResponse,
			Accounts: ChatAccountsResponse,
		},
	})
	if err != nil {
		return nil, system.MarshalError1011(err, params)
	}

	return result, nil
}

func (ws *WsServer) setChatMessage(params []byte) ([]byte, *system.Error) {
	data := &sdk.ChatMessageRequest{}
	chatMessageResponseData := &sdk.ChatMessageResponseData{}
	err := json.Unmarshal(params, data)
	if err != nil {
		return nil, system.UnmarshalRequestError1201(err, params)
	}

	loc, err := system.Location()
	if err != nil {
		return nil, &system.Error{
			Error:   err,
			Message: system.LoadLocationError,
			Code:    system.LoadLocationErrorCode,
		}
	}

	if !database.ValidateType(data.Body.Type) {
		return nil, &system.Error{
			Error:   err,
			Message: database.MysqlChatMessageTypeIncorrect,
			Code:    database.MysqlChatMessageTypeIncorrectCode,
			Data:    params,
		}
	}

	var accountId uuid.UUID
	var role string
	var opponentsId []uuid.UUID

	paramsJson, err := json.Marshal(data.Body.Params)
	if err != nil {
		return nil, system.E(err)
	}

	messageModel := &models.ChatMessage{
		Id:              system.Uuid(),
		ClientMessageId: data.Body.ClientMessageId,
		ChatId:          data.Body.ChatId,
		Type:            data.Body.Type,
		Message:         data.Body.Message,
		FileId:          data.Body.FileId,
		Params:          string(paramsJson),
	}

	subscribes, e := ws.hub.app.DB.GetRoomSubscribers(data.Body.ChatId)
	if e != nil {
		return nil, e
	}

	for _, subscribe := range subscribes {
		if subscribe.AccountId == data.Body.AccountId {
			accountId = subscribe.AccountId
			role = subscribe.Role
			messageModel.SubscribeId = subscribe.Id
		} else {
			opponentsId = append(opponentsId, subscribe.Id)
		}
	}

	if accountId != uuid.Nil {

		e := ws.hub.app.DB.NewMessageTransact(messageModel, opponentsId)
		if e != nil {
			return nil, e
		}

		messages := []interface{}{}
		clients := []sdk.Account{}

		tmpMessageResponse := &models.WSChatMessagesDataMessageResponse{
			Id:              messageModel.Id,
			ClientMessageId: messageModel.ClientMessageId,
			InsertDate:      messageModel.CreatedAt.In(loc).Format(time.RFC3339),
			ChatId:          messageModel.ChatId,
			AccountId:       data.Body.AccountId,
			Sender:          role,
			Status:          database.MessageStatusRecd,
			Type:            data.Body.Type,
			Text:            messageModel.Message,
			Params:          data.Body.Params,
		}
		chatMessageResponseData = &sdk.ChatMessageResponseData{
			Id:              messageModel.Id,
			ClientMessageId: messageModel.ClientMessageId,
			InsertDate:      messageModel.CreatedAt.In(loc).Format(time.RFC3339),
			ChatId:          messageModel.ChatId,
			AccountId:       data.Body.AccountId,
			Sender:          role,
			Status:          database.MessageStatusRecd,
			Type:            messageModel.Type,
			Text:            messageModel.Message,
			Params:          data.Body.Params,
		}

		if len(messageModel.FileId) > 0 {
			file := &sdk.FileModel{Id: messageModel.FileId}
			// TODO: needs to be redevelopped
			// we need to get file parameters from somewhere else
			sdkErr := ws.hub.app.Sdk.File(file, data.Body.ChatId, data.Body.AccountId)
			if sdkErr != nil {
				return nil, &system.Error{
					Error:   sdkErr.Error,
					Message: sdkErr.Message,
					Code:    sdkErr.Code,
					Data:    sdkErr.Data,
				}
			}
			tmpMessageFileResponse := &models.WSChatMessagesDataMessageFileResponse{
				WSChatMessagesDataMessageResponse: *tmpMessageResponse,
				File:                              *file,
			}
			chatMessageResponseData.File = *file
			messages = append(messages, tmpMessageFileResponse)
		} else {
			messages = append(messages, tmpMessageResponse)
		}

		accountModel, err := ws.hub.app.DB.GetAccount(accountId, "")
		if err != nil {
			return nil, err
		}

		clients = append(clients, *ConvertAccountFromModel(accountModel))

		responseData := &models.WSChatMessagesDataResponse{
			Messages: messages,
			Accounts: clients,
		}
		wsChatResponse := &models.WSChatResponse{
			Type: EventMessage,
			Data: responseData,
		}
		roomMessage := &RoomMessage{
			RoomId:  data.Body.ChatId,
			Message: wsChatResponse,
		}

		ws.hub.SendMessageToRoom(roomMessage)

	} else {
		return nil, &system.Error{
			Error:   err,
			Message: database.MysqlChatAccessDenied,
			Code:    database.MysqlChatAccessDeniedCode,
			Data:    params,
		}
	}

	result, err := json.Marshal(sdk.ChatMessageResponse{
		Data: *chatMessageResponseData,
	})
	if err != nil {
		return nil, system.MarshalError1011(err, params)
	}

	return result, nil
}

func (ws *WsServer) sendClientMessage(params []byte) ([]byte, *system.Error) {
	data := &sdk.MessageToMobileClientRequest{}
	err := json.Unmarshal(params, data)
	if err != nil {
		return nil, system.UnmarshalRequestError1201(err, params)
	}

	if len(data.Body.Type) == 0 {
		return nil, &system.Error{
			Error:   err,
			Message: sdk.GetError(1404),
			Code:    1404,
			Data:    params,
		}
	}

	if data.Body.AccountId == uuid.Nil {
		return nil, &system.Error{
			Error:   err,
			Message: sdk.GetError(1403),
			Code:    1403,
			Data:    params,
		}
	}

	roomMessage := &RoomMessage{
		SendPush:  true,
		RoomId:    uuid.Nil,
		AccountId: data.Body.AccountId,
		Message: &models.WSChatResponse{
			Type: data.Body.Type,
			Data: data.Body.Data,
		},
	}

	go ws.hub.SendMessageToRoom(roomMessage)

	result, err := json.Marshal(sdk.ChatStatusResponse{
		sdk.ChatStatusDataResponse{Result: true},
	})
	if err != nil {
		return nil, system.MarshalError1011(err, params)
	}

	return result, nil
}

func (ws *WsServer) changeChatAccountSubscribe(params []byte) ([]byte, *system.Error) {
	data := &sdk.ChatAccountSubscribeChangeRequest{}
	err := json.Unmarshal(params, data)
	if err != nil {
		return nil, system.UnmarshalRequestError1201(err, params)
	}

	if data.Body.ChatId == uuid.Nil {
		return nil, &system.Error{
			Error:   err,
			Message: sdk.GetError(1402),
			Code:    1402,
			Data:    params,
		}
	} else if data.Body.OldAccountId == uuid.Nil || data.Body.NewAccountId == uuid.Nil {
		return nil, &system.Error{
			Error:   err,
			Message: sdk.GetError(1403),
			Code:    1403,
			Data:    params,
		}
	}

	err = ws.hub.app.DB.SubscribeUserChange(data)
	if err != nil {
		return nil, &system.Error{
			Error:   err,
			Message: database.MysqlChatSubscribeChangeError,
			Code:    database.MysqlChatSubscribeChangeErrorCode,
			Data:    params,
		}
	}

	result, err := json.Marshal(&sdk.ChatAccountSubscribeChangeResponse{
		Data: sdk.BoolResponseData{Result: true},
	})
	if err != nil {
		return nil, system.MarshalError1011(err, params)
	}

	return result, nil
}

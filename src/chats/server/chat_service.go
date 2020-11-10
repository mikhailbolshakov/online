package server

import (
	"chats/database"
	"chats/infrastructure"
	"chats/models"
	"chats/sdk"
	"chats/sentry"
	"encoding/json"
	uuid "github.com/satori/go.uuid"
	"time"
)

const (
	ChatStatusOpened = "opened"
	ChatStatusClosed = "closed"
)

/**
Список чатов пользователя
*/
func (ws *WsServer) getChatChats(params []byte) ([]byte, *sentry.SystemError) {
	data := &sdk.ChatListRequest{}
	err := json.Unmarshal(params, data)
	if err != nil {
		return nil, infrastructure.UnmarshalRequestError1201(err, params)
	}

	responseData, err := ws.hub.app.DB.GetAccountChats(data.Body.AccountId, data.Body.Count, ws.hub.app.Sdk)
	if err != nil {
		return nil, &sentry.SystemError{
			Error:   err,
			Message: database.MysqlUserChatListError,
			Code:    database.MysqlUserChatListErrorCode,
			Data:    params,
		}
	}

	result, err := json.Marshal(sdk.ChatListResponse{Data: responseData})
	if err != nil {
		return nil, infrastructure.MarshalError1011(err, params)
	}

	return result, nil
}

//	Получение информации о чате
func (ws *WsServer) getChatById(params []byte) ([]byte, *sentry.SystemError) {
	data := &sdk.ChatInfoRequest{}
	err := json.Unmarshal(params, data)
	if err != nil {
		return nil, infrastructure.UnmarshalRequestError1201(err, params)
	}

	responseData, err := ws.hub.app.DB.GetChatById(data.Body.ChatId, data.Body.AccountId, ws.hub.app.Sdk)
	if err != nil {
		return nil, &sentry.SystemError{
			Error:   err,
			Message: database.MysqlChatInfoError,
			Code:    database.MysqlChatInfoErrorCode,
			Data:    params,
		}
	}

	result, err := json.Marshal(sdk.ChatInfoResponse{Data: *responseData})
	if err != nil {
		return nil, infrastructure.MarshalError1011(err, params)
	}

	return result, nil
}

func (ws *WsServer) ChatByOrder(params []byte) ([]byte, *sentry.SystemError) {
	data := &sdk.RefereneChatRequest{}
	err := json.Unmarshal(params, data)
	if err != nil {
		return nil, infrastructure.UnmarshalRequestError1201(err, params)
	}

	if len(data.Body) == 0 {
		return nil, infrastructure.UnmarshalRequestError1201(err, params)
	}

	responseData, err := ws.hub.app.DB.GetChatsByReference(data.Body, ws.hub.app.Sdk)

	if err != nil {
		return nil, &sentry.SystemError{
			Error:   err,
			Message: database.MysqlChatInfoError,
			Code:    database.MysqlChatInfoErrorCode,
			Data:    params,
		}
	}

	result, err := json.Marshal(sdk.ChatListResponse{Data: responseData})
	if err != nil {
		return nil, infrastructure.MarshalError1011(err, params)
	}

	return result, nil
}

func (ws *WsServer) getChatsInfo(params []byte) ([]byte, *sentry.SystemError) {
	data := &sdk.ChatsInfoRequest{}
	err := json.Unmarshal(params, data)
	if err != nil {
		return nil, infrastructure.UnmarshalRequestError1201(err, params)
	}

	if data.Body.AccountId == uuid.Nil {
		return nil, &sentry.SystemError{
			Error:   err,
			Message: sdk.GetError(1403),
			Code:    1403,
			Data:    params,
		}
	} else if len(data.Body.ChatsId) == 0 {
		return nil, &sentry.SystemError{
			Error:   err,
			Message: sdk.GetError(1402),
			Code:    1402,
			Data:    params,
		}
	}

	responseData, err := ws.hub.app.DB.GetChatsById(data.Body.ChatsId, data.Body.AccountId, ws.hub.app.Sdk)
	if err != nil {
		return nil, &sentry.SystemError{
			Error:   err,
			Message: database.MysqlChatInfoError,
			Code:    database.MysqlChatInfoErrorCode,
			Data:    params,
		}
	}

	result, err := json.Marshal(sdk.ChatsInfoResponse{Data: responseData})
	if err != nil {
		return nil, infrastructure.MarshalError1011(err, params)
	}

	return result, nil
}

func (ws *WsServer) getLastChat(params []byte) ([]byte, *sentry.SystemError) {
	data := &sdk.ChatsLastRequest{}
	err := json.Unmarshal(params, data)

	responseData, _ := ws.hub.app.DB.GetAccountChats(data.Body.AccountId, 1, ws.hub.app.Sdk)

	if err != nil {
		return nil, &sentry.SystemError{
			Error:   err,
			Message: database.MysqlUserChatListError,
			Code:    database.MysqlUserChatListErrorCode,
			Data:    params,
		}
	}

	res := sdk.ChatListResponseDataItem{}

	if len(responseData) != 0 {
		res = responseData[0]
	}

	result, err := json.Marshal(sdk.ChatLastResponse{Data: res})

	if err != nil {
		return nil, infrastructure.MarshalError1011(err, params)
	}

	return result, nil
}

func (ws *WsServer) getChatHistory(params []byte) ([]byte, *sentry.SystemError) {
	data := &sdk.ChatMessagesHistoryRequest{}
	err := json.Unmarshal(params, data)
	if err != nil {
		return nil, infrastructure.UnmarshalRequestError1201(err, params)
	}

	if data.Body.ChatId == uuid.Nil {
		return nil, &sentry.SystemError{
			Error:   err,
			Message: sdk.GetError(1402),
			Code:    1402,
			Data:    params,
		}
	} else if data.Body.AccountId == uuid.Nil {
		return nil, &sentry.SystemError{
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
		return nil, &sentry.SystemError{
			Error:   err,
			Message: err.Error(),
			Code:    database.MysqlErrorCode,
			Data:    params,
		}
	}

	ChatUsersResponse := []sdk.AccountModel{}
	if len(ChatMessagesResponse) == 0 {
		ChatMessagesResponse = []sdk.ChatMessagesResponseDataItem{}
	}

	var chatIds []uuid.UUID

	for _, item := range ChatMessagesResponse {
		chatIds = append(chatIds, item.ChatId)
	}

	opponents, err := ws.hub.app.DB.GetChatOpponents(chatIds, ws.hub.app.Sdk)
	if err != nil {
		return nil, &sentry.SystemError{
			Error:   err,
			Message: err.Error(),
			Code:    database.MysqlErrorCode,
			Data:    params,
		}
	}

	for _, item := range opponents {
		ChatUsersResponse = append(ChatUsersResponse, sdk.AccountModel{
			Id:         item.Id,
			FirstName:  item.FirstName,
			LastName:   item.LastName,
			MiddleName: item.MiddleName,
			Photo:      item.Photo,
		})
	}

	result, err := json.Marshal(sdk.ChatMessagesHistoryResponse{
		Data: sdk.ChatMessagesRecentResponseData{
			Messages: ChatMessagesResponse,
			Accounts: ChatUsersResponse,
		},
	})
	if err != nil {
		return nil, infrastructure.MarshalError1011(err, params)
	}

	return result, nil
}

func (ws *WsServer) getChatRecent(params []byte) ([]byte, *sentry.SystemError) {
	data := &sdk.ChatMessagesRecentRequest{}
	err := json.Unmarshal(params, data)
	if err != nil {
		return nil, infrastructure.UnmarshalRequestError1201(err, params)
	}

	body := data.Body

	if body.ChatId == uuid.Nil {
		return nil, &sentry.SystemError{
			Error:   nil,
			Message: sdk.GetError(1402),
			Code:    1402,
			Data:    params,
		}
	} else if body.AccountId == uuid.Nil {
		return nil, &sentry.SystemError{
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

	ChatMessagesResponse, err := ws.hub.app.DB.GetMessagesRecent(rp)
	if err != nil {
		return nil, &sentry.SystemError{
			Error:   err,
			Message: err.Error(),
			Code:    database.MysqlErrorCode,
			Data:    params,
		}
	}

	ChatUsersResponse := []sdk.AccountModel{}
	if len(ChatMessagesResponse) == 0 {
		ChatMessagesResponse = []sdk.ChatMessagesResponseDataItem{}
	}

	opponents, err := ws.hub.app.DB.GetChatOpponents(append([]uuid.UUID{}, rp.ChatId), ws.hub.app.Sdk)
	if err != nil {
		return nil, &sentry.SystemError{
			Error:   err,
			Message: err.Error(),
			Code:    database.MysqlErrorCode,
			Data:    params,
		}
	}

	for _, item := range opponents {
		ChatUsersResponse = append(ChatUsersResponse, sdk.AccountModel{
			Id:         item.Id,
			FirstName:  item.FirstName,
			LastName:   item.LastName,
			MiddleName: item.MiddleName,
			Photo:      item.Photo,
		})
	}

	result, err := json.Marshal(sdk.ChatMessagesRecentResponse{
		Data: sdk.ChatMessagesRecentResponseData{
			Messages: ChatMessagesResponse,
			Accounts: ChatUsersResponse,
		},
	})
	if err != nil {
		return nil, infrastructure.MarshalError1011(err, params)
	}

	return result, nil
}

/**
SET
*/
func (ws *WsServer) setChatNew(params []byte) ([]byte, *sentry.SystemError) {
	data := &sdk.ChatNewRequest{}
	err := json.Unmarshal(params, data)
	if err != nil {
		return nil, infrastructure.UnmarshalRequestError1201(err, params)
	}

	id, _ := uuid.NewV4()
	chatModel := &models.Chat{
		Id:          id,
		ReferenceId: data.Body.ReferenceId,
		Status:      ChatStatusOpened,
	}

	chatId, err := ws.hub.app.DB.ChatCreate(chatModel)
	if err != nil {
		return nil, &sentry.SystemError{
			Error:   err,
			Message: database.MysqlChatCreateError,
			Code:    database.MysqlChatCreateErrorCode,
			Data:    params,
		}
	}

	response := sdk.ChatNewResponseData{ChatId: chatId}
	responseData := sdk.ChatNewResponse{Data: response}
	result, err := json.Marshal(responseData)
	if err != nil {
		return nil, infrastructure.MarshalError1011(err, params)
	}

	return result, nil
}

/**
SET
*/
func (ws *WsServer) setChatNewSubscribe(params []byte) ([]byte, *sentry.SystemError) {

	data := &sdk.ChatNewSubscribeRequest{}
	err := json.Unmarshal(params, data)
	if err != nil {
		return nil, infrastructure.UnmarshalRequestError1201(err, params)
	}

	chatModel := &models.Chat{
		ReferenceId: data.Body.ReferenceId,
		Status:      ChatStatusOpened,
	}

	chatId, err := ws.hub.app.DB.ChatCreate(chatModel)
	if err != nil {
		return nil, &sentry.SystemError{
			Error:   err,
			Message: database.MysqlChatCreateError,
			Code:    database.MysqlChatCreateErrorCode,
			Data:    params,
		}
	}

	account, err := ws.hub.app.DB.GetAccount(data.Body.Account.AccountId, data.Body.Account.ExternalId)
	if err != nil {
		return nil, &sentry.SystemError{
			Error:   err,
			Message: database.MysqlChatCreateError,
			Code:    database.MysqlChatCreateErrorCode,
			Data:    params,
		}
	}

	_, err = ws.hub.app.DB.SubscribeAccount(chatModel, account.Id)
	if err != nil {
		return nil, &sentry.SystemError{
			Error:   err,
			Message: database.MysqlChatUserSubscribeError + err.Error(),
			Code:    database.MysqlChatUserSubscribeErrorCode,
			Data:    params,
		}
	}

	//	subscribe websocket hub
	roomMessage := &RoomMessage{
		SendPush:  false,
		AccountId: uuid.Nil,
		RoomId:    uuid.Nil,
		Message: &models.WSChatResponse{
			Type: infrastructure.SystemMsgTypeUserSubscribe,
			Data: &sdk.ChatAccountSubscribeRequestBody{
				Account: &sdk.AccountRequest{AccountId: account.Id},
				ChatId:  chatId,
			},
		},
	}

	go ws.hub.SendMessageToRoom(roomMessage)

	response := sdk.ChatNewResponseData{ChatId: chatId}
	responseData := sdk.ChatNewResponse{Data: response}
	result, err := json.Marshal(responseData)
	if err != nil {
		return nil, infrastructure.MarshalError1011(err, params)
	}

	return result, nil
}

func (ws *WsServer) setChatAccountSubscribe(params []byte) ([]byte, *sentry.SystemError) {

	data := &sdk.ChatAccountSubscribeRequest{}
	err := json.Unmarshal(params, data)
	if err != nil {
		return nil, infrastructure.UnmarshalRequestError1201(err, params)
	}

	account, err := ws.hub.app.DB.GetAccount(data.Body.Account.AccountId, data.Body.Account.ExternalId)
	if err != nil {
		return nil, &sentry.SystemError{
			Error:   err,
			Message: database.MysqlChatCreateError,
			Code:    database.MysqlChatCreateErrorCode,
			Data:    params,
		}
	}

	chat := ws.hub.app.DB.Chat(data.Body.ChatId)
	if chat.Id == uuid.Nil {
		return nil, &sentry.SystemError{
			Error:   err,
			Message: database.MysqlChatCreateError,
			Code:    database.MysqlChatCreateErrorCode,
			Data:    params,
		}
	}

	_, err = ws.hub.app.DB.SubscribeAccount(chat, account.Id)
	if err != nil {
		return nil, &sentry.SystemError{
			Error:   err,
			Message: database.MysqlChatUserSubscribeError + err.Error(),
			Code:    database.MysqlChatUserSubscribeErrorCode,
			Data:    params,
		}
	}

	//	subscribe websocket hub
	roomMessage := &RoomMessage{
		SendPush:  false,
		AccountId: uuid.Nil,
		RoomId:    uuid.Nil,
		Message: &models.WSChatResponse{
			Type: infrastructure.SystemMsgTypeUserSubscribe,
			Data: data.Body,
		},
	}

	go ws.hub.SendMessageToRoom(roomMessage)

	result, err := json.Marshal(sdk.ChatAccountSubscribeResponse{
		sdk.ChatAccountSubscribeResponseData{Result: true},
	})
	if err != nil {
		return nil, infrastructure.MarshalError1011(err, params)
	}

	return result, nil
}

func (ws *WsServer) setChatUserUnsubscribe(params []byte) ([]byte, *sentry.SystemError) {
	data := &sdk.ChatUserUnsubscribeRequest{}
	err := json.Unmarshal(params, data)
	if err != nil {
		return nil, infrastructure.UnmarshalRequestError1201(err, params)
	}

	err = ws.hub.app.DB.UnsubscribeUser(data.Body.ChatId, data.Body.AccountId)
	if err != nil {
		return nil, &sentry.SystemError{
			Error:   err,
			Message: database.MysqlChatUserUnsubscribeError,
			Code:    database.MysqlChatUserUnsubscribeErrorCode,
			Data:    params,
		}
	}

	//	unsubscribe websocket hub
	roomMessage := &RoomMessage{
		Message: &models.WSChatResponse{
			Type: infrastructure.SystemMsgTypeUserUnsubscribe,
			Data: data.Body,
		},
	}
	go ws.hub.SendMessageToRoom(roomMessage)

	result, err := json.Marshal(sdk.ChatAccountSubscribeResponse{
		sdk.ChatAccountSubscribeResponseData{Result: true},
	})

	if err != nil {
		return nil, infrastructure.MarshalError1011(err, params)
	}

	return result, nil
}

func (ws *WsServer) setChatMessage(params []byte) ([]byte, *sentry.SystemError) {
	data := &sdk.ChatMessageRequest{}
	chatMessageResponseData := &sdk.ChatMessageResponseData{}
	err := json.Unmarshal(params, data)
	if err != nil {
		return nil, infrastructure.UnmarshalRequestError1201(err, params)
	}

	loc, err := infrastructure.Location()
	if err != nil {
		return nil, &sentry.SystemError{
			Error:   err,
			Message: infrastructure.LoadLocationError,
			Code:    infrastructure.LoadLocationErrorCode,
		}
	}

	if !database.ValidateType(data.Body.Type) {
		return nil, &sentry.SystemError{
			Error:   err,
			Message: database.MysqlChatMessageTypeIncorrect,
			Code:    database.MysqlChatMessageTypeIncorrectCode,
			Data:    params,
		}
	}

	var accountId uuid.UUID
	var role string
	var opponentsId []uuid.UUID
	messageModel := &models.ChatMessage{
		ClientMessageId: data.Body.ClientMessageId,
		ChatId:          data.Body.ChatId,
		Type:            data.Body.Type,
		Message:         data.Body.Message,
		FileId:          data.Body.FileId,
	}

	subscribes := ws.hub.app.DB.ChatSubscribes(data.Body.ChatId)
	for _, subscribe := range subscribes {
		if subscribe.AccountId == data.Body.AccountId {
			accountId = subscribe.AccountId
			role = subscribe.Role
			messageModel.SubscribeId = subscribe.SubscribeId
		} else {
			opponentsId = append(opponentsId, subscribe.SubscribeId)
		}
	}

	if accountId != uuid.Nil {
		err = ws.hub.app.DB.NewMessageTransact(messageModel, data.Body.Params, opponentsId)
		if err != nil {
			return nil, &sentry.SystemError{
				Error:   err,
				Message: database.MysqlChatCreateMessageError,
				Code:    database.MysqlChatCreateMessageErrorCode,
				Data:    params,
			}
		}

		messages := []interface{}{}
		clients := []sdk.AccountModel{}
		messageParams, paramsErr := ws.hub.app.DB.GetParamsMap(messageModel.Id)
		if paramsErr != nil {
			return nil, &sentry.SystemError{
				Error:   paramsErr,
				Message: database.MysqlChatMessageParamsError,
				Code:    database.MysqlChatMessageParamsErrorCode,
				Data:    params,
			}
		}
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
			Params:          messageParams,
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
			Params:          messageParams,
		}

		if len(messageModel.FileId) > 0 {
			file := &sdk.FileModel{Id: messageModel.FileId}
			sdkErr := ws.hub.app.Sdk.File(file, data.Body.ChatId, data.Body.AccountId)
			if sdkErr != nil {
				return nil, &sentry.SystemError{
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

		accountModel := &sdk.AccountModel{
			Id: accountId,
		}
		consultation := ws.hub.app.DB.Chat(data.Body.ChatId)
		err := ws.hub.app.Sdk.VagueUserById(accountModel, role, consultation.ReferenceId)
		if err != nil {
			return nil, &sentry.SystemError{
				Error:   err.Error,
				Code:    err.Code,
				Message: err.Message,
				Data:    err.Data,
			}
		}
		clients = append(clients, *accountModel)

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

		if data.Body.Type == database.MessageTypeOrderDetail {
			_, exists := data.Body.Params["orderId"]

			if exists {
				ws.hub.app.DB.ChatDeactivateNotice(data.Body.ChatId)
			}
		}
	} else {
		return nil, &sentry.SystemError{
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
		return nil, infrastructure.MarshalError1011(err, params)
	}

	return result, nil
}

func (ws *WsServer) setChatStatus(params []byte) ([]byte, *sentry.SystemError) {
	data := &sdk.ChatStatusRequest{}
	err := json.Unmarshal(params, data)
	if err != nil {
		return nil, infrastructure.UnmarshalRequestError1201(err, params)
	}

	err = ws.hub.app.DB.ChatChangeStatus(data.Body.ChatId, database.ChatStatusClosed)
	if err != nil {
		return nil, &sentry.SystemError{
			Error:   err,
			Message: database.MysqlChatChangeStatusError,
			Code:    database.MysqlChatChangeStatusErrorCode,
			Data:    params,
		}
	}

	delete(ws.hub.rooms, data.Body.ChatId)

	result, err := json.Marshal(sdk.ChatStatusResponse{
		sdk.ChatStatusDataResponse{Result: true},
	})
	if err != nil {
		return nil, infrastructure.MarshalError1011(err, params)
	}

	return result, nil
}

func (ws *WsServer) sendClientMessage(params []byte) ([]byte, *sentry.SystemError) {
	data := &sdk.MessageToMobileClientRequest{}
	err := json.Unmarshal(params, data)
	if err != nil {
		return nil, infrastructure.UnmarshalRequestError1201(err, params)
	}

	if len(data.Body.Type) == 0 {
		return nil, &sentry.SystemError{
			Error:   err,
			Message: sdk.GetError(1404),
			Code:    1404,
			Data:    params,
		}
	}

	if data.Body.AccountId == uuid.Nil {
		return nil, &sentry.SystemError{
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
		return nil, infrastructure.MarshalError1011(err, params)
	}

	return result, nil
}

func (ws *WsServer) clientConsultationUpdate(params []byte) ([]byte, *sentry.SystemError) {
	data := &sdk.ClientConsultationUpdateRequest{}
	err := json.Unmarshal(params, data)
	if err != nil {
		return nil, infrastructure.UnmarshalRequestError1201(err, params)
	}

	if data.Body.AccountId != uuid.Nil {
		response := &RoomMessage{
			AccountId: data.Body.AccountId,
			Message: &models.WSChatResponse{
				Type: EventConsultationUpdate,
				Data: data.Body.Data,
			},
		}

		ws.hub.SendMessageToRoom(response)
	} else {
		return nil, &sentry.SystemError{
			Message: sdk.GetError(1403),
			Code:    1403,
			Data:    params,
		}
	}

	result, err := json.Marshal(sdk.ChatStatusResponse{
		sdk.ChatStatusDataResponse{Result: true},
	})
	if err != nil {
		return nil, infrastructure.MarshalError1011(err, params)
	}

	return result, nil
}

func (ws *WsServer) changeChatUserSubscribe(params []byte) ([]byte, *sentry.SystemError) {
	data := &sdk.ChatAccountSubscribeChangeRequest{}
	err := json.Unmarshal(params, data)
	if err != nil {
		return nil, infrastructure.UnmarshalRequestError1201(err, params)
	}

	if data.Body.ChatId == uuid.Nil {
		return nil, &sentry.SystemError{
			Error:   err,
			Message: sdk.GetError(1402),
			Code:    1402,
			Data:    params,
		}
	} else if data.Body.OldAccountId == uuid.Nil || data.Body.NewAccountId == uuid.Nil {
		return nil, &sentry.SystemError{
			Error:   err,
			Message: sdk.GetError(1403),
			Code:    1403,
			Data:    params,
		}
	}

	err = ws.hub.app.DB.SubscribeUserChange(data)
	if err != nil {
		return nil, &sentry.SystemError{
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
		return nil, infrastructure.MarshalError1011(err, params)
	}

	return result, nil
}

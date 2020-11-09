package server

import (
	"chats/database"
	"chats/models"
	"chats/service"
	"encoding/json"
	"github.com/mkevac/gopinba"
	"gitlab.medzdrav.ru/health-service/go-sdk"
	"net/http"
	"strings"
	"time"
)

func (ws *WsServer) ApiConsumer() {
	apiConsumerErrorChan := make(chan *sdk.Error, 2048)
	go ws.hub.app.Sdk.
		Subject(ws.apiTopic).
		ApiConsumer(ws.GetHandler, apiConsumerErrorChan)

	for {
		err := <-apiConsumerErrorChan
		service.SetError(&models.SystemError{
			Error:   err.Error,
			Message: err.Message,
			Code:    err.Code,
			Data:    err.Data,
		})
	}
}

func (ws *WsServer) GetHandler(request []byte) ([]byte, *sdk.Error) {
	data, err := ws.Router(request)
	if err != nil {
		return nil, &sdk.Error{
			Error:   err.Error,
			Message: err.Message,
			Code:    err.Code,
			Data:    err.Data,
		}
	} else {
		return data, nil
	}
}

func (ws *WsServer) Router(request []byte) ([]byte, *models.SystemError) {
	clientRequest := &sdk.ApiRequest{}
	err := json.Unmarshal(request, &clientRequest)
	if err != nil {
		return nil, service.UnmarshalRequestError1201(err, request)
	}

	if ws.hub.app.DB.Pinba != nil {
		tags := map[string]string{
			"group":  "sdk",
			"topic":  ws.apiTopic,
			"method": clientRequest.Method,
			"path":   clientRequest.Path,
		}
		timer := gopinba.TimerStart(tags)
		defer func(timer *gopinba.Timer, db *database.Storage) {
			timer.Stop()
			db.Pinba.SendRequest(&gopinba.Request{
				Tags:        timer.Tags,
				RequestTime: timer.GetDuration(),
			})
		}(timer, ws.hub.app.DB)
	}

	switch strings.ToUpper(clientRequest.Method) {
	case http.MethodGet:
		switch clientRequest.Path {
		case "/chats/chats":
			return ws.getChatChats(request)

		case "/chats/chat":
			return ws.getChatById(request)

		case "/order/chats":
			return ws.ChatByOrder(request)

		case "/chats/info":
			return ws.getChatsInfo(request)

		case "/chats/messages/update",
			"/chats/chat/recent":
			return ws.getChatRecent(request)

		case "/chats/messages/history",
			"/chats/chat/history":
			return ws.getChatHistory(request)

		case "/chats/last":
			return ws.getLastChat(request)

		}
	case http.MethodPost:
		switch clientRequest.Path {
		case "/chats/new":
			return ws.setChatNew(request)

			//	Создаём новый чат и подписываем
		case "/chats/new/subscribe":
			return ws.setChatNewSubscribe(request)

			//	Подписываем {клиента|доктора|оператора} на чат
		case "/chats/user/subscribe":
			return ws.setChatUserSubscribe(request)

		case "/chats/user/unsubscribe":
			return ws.setChatUserUnsubscribe(request)

		case "/chats/message":
			return ws.setChatMessage(request)

		case "/chats/status":
			return ws.setChatStatus(request)

		case "/ws/client/message":
			return ws.sendClientMessage(request)

		case "/ws/client/consultation/update":
			return ws.clientConsultationUpdate(request)
		}

	case http.MethodPut:
		switch clientRequest.Path {
		case "/chats/user/subscribe":
			return ws.changeChatUserSubscribe(request)
		}
	}

	return nil, &models.SystemError{
		Error:   nil,
		Message: sdk.GetError(1203),
		Code:    1203,
		Data:    request,
	}
}

/**
Список чатов пользователя
*/
func (ws *WsServer) getChatChats(params []byte) ([]byte, *models.SystemError) {
	data := &sdk.ChatListRequest{}
	err := json.Unmarshal(params, data)
	if err != nil {
		return nil, service.UnmarshalRequestError1201(err, params)
	}

	responseData, err := ws.hub.app.DB.GetUserChats(data.Body.UserId, data.Body.Count, ws.hub.app.Sdk)
	if err != nil {
		return nil, &models.SystemError{
			Error:   err,
			Message: database.MysqlUserChatListError,
			Code:    database.MysqlUserChatListErrorCode,
			Data:    params,
		}
	}

	result, err := json.Marshal(sdk.ChatListResponse{Data: responseData})
	if err != nil {
		return nil, service.MarshalError1011(err, params)
	}

	return result, nil
}

//	Получение информации о чате
func (ws *WsServer) getChatById(params []byte) ([]byte, *models.SystemError) {
	data := &sdk.ChatInfoRequest{}
	err := json.Unmarshal(params, data)
	if err != nil {
		return nil, service.UnmarshalRequestError1201(err, params)
	}

	responseData, err := ws.hub.app.DB.GetChatById(data.Body.ChatId, data.Body.UserId, ws.hub.app.Sdk)
	if err != nil {
		return nil, &models.SystemError{
			Error:   err,
			Message: database.MysqlChatInfoError,
			Code:    database.MysqlChatInfoErrorCode,
			Data:    params,
		}
	}

	result, err := json.Marshal(sdk.ChatInfoResponse{Data: *responseData})
	if err != nil {
		return nil, service.MarshalError1011(err, params)
	}

	return result, nil
}

func (ws *WsServer) ChatByOrder(params []byte) ([]byte, *models.SystemError) {
	data := &sdk.OrderChatRequest{}
	err := json.Unmarshal(params, data)
	if err != nil {
		return nil, service.UnmarshalRequestError1201(err, params)
	}

	if len(data.Body) == 0 {
		return nil, service.UnmarshalRequestError1201(err, params)
	}

	responseData, err := ws.hub.app.DB.GetChatsByOrder(data.Body, ws.hub.app.Sdk)

	if err != nil {
		return nil, &models.SystemError{
			Error:   err,
			Message: database.MysqlChatInfoError,
			Code:    database.MysqlChatInfoErrorCode,
			Data:    params,
		}
	}

	result, err := json.Marshal(sdk.ChatListResponse{Data: responseData})
	if err != nil {
		return nil, service.MarshalError1011(err, params)
	}

	return result, nil
}

func (ws *WsServer) getChatsInfo(params []byte) ([]byte, *models.SystemError) {
	data := &sdk.ChatsInfoRequest{}
	err := json.Unmarshal(params, data)
	if err != nil {
		return nil, service.UnmarshalRequestError1201(err, params)
	}

	if data.Body.UserId <= 0 {
		return nil, &models.SystemError{
			Error:   err,
			Message: sdk.GetError(1403),
			Code:    1403,
			Data:    params,
		}
	} else if len(data.Body.ChatsId) == 0 {
		return nil, &models.SystemError{
			Error:   err,
			Message: sdk.GetError(1402),
			Code:    1402,
			Data:    params,
		}
	}

	responseData, err := ws.hub.app.DB.GetChatsById(data.Body.ChatsId, data.Body.UserId, ws.hub.app.Sdk)
	if err != nil {
		return nil, &models.SystemError{
			Error:   err,
			Message: database.MysqlChatInfoError,
			Code:    database.MysqlChatInfoErrorCode,
			Data:    params,
		}
	}

	result, err := json.Marshal(sdk.ChatsInfoResponse{Data: responseData})
	if err != nil {
		return nil, service.MarshalError1011(err, params)
	}

	return result, nil
}

func (ws *WsServer) getLastChat(params []byte) ([]byte, *models.SystemError) {
	data := &sdk.ChatsLastRequest{}
	err := json.Unmarshal(params, data)

	responseData, _ := ws.hub.app.DB.GetUserChats(data.Body.UserId, 1, ws.hub.app.Sdk)

	if err != nil {
		return nil, &models.SystemError{
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
		return nil, service.MarshalError1011(err, params)
	}

	return result, nil
}

func (ws *WsServer) getChatHistory(params []byte) ([]byte, *models.SystemError) {
	data := &sdk.ChatMessagesHistoryRequest{}
	err := json.Unmarshal(params, data)
	if err != nil {
		return nil, service.UnmarshalRequestError1201(err, params)
	}

	if data.Body.ChatId <= 0 {
		return nil, &models.SystemError{
			Error:   err,
			Message: sdk.GetError(1402),
			Code:    1402,
			Data:    params,
		}
	} else if data.Body.UserId <= 0 {
		return nil, &models.SystemError{
			Error:   err,
			Message: sdk.GetError(1403),
			Code:    1403,
			Data:    params,
		}
	}

	hp := &models.ChatMessageHistory{
		UserId:      data.Body.UserId,
		ChatId:      data.Body.ChatId,
		MessageId:   data.Body.MessageId,
		NewMessages: data.Body.NewMessages,
		UserType:    data.Body.UserType,
		Count:       data.Body.Count,
		Admin:       data.Body.Admin,
		Search:      data.Body.Search,
		Date:        data.Body.Date,
		OnlyOneChat: data.Body.OnlyOneChat,
	}

	ChatMessagesResponse, err := ws.hub.app.DB.GetMessagesHistory(hp)
	if err != nil {
		return nil, &models.SystemError{
			Error:   err,
			Message: err.Error(),
			Code:    database.MysqlErrorCode,
			Data:    params,
		}
	}

	ChatUsersResponse := []sdk.UserModel{}
	if len(ChatMessagesResponse) == 0 {
		ChatMessagesResponse = []sdk.ChatMessagesResponseDataItem{}
	}

	var chatIds []uint

	for _, item := range ChatMessagesResponse {
		chatIds = append(chatIds, item.ChatId)
	}

	opponents, err := ws.hub.app.DB.GetChatOpponents(chatIds, ws.hub.app.Sdk)
	if err != nil {
		return nil, &models.SystemError{
			Error:   err,
			Message: err.Error(),
			Code:    database.MysqlErrorCode,
			Data:    params,
		}
	}

	for _, item := range opponents {
		ChatUsersResponse = append(ChatUsersResponse, sdk.UserModel{
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
			Users:    ChatUsersResponse,
		},
	})
	if err != nil {
		return nil, service.MarshalError1011(err, params)
	}

	return result, nil
}

func (ws *WsServer) getChatRecent(params []byte) ([]byte, *models.SystemError) {
	data := &sdk.ChatMessagesRecentRequest{}
	err := json.Unmarshal(params, data)
	if err != nil {
		return nil, service.UnmarshalRequestError1201(err, params)
	}

	body := data.Body

	if body.ChatId <= 0 {
		return nil, &models.SystemError{
			Error:   nil,
			Message: sdk.GetError(1402),
			Code:    1402,
			Data:    params,
		}
	} else if body.UserId <= 0 {
		return nil, &models.SystemError{
			Error:   nil,
			Message: sdk.GetError(1403),
			Code:    1403,
			Data:    params,
		}
	}

	rp := &models.ChatMessageHistory{
		UserId:      body.UserId,
		ChatId:      body.ChatId,
		MessageId:   body.MessageId,
		NewMessages: body.NewMessages,
		UserType:    body.UserType,
		Admin:       body.Admin,
		Count:       body.Count,
		Search:      data.Body.Search,
		Date:        data.Body.Date,
	}

	ChatMessagesResponse, err := ws.hub.app.DB.GetMessagesRecent(rp)
	if err != nil {
		return nil, &models.SystemError{
			Error:   err,
			Message: err.Error(),
			Code:    database.MysqlErrorCode,
			Data:    params,
		}
	}

	ChatUsersResponse := []sdk.UserModel{}
	if len(ChatMessagesResponse) == 0 {
		ChatMessagesResponse = []sdk.ChatMessagesResponseDataItem{}
	}

	opponents, err := ws.hub.app.DB.GetChatOpponents(append([]uint{}, rp.ChatId), ws.hub.app.Sdk)
	if err != nil {
		return nil, &models.SystemError{
			Error:   err,
			Message: err.Error(),
			Code:    database.MysqlErrorCode,
			Data:    params,
		}
	}

	for _, item := range opponents {
		ChatUsersResponse = append(ChatUsersResponse, sdk.UserModel{
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
			Users:    ChatUsersResponse,
		},
	})
	if err != nil {
		return nil, service.MarshalError1011(err, params)
	}

	return result, nil
}

/**
SET
*/
func (ws *WsServer) setChatNew(params []byte) ([]byte, *models.SystemError) {
	data := &sdk.ChatNewRequest{}
	err := json.Unmarshal(params, data)
	if err != nil {
		return nil, service.UnmarshalRequestError1201(err, params)
	}

	chatId, err := ws.hub.app.DB.ChatCreate(data.Body.OrderId)
	if err != nil {
		return nil, &models.SystemError{
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
		return nil, service.MarshalError1011(err, params)
	}

	return result, nil
}

/**
SET
*/
func (ws *WsServer) setChatNewSubscribe(params []byte) ([]byte, *models.SystemError) {
	data := &sdk.ChatNewSubscribeRequest{}
	err := json.Unmarshal(params, data)
	if err != nil {
		return nil, service.UnmarshalRequestError1201(err, params)
	}

	chatId, err := ws.hub.app.DB.ChatCreate(data.Body.OrderId)
	if err != nil {
		return nil, &models.SystemError{
			Error:   err,
			Message: database.MysqlChatCreateError,
			Code:    database.MysqlChatCreateErrorCode,
			Data:    params,
		}
	}

	err = ws.hub.app.DB.SubscribeUser(
		chatId,
		data.Body.UserId,
		data.Body.UserType,
	)
	if err != nil {
		return nil, &models.SystemError{
			Error:   err,
			Message: database.MysqlChatUserSubscribeError + err.Error(),
			Code:    database.MysqlChatUserSubscribeErrorCode,
			Data:    params,
		}
	}

	//	subscribe websocket hub
	roomMessage := &RoomMessage{
		SendPush: false,
		UserId:   0,
		RoomId:   0,
		Message: &models.WSChatResponse{
			Type: service.SystemMsgTypeUserSubscribe,
			Data: &sdk.ChatUserSubscribeRequestBody{
				UserId:   data.Body.UserId,
				UserType: data.Body.UserType,
				ChatId:   chatId,
			},
		},
	}

	go ws.hub.SendMessageToRoom(roomMessage)

	response := sdk.ChatNewResponseData{ChatId: chatId}
	responseData := sdk.ChatNewResponse{Data: response}
	result, err := json.Marshal(responseData)
	if err != nil {
		return nil, service.MarshalError1011(err, params)
	}

	return result, nil
}

func (ws *WsServer) setChatUserSubscribe(params []byte) ([]byte, *models.SystemError) {
	data := &sdk.ChatUserSubscribeRequest{}
	err := json.Unmarshal(params, data)
	if err != nil {
		return nil, service.UnmarshalRequestError1201(err, params)
	}

	err = ws.hub.app.DB.SubscribeUser(
		data.Body.ChatId,
		data.Body.UserId,
		data.Body.UserType,
	)
	if err != nil {
		return nil, &models.SystemError{
			Error:   err,
			Message: database.MysqlChatUserSubscribeError + err.Error(),
			Code:    database.MysqlChatUserSubscribeErrorCode,
			Data:    params,
		}
	}

	//	subscribe websocket hub
	roomMessage := &RoomMessage{
		SendPush: false,
		UserId:   0,
		RoomId:   0,
		Message: &models.WSChatResponse{
			Type: service.SystemMsgTypeUserSubscribe,
			Data: data.Body,
		},
	}

	go ws.hub.SendMessageToRoom(roomMessage)

	result, err := json.Marshal(sdk.ChatUserSubscribeResponse{
		sdk.ChatUserSubscribeResponseData{Result: true},
	})
	if err != nil {
		return nil, service.MarshalError1011(err, params)
	}

	return result, nil
}

func (ws *WsServer) setChatUserUnsubscribe(params []byte) ([]byte, *models.SystemError) {
	data := &sdk.ChatUserUnsubscribeRequest{}
	err := json.Unmarshal(params, data)
	if err != nil {
		return nil, service.UnmarshalRequestError1201(err, params)
	}

	err = ws.hub.app.DB.UnsubscribeUser(data.Body.ChatId, data.Body.UserId)
	if err != nil {
		return nil, &models.SystemError{
			Error:   err,
			Message: database.MysqlChatUserUnsubscribeError,
			Code:    database.MysqlChatUserUnsubscribeErrorCode,
			Data:    params,
		}
	}

	//	unsubscribe websocket hub
	roomMessage := &RoomMessage{
		Message: &models.WSChatResponse{
			Type: service.SystemMsgTypeUserUnsubscribe,
			Data: data.Body,
		},
	}
	go ws.hub.SendMessageToRoom(roomMessage)

	result, err := json.Marshal(sdk.ChatUserSubscribeResponse{
		sdk.ChatUserSubscribeResponseData{Result: true},
	})

	if err != nil {
		return nil, service.MarshalError1011(err, params)
	}

	return result, nil
}

func (ws *WsServer) setChatMessage(params []byte) ([]byte, *models.SystemError) {
	data := &sdk.ChatMessageRequest{}
	chatMessageResponseData := &sdk.ChatMessageResponseData{}
	err := json.Unmarshal(params, data)
	if err != nil {
		return nil, service.UnmarshalRequestError1201(err, params)
	}

	loc, err := service.Location()
	if err != nil {
		return nil, &models.SystemError{
			Error:   err,
			Message: service.LoadLocationError,
			Code:    service.LoadLocationErrorCode,
		}
	}

	if !database.ValidateType(data.Body.Type) {
		return nil, &models.SystemError{
			Error:   err,
			Message: database.MysqlChatMessageTypeIncorrect,
			Code:    database.MysqlChatMessageTypeIncorrectCode,
			Data:    params,
		}
	}

	var userId uint
	var userType string
	var opponentsId []uint
	messageModel := &models.ChatMessage{
		ClientMessageId: data.Body.ClientMessageId,
		ChatId:          data.Body.ChatId,
		Type:            data.Body.Type,
		Message:         data.Body.Message,
		FileId:          data.Body.FileId,
	}

	subscribes := ws.hub.app.DB.ChatSubscribes(data.Body.ChatId)
	for _, subscribe := range subscribes {
		if subscribe.UserId == data.Body.UserId {
			userId = subscribe.UserId
			userType = subscribe.UserType
			messageModel.SubscribeId = subscribe.SubscribeId
		} else {
			opponentsId = append(opponentsId, subscribe.SubscribeId)
		}
	}

	if userId > 0 {
		err = ws.hub.app.DB.NewMessageTransact(messageModel, data.Body.Params, opponentsId)
		if err != nil {
			return nil, &models.SystemError{
				Error:   err,
				Message: database.MysqlChatCreateMessageError,
				Code:    database.MysqlChatCreateMessageErrorCode,
				Data:    params,
			}
		}

		messages := []interface{}{}
		clients := []sdk.UserModel{}
		messageParams, paramsErr := ws.hub.app.DB.GetParamsMap(messageModel.ID)
		if paramsErr != nil {
			return nil, &models.SystemError{
				Error:   paramsErr,
				Message: database.MysqlChatMessageParamsError,
				Code:    database.MysqlChatMessageParamsErrorCode,
				Data:    params,
			}
		}
		tmpMessageResponse := &models.WSChatMessagesDataMessageResponse{
			Id:              messageModel.ID,
			ClientMessageId: messageModel.ClientMessageId,
			InsertDate:      messageModel.CreatedAt.In(loc).Format(time.RFC3339),
			ChatId:          messageModel.ChatId,
			UserId:          data.Body.UserId,
			Sender:          userType,
			Status:          database.MessageStatusRecd,
			Type:            data.Body.Type,
			Text:            messageModel.Message,
			Params:          messageParams,
		}
		chatMessageResponseData = &sdk.ChatMessageResponseData{
			Id:              messageModel.ID,
			ClientMessageId: messageModel.ClientMessageId,
			InsertDate:      messageModel.CreatedAt.In(loc).Format(time.RFC3339),
			ChatId:          messageModel.ChatId,
			UserId:          data.Body.UserId,
			Sender:          userType,
			Status:          database.MessageStatusRecd,
			Type:            messageModel.Type,
			Text:            messageModel.Message,
			Params:          messageParams,
		}

		if len(messageModel.FileId) > 0 {
			file := &sdk.FileModel{Id: messageModel.FileId}
			sdkErr := ws.hub.app.Sdk.File(file, data.Body.ChatId, data.Body.UserId)
			if sdkErr != nil {
				return nil, &models.SystemError{
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

		userModel := &sdk.UserModel{
			Id: userId,
		}
		consultation := ws.hub.app.DB.Chat(data.Body.ChatId)
		err := ws.hub.app.Sdk.VagueUserById(userModel, userType, consultation.OrderId)
		if err != nil {
			return nil, &models.SystemError{
				Error:   err.Error,
				Code:    err.Code,
				Message: err.Message,
				Data:    err.Data,
			}
		}
		clients = append(clients, *userModel)

		responseData := &models.WSChatMessagesDataResponse{
			Messages: messages,
			Users:    clients,
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
		return nil, &models.SystemError{
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
		return nil, service.MarshalError1011(err, params)
	}

	return result, nil
}

func (ws *WsServer) setChatStatus(params []byte) ([]byte, *models.SystemError) {
	data := &sdk.ChatStatusRequest{}
	err := json.Unmarshal(params, data)
	if err != nil {
		return nil, service.UnmarshalRequestError1201(err, params)
	}

	err = ws.hub.app.DB.ChatChangeStatus(data.Body.ChatId, database.ChatStatusClosed)
	if err != nil {
		return nil, &models.SystemError{
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
		return nil, service.MarshalError1011(err, params)
	}

	return result, nil
}

func (ws *WsServer) sendClientMessage(params []byte) ([]byte, *models.SystemError) {
	data := &sdk.MessageToMobileClientRequest{}
	err := json.Unmarshal(params, data)
	if err != nil {
		return nil, service.UnmarshalRequestError1201(err, params)
	}

	if len(data.Body.Type) == 0 {
		return nil, &models.SystemError{
			Error:   err,
			Message: sdk.GetError(1404),
			Code:    1404,
			Data:    params,
		}
	}

	if data.Body.UserId <= 0 {
		return nil, &models.SystemError{
			Error:   err,
			Message: sdk.GetError(1403),
			Code:    1403,
			Data:    params,
		}
	}

	roomMessage := &RoomMessage{
		SendPush: true,
		RoomId:   0,
		UserId:   data.Body.UserId,
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
		return nil, service.MarshalError1011(err, params)
	}

	return result, nil
}

func (ws *WsServer) clientConsultationUpdate(params []byte) ([]byte, *models.SystemError) {
	data := &sdk.ClientConsultationUpdateRequest{}
	err := json.Unmarshal(params, data)
	if err != nil {
		return nil, service.UnmarshalRequestError1201(err, params)
	}

	if data.Body.UserId > 0 {
		response := &RoomMessage{
			UserId: data.Body.UserId,
			Message: &models.WSChatResponse{
				Type: EventConsultationUpdate,
				Data: data.Body.Data,
			},
		}

		ws.hub.SendMessageToRoom(response)
	} else {
		return nil, &models.SystemError{
			Message: sdk.GetError(1403),
			Code:    1403,
			Data:    params,
		}
	}

	result, err := json.Marshal(sdk.ChatStatusResponse{
		sdk.ChatStatusDataResponse{Result: true},
	})
	if err != nil {
		return nil, service.MarshalError1011(err, params)
	}

	return result, nil
}

func (ws *WsServer) changeChatUserSubscribe(params []byte) ([]byte, *models.SystemError) {
	data := &sdk.ChatUserSubscribeChangeRequest{}
	err := json.Unmarshal(params, data)
	if err != nil {
		return nil, service.UnmarshalRequestError1201(err, params)
	}

	if data.Body.ChatId == 0 {
		return nil, &models.SystemError{
			Error:   err,
			Message: sdk.GetError(1402),
			Code:    1402,
			Data:    params,
		}
	} else if data.Body.OldUserId == 0 || data.Body.NewUserId == 0 {
		return nil, &models.SystemError{
			Error:   err,
			Message: sdk.GetError(1403),
			Code:    1403,
			Data:    params,
		}
	}

	err = ws.hub.app.DB.SubscribeUserChange(data)
	if err != nil {
		return nil, &models.SystemError{
			Error:   err,
			Message: database.MysqlChatSubscribeChangeError,
			Code:    database.MysqlChatSubscribeChangeErrorCode,
			Data:    params,
		}
	}

	result, err := json.Marshal(&sdk.ChatUserSubscribeChangeResponse{
		Data: sdk.BoolResponseData{Result: true},
	})
	if err != nil {
		return nil, service.MarshalError1011(err, params)
	}

	return result, nil
}

package server

import (
	"chats/database"
	"chats/models"
	"chats/sdk"
	"chats/system"
	"encoding/json"
	uuid "github.com/satori/go.uuid"
	"log"
	"time"
)

const (
	EventEcho                  = "ping"
	EventMessage               = "message"
	EventJoin                  = "join"
	EventMessageStatus         = "messageStatus"
	EventTyping                = "typing"
	EventOpponentStatus        = "opponentStatus"
	EventClientConnectionError = "clientConnectionError"
)

const (
	EventTypingMessage = "печатает"

	UserStatusOnline  = "online"
	UserStatusOffline = "offline"
)

type Event struct{}

func NewEvent() *Event {
	return &Event{}
}

func (e *Event) EventMessage(h *Hub, c *Session, clientRequest []byte) {

	loc, err := system.Location()
	if err != nil {
		system.ErrHandler.SetError(&system.Error{
			Error:   err,
			Message: system.LoadLocationError,
			Code:    system.LoadLocationErrorCode,
		})
	}

	request := &models.WSChatMessagesRequest{}
	err = json.Unmarshal(clientRequest, request)
	if err != nil {
		system.ErrHandler.SetError(&system.Error{
			Error:   err,
			Message: system.UnmarshallingError,
			Code:    system.UnmarshallingErrorCode,
			Data:    clientRequest,
		})

		return
	}

	var roomId uuid.UUID
	var messages []interface{}
	accounts := make(map[uuid.UUID]sdk.Account)
	var subscribers []models.RoomSubscriber

	for _, item := range request.Data.Messages {
		if len(item.Text) > maxMessageSize {
			system.ErrHandler.SetError(&system.Error{
				Message: system.MessageTooLongError,
				Code:    system.MessageTooLongErrorCode,
				Data:    clientRequest,
			})
			return
		}
		if item.RoomId == uuid.Nil {
			system.ErrHandler.SetError(&system.Error{
				Message: database.MysqlChatIdIncorrect,
				Code:    database.MysqlChatIdIncorrectCode,
				Data:    clientRequest,
			})
			return
		}

		if roomId == uuid.Nil {
			roomId = item.RoomId
			subscribers, e := h.app.DB.GetRoomSubscribers(roomId)
			if e != nil {
				system.ErrHandler.SetError(e)
			}

			if len(subscribers) == 0 {
				system.ErrHandler.SetError(&system.Error{
					Error:   err,
					Message: database.MysqlChatSubscribeEmpty,
					Code:    database.MysqlChatSubscribeEmptyCode,
					Data:    clientRequest,
				})
				return
			}
		}

		var subscriberId = uuid.Nil
		var subscriberType string
		var opponentsId []uuid.UUID

		for _, subscriber := range subscribers {
			if _, ok := accounts[subscriber.AccountId]; !ok {

				account, err := h.app.DB.GetAccount(subscriber.AccountId, "")

				if err != nil {
					system.ErrHandler.SetError(err)
					return
				}
				accounts[subscriber.AccountId] = *ConvertAccountFromModel(account)
			}
			if subscriber.AccountId == c.account.Id {
				subscriberId = subscriber.Id
				subscriberType = subscriber.Role
			} else {
				opponentsId = append(opponentsId, subscriber.Id)
			}
		}

		if subscriberId == uuid.Nil {
			system.ErrHandler.SetError(&system.Error{
				Error:   err,
				Message: database.MysqlChatAccessDenied,
				Code:    database.MysqlChatAccessDeniedCode,
				Data:    clientRequest,
			})
			return
		}

		paramsJson, err := json.Marshal(item.Params)
		if err != nil {
			system.ErrHandler.SetError(&system.Error{Error: err})
		}

		id, _ := uuid.NewV4()
		dbMessage := &models.ChatMessage{
			Id:              id,
			ClientMessageId: item.ClientMessageId,
			ChatId:          roomId,
			Type:            item.Type,
			SubscribeId:     subscriberId,
			Message:         item.Text,
			Params:          string(paramsJson),
		}

		sentryErr := h.app.DB.NewMessageTransact(dbMessage, opponentsId)
		if sentryErr != nil {
			system.ErrHandler.SetError(sentryErr)
			return
		}

		tmpMessageResponse := &models.WSChatMessagesDataMessageResponse{
			Id:              dbMessage.Id,
			ClientMessageId: item.ClientMessageId,
			InsertDate:      dbMessage.CreatedAt.In(loc).Format(time.RFC3339),
			ChatId:          roomId,
			AccountId:       c.account.Id,
			Sender:          subscriberType,
			Status:          database.MessageStatusRecd,
			Type:            item.Type,
			Text:            item.Text,
		}
		if len(dbMessage.FileId) > 0 {
			file := &sdk.FileModel{Id: dbMessage.FileId}
			sdkErr := h.app.Sdk.File(file, roomId, c.account.Id)
			if sdkErr != nil {
				system.ErrHandler.SetError(&system.Error{
					Error:   sdkErr.Error,
					Message: sdkErr.Message,
					Code:    sdkErr.Code,
					Data:    sdkErr.Data,
				})
				return
			}
			tmpMessageResponseData := &models.WSChatMessagesDataMessageFileResponse{
				WSChatMessagesDataMessageResponse: *tmpMessageResponse,
				File:                              *file,
			}
			messages = append(messages, tmpMessageResponseData)
		} else {
			messages = append(messages, tmpMessageResponse)
		}
	}

	//	отправка обратно в веб-сокет
	if roomId != uuid.Nil {
		clients := []sdk.Account{}
		for _, item := range accounts {
			clients = append(clients, item)
		}

		responseMessage := &models.WSChatResponse{
			Type: EventMessage,
			Data: models.WSChatMessagesDataResponse{
				Messages: messages,
				Accounts: clients,
			},
		}

		response := &RoomMessage{
			RoomId:  roomId,
			Message: responseMessage,
		}

		h.SendMessageToRoom(response)
	}
}

func (e *Event) EventMessageStatus(h *Hub, c *Session, clientRequest []byte) {
	request := &models.WSChatMessageStatusRequest{}
	err := json.Unmarshal(clientRequest, request)
	if err != nil {
		system.UnmarshalRequestError1201(err, clientRequest)
		return
	}

	err = h.app.DB.SetReadStatus(request.Data.MessageId)

	if err != nil {
		system.ErrHandler.SetError(&system.Error{
			Error:   err,
			Message: WsChangeMessageStatusError,
			Code:    WsChangeMessageStatusErrorCode,
			Data:    clientRequest,
		})

		return
	}

	response := &RoomMessage{
		RoomId: request.Data.ChatId,
		Message: &models.WSChatResponse{
			Type: EventMessageStatus,
			Data: models.WSChatMessageStatusDataResponse{
				Status:    database.MessageStatusRead,
				ChatId:    request.Data.ChatId,
				MessageId: request.Data.MessageId,
			},
		},
	}

	h.SendMessageToRoom(response)
}

func (e *Event) EventOpponentStatus(h *Hub, c *Session, clientRequest []byte) {
	request := &models.WSChatOpponentStatusRequest{}
	err := json.Unmarshal(clientRequest, request)
	if err != nil {
		system.ErrHandler.SetError(system.UnmarshalRequestError1201(err, clientRequest))
		return
	}

	roomId := request.Data.RoomId
	subscribes, error := h.app.DB.GetRoomSubscribers(roomId)
	if error != nil {
		system.ErrHandler.SetError(error)
	}

	accounts := []models.AccountStatusModel{}

	for _, subscribe := range subscribes {

		account := &models.AccountStatusModel{AccountId: subscribe.AccountId}

		if c.account.Id != subscribe.AccountId && subscribe.Role != UserTypeBot {

			cronGetUserStatusRequestMessage := &models.CronGetUserStatusRequest{
				Type: MessageTypeGetUserStatus,
				Data: models.CronGetAccountStatusRequestData{
					AccountId: subscribe.AccountId,
				},
			}

			request, err := json.Marshal(cronGetUserStatusRequestMessage)
			if err != nil {
				system.ErrHandler.SetError(system.MarshalError1011(err, clientRequest))
				return
			}

			response, cronRequestError := h.app.Sdk.
				Subject(system.CronTopic()).
				Request(request)

			if cronRequestError != nil {
				system.ErrHandler.SetError(&system.Error{
					Error:   cronRequestError.Error,
					Message: system.CronResponseError,
					Code:    system.CronResponseErrorCode,
					Data:    append(request, response...),
				})
				return
			}

			cronGetUserResponse := &models.CronGetAccountStatusResponse{}
			err = json.Unmarshal(response, cronGetUserResponse)
			if err != nil {
				system.ErrHandler.SetError(system.MarshalError1011(err, response))
				return
			}

			if cronGetUserResponse.Online {
				account.Status = UserStatusOnline
			} else {
				account.Status = UserStatusOffline
			}
		}

		accounts = append(accounts, *account)
	}

	response := &RoomMessage{
		RoomId: roomId,
		Message: &models.WSChatResponse{
			Type: EventOpponentStatus,
			Data: models.WSChatOpponentStatusDataResponse{
				RoomId:   roomId,
				Accounts: accounts,
			},
		},
	}

	h.SendMessageToRoom(response)

	return
}

func (e *Event) EventJoin(h *Hub, c *Session, clientRequest []byte) {
	request := &models.WSChatJoinRequest{}
	err := json.Unmarshal(clientRequest, request)
	if err != nil {
		system.UnmarshalRequestError1201(err, clientRequest)
		return
	}

	//go h.app.Sdk.UserConsultationJoin(request.Data.ConsultationId, c.account.Id)

	return
}

func (e *Event) EventTyping(h *Hub, c *Session, clientRequest []byte) {
	request := &models.WSChatTypingRequest{}
	err := json.Unmarshal(clientRequest, request)
	if err != nil {
		system.UnmarshalRequestError1201(err, clientRequest)
		return
	}

	if room, ok := h.rooms[request.Data.RoomId]; ok {
		for _, subscriber := range room.subscribers {
			if subscriber.AccountId != c.account.Id {

				message := EventTypingMessage

				response := &RoomMessage{
					AccountId: subscriber.AccountId,
					Message: &models.WSChatResponse{
						Type: EventTyping,
						Data: models.WSChatTypingDataResponse{
							AccountId: c.account.Id,
							Message:   message,
							Status:    request.Data.Status,
						},
					},
				}

				h.SendMessageToRoom(response)

				return
			}
		}
	}

	return
}

func (e *Event) EventEcho(h *Hub, c *Session, clientRequest []byte) {
	request := &models.WSChatEchoRequest{}
	err := json.Unmarshal(clientRequest, request)
	if err != nil {
		system.UnmarshalRequestError1201(err, clientRequest)
		return
	}

	message := &models.WSChatResponse{
		Type: "pong",
		Data: models.WSChatEchoResponse{},
	}

	response, err := json.Marshal(message)
	if err != nil {
		log.Println("Не удалось сформировать ответ клиенту")
		log.Println(err)
	} else {
		h.sendMessage(c, response)
	}

	return
}

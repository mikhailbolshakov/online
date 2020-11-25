package server

import (
	"chats/app"
	a "chats/repository/account"
	r "chats/repository/room"
	"chats/system"
	"encoding/json"
	uuid "github.com/satori/go.uuid"
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

	roomRepository := r.CreateRepository(app.GetDB())
	accountRepository := a.CreateRepository(app.GetDB())

	loc, err := app.Instance.GetLocation()
	if err != nil {
		app.E().SetError(&system.Error{
			Error:   err,
			Message: system.GetError(system.LoadLocationErrorCode),
			Code:    system.LoadLocationErrorCode,
		})
	}

	request := &WSChatMessagesRequest{}
	err = json.Unmarshal(clientRequest, request)
	if err != nil {
		app.E().SetError(&system.Error{
			Error:   err,
			Message: system.GetError(system.UnmarshallingErrorCode),
			Code:    system.UnmarshallingErrorCode,
			Data:    clientRequest,
		})

		return
	}

	app.L().Debugf("Handler[EventMessage]. Request: %v \n", request)

	var roomId uuid.UUID
	var messages []interface{}
	accounts := make(map[uuid.UUID]Account)
	var subscribers []r.RoomSubscriber
	var sysErr = &system.Error{}

	for _, item := range request.Data.Messages {
		if len(item.Text) > maxMessageSize {
			app.E().SetError(&system.Error{
				Message: system.GetError(system.MessageTooLongErrorCode),
				Code:    system.MessageTooLongErrorCode,
				Data:    clientRequest,
			})
			return
		}
		if item.RoomId == uuid.Nil {
			app.E().SetError(&system.Error{
				Message: system.GetError(system.MysqlChatIdIncorrectCode),
				Code:    system.MysqlChatIdIncorrectCode,
				Data:    clientRequest,
			})
			return
		}

		if roomId == uuid.Nil {
			roomId = item.RoomId
			subscribers, sysErr = roomRepository.GetRoomSubscribers(roomId)
			if sysErr != nil {
				app.E().SetError(sysErr)
			}

			if len(subscribers) == 0 {
				app.E().SetError(&system.Error{
					Error:   err,
					Message: system.GetError(system.MysqlChatSubscribeEmptyCode),
					Code:    system.MysqlChatSubscribeEmptyCode,
					Data:    clientRequest,
				})
				return
			}
		}

		var subscriberId = uuid.Nil
		var accountId = uuid.Nil
		var subscriberType string
		var opponents []r.ChatOpponent

		for _, s := range subscribers {
			if _, ok := accounts[s.AccountId]; !ok {

				account, err := accountRepository.GetAccount(s.AccountId, "")

				if err != nil {
					app.E().SetError(err)
					return
				}
				accounts[s.AccountId] = *ConvertAccountFromModel(account)
			}
			if s.AccountId == c.account.Id {
				accountId = s.AccountId
				subscriberId = s.Id
				subscriberType = s.Role
			} else {
				opponents = append(opponents, r.ChatOpponent{
					SubscriberId: s.Id,
					AccountId:    s.AccountId,
				})
			}
		}

		if subscriberId == uuid.Nil {
			app.E().SetError(&system.Error{
				Error:   err,
				Message: system.GetError(system.MysqlChatAccessDeniedCode),
				Code:    system.MysqlChatAccessDeniedCode,
				Data:    clientRequest,
			})
			return
		}

		paramsJson, err := json.Marshal(item.Params)
		if err != nil {
			app.E().SetError(&system.Error{Error: err})
		}

		dbMessage := &r.ChatMessage{
			Id:              system.Uuid(),
			ClientMessageId: item.ClientMessageId,
			RoomId:          roomId,
			AccountId:       accountId,
			Type:            item.Type,
			SubscribeId:     subscriberId,
			Message:         item.Text,
			Params:          string(paramsJson),
		}

		sysErr = roomRepository.CreateMessage(dbMessage, opponents)
		if sysErr != nil {
			app.E().SetError(sysErr)
			return
		}

		tmpMessageResponse := &WSChatMessagesDataMessageResponse{
			Id:              dbMessage.Id,
			ClientMessageId: item.ClientMessageId,
			InsertDate:      dbMessage.CreatedAt.In(loc).Format(time.RFC3339),
			ChatId:          roomId,
			AccountId:       c.account.Id,
			Sender:          subscriberType,
			Status:          r.MessageStatusRecd,
			Type:            item.Type,
			Text:            item.Text,
		}
		if len(dbMessage.FileId) > 0 {
			//file := &FileModel{Id: dbMessage.FileId}
			//sdkErr := h.inf.Nats.File(file, roomId, c.account.Id)
			//if sdkErr != nil {
			//	app.E().SetError(&system.Error{
			//		Error:   sdkErr.Error,
			//		Message: sdkErr.Message,
			//		Code:    sdkErr.Code,
			//		Data:    sdkErr.Data,
			//	})
			//	return
			//}
			tmpMessageResponseData := &WSChatMessagesDataMessageFileResponse{
				WSChatMessagesDataMessageResponse: *tmpMessageResponse,
				//File:                              nil,
			}
			messages = append(messages, tmpMessageResponseData)
		} else {
			messages = append(messages, tmpMessageResponse)
		}
	}

	//	send back to the socket
	if roomId != uuid.Nil {
		clients := []Account{}
		for _, item := range accounts {
			clients = append(clients, item)
		}

		responseMessage := &WSChatResponse{
			Type: EventMessage,
			Data: WSChatMessagesDataResponse{
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

	request := &WSChatMessageStatusRequest{}
	err := json.Unmarshal(clientRequest, request)
	if err != nil {
		system.UnmarshalRequestError1201(err, clientRequest)
		return
	}

	rep := r.CreateRepository(app.GetDB())
	sysErr := rep.SetReadStatus(request.Data.MessageId, c.account.Id)

	if sysErr != nil {
		app.E().SetError(system.SysErr(err, system.WsChangeMessageStatusErrorCode, clientRequest))
		return
	}

	response := &RoomMessage{
		RoomId: request.Data.RoomId,
		Message: &WSChatResponse{
			Type: EventMessageStatus,
			Data: WSChatMessageStatusDataResponse{
				Status:    r.MessageStatusRead,
				RoomId:    request.Data.RoomId,
				MessageId: request.Data.MessageId,
			},
		},
	}

	h.SendMessageToRoom(response)
}

func (e *Event) EventOpponentStatus(h *Hub, c *Session, clientRequest []byte) {

	rep := r.CreateRepository(app.GetDB())

	request := &WSChatOpponentStatusRequest{}
	err := json.Unmarshal(clientRequest, request)
	if err != nil {
		app.E().SetError(system.UnmarshalRequestError1201(err, clientRequest))
		return
	}

	roomId := request.Data.RoomId
	subscribes, error := rep.GetRoomSubscribers(roomId)
	if error != nil {
		app.E().SetError(error)
	}

	accounts := []WSAccountStatusModel{}

	for _, subscribe := range subscribes {

		account := &WSAccountStatusModel{AccountId: subscribe.AccountId}

		if c.account.Id != subscribe.AccountId && subscribe.SystemAccount == 1 {

			cronGetUserStatusRequestMessage := &CronGetUserStatusRequest{
				Type: MessageTypeGetUserStatus,
				Data: CronGetAccountStatusRequestData{
					AccountId: subscribe.AccountId,
				},
			}

			request, err := json.Marshal(cronGetUserStatusRequestMessage)
			if err != nil {
				app.E().SetError(system.MarshalError1011(err, clientRequest))
				return
			}

			response, cronRequestError := app.GetNats().
				Subject(app.GetNats().CronTopic()).
				Request(request)

			if cronRequestError != nil {
				app.E().SetError(&system.Error{
					Error:   cronRequestError.Error,
					Message: system.GetError(system.CronResponseErrorCode),
					Code:    system.CronResponseErrorCode,
					Data:    append(request, response...),
				})
				return
			}

			cronGetUserResponse := &CronGetAccountStatusResponse{}
			err = json.Unmarshal(response, cronGetUserResponse)
			if err != nil {
				app.E().SetError(system.MarshalError1011(err, response))
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
		Message: &WSChatResponse{
			Type: EventOpponentStatus,
			Data: WSChatOpponentStatusDataResponse{
				RoomId:   roomId,
				Accounts: accounts,
			},
		},
	}

	h.SendMessageToRoom(response)

	return
}

func (e *Event) EventJoin(h *Hub, c *Session, clientRequest []byte) {
	request := &WSChatJoinRequest{}
	err := json.Unmarshal(clientRequest, request)
	if err != nil {
		system.UnmarshalRequestError1201(err, clientRequest)
		return
	}

	//go h.inf.Nats.UserConsultationJoin(request.Data.ConsultationId, c.account.Id)

	return
}

func (e *Event) EventTyping(h *Hub, c *Session, clientRequest []byte) {
	request := &WSChatTypingRequest{}
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
					Message: &WSChatResponse{
						Type: EventTyping,
						Data: WSChatTypingDataResponse{
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
	request := &WSChatEchoRequest{}
	err := json.Unmarshal(clientRequest, request)
	if err != nil {
		system.UnmarshalRequestError1201(err, clientRequest)
		return
	}

	message := &WSChatResponse{
		Type: "pong",
		Data: WSChatEchoResponse{},
	}

	response, err := json.Marshal(message)
	if err != nil {
		app.L().Debug("Не удалось сформировать ответ клиенту")
		app.L().Debug(err)
	} else {
		h.sendMessage(c, response)
	}

	return
}

package server

import (
	"chats/database"
	"chats/infrastructure"
	"chats/models"
	"chats/sdk"
	"chats/sentry"
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
	EventConsultationUpdate    = "consultationUpdate"
	EventClientConnectionError = "clientConnectionError"
)

const (
	EventTypingOperatorMessage = "печатает"
	EventTypingDoctorMessage   = "печатает"
	EventTypingClientMessage   = "Пациент пишет ответ"

	UserStatusOnline  = "online"
	UserStatusOffline = "offline"
)

type Event struct{}

func NewEvent() *Event {
	return &Event{}
}

func (e *Event) EventMessage(h *Hub, c *Client, clientRequest []byte) {
	loc, err := infrastructure.Location()
	if err != nil {
		infrastructure.SetError(&sentry.SystemError{
			Error:   err,
			Message: infrastructure.LoadLocationError,
			Code:    infrastructure.LoadLocationErrorCode,
		})
	}

	request := &models.WSChatMessagesRequest{}
	err = json.Unmarshal(clientRequest, request)
	if err != nil {
		infrastructure.SetError(&sentry.SystemError{
			Error:   err,
			Message: infrastructure.UnmarshallingError,
			Code:    infrastructure.UnmarshallingErrorCode,
			Data:    clientRequest,
		})

		return
	}

	var chatId uuid.UUID
	messages := []interface{}{}
	accounts := make(map[uuid.UUID]sdk.AccountModel)
	subscribers := []database.SubscribeAccountModel{}

	for _, item := range request.Data.Messages {
		if len(item.Text) > maxMessageSize {
			infrastructure.SetError(&sentry.SystemError{
				Message: infrastructure.MessageTooLongError,
				Code:    infrastructure.MessageTooLongErrorCode,
				Data:    clientRequest,
			})
			return
		}
		if item.ChatId == uuid.Nil {
			infrastructure.SetError(&sentry.SystemError{
				Message: database.MysqlChatIdIncorrect,
				Code:    database.MysqlChatIdIncorrectCode,
				Data:    clientRequest,
			})
			return
		}

		if chatId == uuid.Nil {
			chatId = item.ChatId
			subscribers = h.app.DB.ChatSubscribes(chatId)

			if len(subscribers) == 0 {
				infrastructure.SetError(&sentry.SystemError{
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
		opponentsId := []uuid.UUID{}

		for _, subscriber := range subscribers {
			if _, ok := accounts[subscriber.AccountId]; !ok {
				user := &sdk.AccountModel{
					Id: subscriber.AccountId,
				}
				chat := h.app.DB.Chat(chatId)
				err := h.app.Sdk.VagueUserById(user, subscriber.Role, chat.ReferenceId)
				if err != nil {
					infrastructure.SetError(&sentry.SystemError{
						Error:   err.Error,
						Message: err.Message,
						Code:    err.Code,
						Data:    err.Data,
					})
					return
				}
				accounts[subscriber.AccountId] = *user
			}
			if subscriber.AccountId == c.account.Id {
				subscriberId = subscriber.SubscribeId
				subscriberType = subscriber.Role
			} else {
				opponentsId = append(opponentsId, subscriber.SubscribeId)
			}
		}

		if subscriberId == uuid.Nil {
			infrastructure.SetError(&sentry.SystemError{
				Error:   err,
				Message: database.MysqlChatAccessDenied,
				Code:    database.MysqlChatAccessDeniedCode,
				Data:    clientRequest,
			})
			return
		}

		dbMessage := &models.ChatMessage{
			ClientMessageId: item.ClientMessageId,
			ChatId:          chatId,
			Type:            item.Type,
			SubscribeId:     subscriberId,
			Message:         item.Text,
		}

		err = h.app.DB.NewMessageTransact(dbMessage, item.Params, opponentsId)
		if err != nil {
			infrastructure.SetError(&sentry.SystemError{
				Error:   err,
				Message: database.MysqlChatCreateMessageError,
				Code:    database.MysqlChatCreateMessageErrorCode,
				Data:    clientRequest,
			})
			return
		}

		tmpMessageResponse := &models.WSChatMessagesDataMessageResponse{
			Id:              dbMessage.Id,
			ClientMessageId: item.ClientMessageId,
			InsertDate:      dbMessage.CreatedAt.In(loc).Format(time.RFC3339),
			ChatId:          chatId,
			AccountId:       c.account.Id,
			Sender:          subscriberType,
			Status:          database.MessageStatusRecd,
			Type:            item.Type,
			Text:            item.Text,
		}
		if len(dbMessage.FileId) > 0 {
			file := &sdk.FileModel{Id: dbMessage.FileId}
			sdkErr := h.app.Sdk.File(file, chatId, c.account.Id)
			if sdkErr != nil {
				infrastructure.SetError(&sentry.SystemError{
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
	if chatId != uuid.Nil {
		clients := []sdk.AccountModel{}
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
			RoomId:  chatId,
			Message: responseMessage,
		}

		h.SendMessageToRoom(response)
	}
}

func (e *Event) EventMessageStatus(h *Hub, c *Client, clientRequest []byte) {
	request := &models.WSChatMessageStatusRequest{}
	err := json.Unmarshal(clientRequest, request)
	if err != nil {
		infrastructure.UnmarshalRequestError1201(err, clientRequest)
		return
	}

	err = h.app.DB.SetReadStatus(request.Data.MessageId)

	if err != nil {
		infrastructure.SetError(&sentry.SystemError{
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

func (e *Event) EventOpponentStatus(h *Hub, c *Client, clientRequest []byte) {
	request := &models.WSChatOpponentStatusRequest{}
	err := json.Unmarshal(clientRequest, request)
	if err != nil {
		infrastructure.SetError(infrastructure.UnmarshalRequestError1201(err, clientRequest))
		return
	}

	chatId := request.Data.ChatId
	subscribes := h.app.DB.ChatSubscribes(chatId)
	users := []models.UserStatusModel{}

	for _, subscribe := range subscribes {
		user := &models.UserStatusModel{AccountId: subscribe.AccountId}

		switch subscribe.Role {
		case database.UserTypeClient:
			cronGetUserStatusRequestMessage := &models.CronGetUserStatusRequest{
				Type: MessageTypeGetUserStatus,
				Data: models.CronGetAccountStatusRequestData{
					AccountId: subscribe.AccountId,
				},
			}

			request, err := json.Marshal(cronGetUserStatusRequestMessage)
			if err != nil {
				infrastructure.SetError(infrastructure.MarshalError1011(err, clientRequest))
				return
			}

			response, cronRequestError := h.app.Sdk.
				Subject(infrastructure.CronTopic()).
				Request(request)

			if cronRequestError != nil {
				infrastructure.SetError(&sentry.SystemError{
					Error:   cronRequestError.Error,
					Message: infrastructure.CronResponseError,
					Code:    infrastructure.CronResponseErrorCode,
					Data:    append(request, response...),
				})
				return
			}

			cronGetUserResponse := &models.CronGetAccountStatusResponse{}
			err = json.Unmarshal(response, cronGetUserResponse)
			if err != nil {
				infrastructure.SetError(infrastructure.MarshalError1011(err, response))
				return
			}

			if cronGetUserResponse.Online {
				user.Status = UserStatusOnline
			} else {
				user.Status = UserStatusOffline
			}
			break
		default:
			//	оператор и доктор всегда в сети
			user.Status = UserStatusOnline
			break
		}

		users = append(users, *user)
	}

	response := &RoomMessage{
		RoomId: chatId,
		Message: &models.WSChatResponse{
			Type: EventOpponentStatus,
			Data: models.WSChatOpponentStatusDataResponse{
				ChatId:   chatId,
				Accounts: users,
			},
		},
	}

	h.SendMessageToRoom(response)

	return
}

func (e *Event) EventJoin(h *Hub, c *Client, clientRequest []byte) {
	request := &models.WSChatJoinRequest{}
	err := json.Unmarshal(clientRequest, request)
	if err != nil {
		infrastructure.UnmarshalRequestError1201(err, clientRequest)
		return
	}

	go h.app.Sdk.UserConsultationJoin(request.Data.ConsultationId, c.account.Id)

	return
}

func (e *Event) EventTyping(h *Hub, c *Client, clientRequest []byte) {
	request := &models.WSChatTypingRequest{}
	err := json.Unmarshal(clientRequest, request)
	if err != nil {
		infrastructure.UnmarshalRequestError1201(err, clientRequest)
		return
	}

	if chat, ok := h.rooms[request.Data.ChatId]; ok {
		for _, subscriber := range chat.subscribers {
			if subscriber.AccountId == c.account.Id {
				var message string

				switch subscriber.Role {
				case database.UserTypeOperator:
					message = EventTypingOperatorMessage
					break
				case database.UserTypeDoctor:
					message = EventTypingDoctorMessage
					break
				case database.UserTypeClient:
					message = EventTypingClientMessage
					break
				}

				if len(message) > 0 {
					for _, subscriber := range chat.subscribers {
						//	Не отправлять сообщение о тайпенге пользователю, который этот тайпинг совершает
						if subscriber.AccountId != c.account.Id {
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
						}
					}
				}

				return
			}
		}
	}

	return
}

func (e *Event) EventEcho(h *Hub, c *Client, clientRequest []byte) {
	request := &models.WSChatEchoRequest{}
	err := json.Unmarshal(clientRequest, request)
	if err != nil {
		infrastructure.UnmarshalRequestError1201(err, clientRequest)
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
		h.sendMessageToClient(c, response)
	}

	return
}

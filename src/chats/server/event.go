package server

import (
	"chats/database"
	"chats/models"
	"chats/service"
	"encoding/json"
	"gitlab.medzdrav.ru/health-service/go-sdk"
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
	loc, err := service.Location()
	if err != nil {
		service.SetError(&models.SystemError{
			Error:   err,
			Message: service.LoadLocationError,
			Code:    service.LoadLocationErrorCode,
		})
	}

	request := &models.WSChatMessagesRequest{}
	err = json.Unmarshal(clientRequest, request)
	if err != nil {
		service.SetError(&models.SystemError{
			Error:   err,
			Message: service.UnmarshallingError,
			Code:    service.UnmarshallingErrorCode,
			Data:    clientRequest,
		})

		return
	}

	var chatId uint
	messages := []interface{}{}
	users := make(map[uint]sdk.UserModel)
	subscribers := []database.SubscribeUserModel{}

	for _, item := range request.Data.Messages {
		if len(item.Text) > maxMessageSize {
			service.SetError(&models.SystemError{
				Message: service.MessageTooLongError,
				Code:    service.MessageTooLongErrorCode,
				Data:    clientRequest,
			})
			return
		}
		if item.ChatId == 0 {
			service.SetError(&models.SystemError{
				Message: database.MysqlChatIdIncorrect,
				Code:    database.MysqlChatIdIncorrectCode,
				Data:    clientRequest,
			})
			return
		}

		if chatId == 0 {
			chatId = item.ChatId
			subscribers = h.app.DB.ChatSubscribes(chatId)

			if len(subscribers) == 0 {
				service.SetError(&models.SystemError{
					Error:   err,
					Message: database.MysqlChatSubscribeEmpty,
					Code:    database.MysqlChatSubscribeEmptyCode,
					Data:    clientRequest,
				})
				return
			}
		}

		var subscriberId uint = 0
		var subscriberType string
		opponentsId := []uint{}

		for _, subscriber := range subscribers {
			if _, ok := users[subscriber.UserId]; !ok {
				user := &sdk.UserModel{
					Id: subscriber.UserId,
				}
				consultation := h.app.DB.Chat(chatId)
				err := h.app.Sdk.VagueUserById(user, subscriber.UserType, consultation.OrderId)
				if err != nil {
					service.SetError(&models.SystemError{
						Error:   err.Error,
						Message: err.Message,
						Code:    err.Code,
						Data:    err.Data,
					})
					return
				}
				users[subscriber.UserId] = *user
			}
			if subscriber.UserId == c.user.Id {
				subscriberId = subscriber.SubscribeId
				subscriberType = subscriber.UserType
			} else {
				opponentsId = append(opponentsId, subscriber.SubscribeId)
			}
		}

		if subscriberId == 0 {
			service.SetError(&models.SystemError{
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
			service.SetError(&models.SystemError{
				Error:   err,
				Message: database.MysqlChatCreateMessageError,
				Code:    database.MysqlChatCreateMessageErrorCode,
				Data:    clientRequest,
			})
			return
		}

		tmpMessageResponse := &models.WSChatMessagesDataMessageResponse{
			Id:              dbMessage.ID,
			ClientMessageId: item.ClientMessageId,
			InsertDate:      dbMessage.CreatedAt.In(loc).Format(time.RFC3339),
			ChatId:          chatId,
			UserId:          c.user.Id,
			Sender:          subscriberType,
			Status:          database.MessageStatusRecd,
			Type:            item.Type,
			Text:            item.Text,
		}
		if len(dbMessage.FileId) > 0 {
			file := &sdk.FileModel{Id: dbMessage.FileId}
			sdkErr := h.app.Sdk.File(file, chatId, c.user.Id)
			if sdkErr != nil {
				service.SetError(&models.SystemError{
					Error:   sdkErr.Error,
					Message: sdkErr.Message,
					Code:    sdkErr.Code,
					Data:    sdkErr.Data,
				})
				return
			}
			tmpMessageResponseData := &models.WSChatMessagesDataMessageFileResponse{
				WSChatMessagesDataMessageResponse: *tmpMessageResponse,
				File: *file,
			}
			messages = append(messages, tmpMessageResponseData)
		} else {
			messages = append(messages, tmpMessageResponse)
		}
	}

	//	отправка обратно в веб-сокет
	if chatId > 0 {
		clients := []sdk.UserModel{}
		for _, item := range users {
			clients = append(clients, item)
		}

		responseMessage := &models.WSChatResponse{
			Type: EventMessage,
			Data: models.WSChatMessagesDataResponse{
				Messages: messages,
				Users:    clients,
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
		service.UnmarshalRequestError1201(err, clientRequest)
		return
	}

	err = h.app.DB.SetReadStatus(request.Data.MessageId)

	if err != nil {
		service.SetError(&models.SystemError{
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
		service.SetError(service.UnmarshalRequestError1201(err, clientRequest))
		return
	}

	chatId := request.Data.ChatId
	subscribes := h.app.DB.ChatSubscribes(chatId)
	users := []models.UserStatusModel{}

	for _, subscribe := range subscribes {
		user := &models.UserStatusModel{UserId: subscribe.UserId}

		switch subscribe.UserType {
		case database.UserTypeClient:
			cronGetUserStatusRequestMessage := &models.CronGetUserStatusRequest{
				Type: MessageTypeGetUserStatus,
				Data: models.CronGetUserStatusRequestData{
					UserId: subscribe.UserId,
				},
			}

			request, err := json.Marshal(cronGetUserStatusRequestMessage)
			if err != nil {
				service.SetError(service.MarshalError1011(err, clientRequest))
				return
			}

			response, cronRequestError := h.app.Sdk.
				Subject(service.CronTopic()).
				Request(request)

			if cronRequestError != nil {
				service.SetError(&models.SystemError{
					Error:   cronRequestError.Error,
					Message: service.CronResponseError,
					Code:    service.CronResponseErrorCode,
					Data:    append(request, response...),
				})
				return
			}

			cronGetUserResponse := &models.CronGetUserStatusResponse{}
			err = json.Unmarshal(response, cronGetUserResponse)
			if err != nil {
				service.SetError(service.MarshalError1011(err, response))
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
				ChatId: chatId,
				Users:  users,
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
		service.UnmarshalRequestError1201(err, clientRequest)
		return
	}

	go h.app.Sdk.UserConsultationJoin(request.Data.ConsultationId, c.user.Id)

	return
}

func (e *Event) EventTyping(h *Hub, c *Client, clientRequest []byte) {
	request := &models.WSChatTypingRequest{}
	err := json.Unmarshal(clientRequest, request)
	if err != nil {
		service.UnmarshalRequestError1201(err, clientRequest)
		return
	}

	if chat, ok := h.rooms[request.Data.ChatId]; ok {
		for _, subscriber := range chat.subscribers {
			if subscriber.UserId == c.user.Id {
				var message string

				switch subscriber.UserType {
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
						if subscriber.UserId != c.user.Id {
							response := &RoomMessage{
								UserId: subscriber.UserId,
								Message: &models.WSChatResponse{
									Type: EventTyping,
									Data: models.WSChatTypingDataResponse{
										UserId:  c.user.Id,
										Message: message,
										Status:  request.Data.Status,
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
		service.UnmarshalRequestError1201(err, clientRequest)
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

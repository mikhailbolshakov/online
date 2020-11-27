package server

import (
	"chats/app"
	r "chats/repository/room"
	"chats/system"
	"encoding/json"
	uuid "github.com/satori/go.uuid"
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

	defer app.E().CatchPanic("EventMessage")

	clRq := &WSChatMessagesRequest{}
	err := json.Unmarshal(clientRequest, clRq)
	if err != nil {
		app.E().SetError(system.SysErr(err, system.UnmarshallingErrorCode, clientRequest))
		return
	}

	request := &SendChatMessagesRequest{
		Type:            clRq.Type,
		Data:            SendChatMessagesDataRequest{
			Messages: []SendChatMessageDataRequest{},
		},
	}

	if clRq.SenderAccountId != uuid.Nil {
		request.SenderAccountId = clRq.SenderAccountId
	} else {
		request.SenderAccountId = c.account.Id
	}

	for _, m := range clRq.Data.Messages {
		request.Data.Messages = append(request.Data.Messages, SendChatMessageDataRequest{
			ClientMessageId:    m.ClientMessageId,
			RoomId:             m.RoomId,
			Type:               m.Type,
			Text:               m.Text,
			Params:             m.Params,
			RecipientAccountId: m.RecipientAccountId,
		})
	}

	_, srvErr := wsServer.SendChatMessages(request)
	if srvErr != nil {
		app.E().SetError(srvErr)
	}

}

func (e *Event) EventMessageStatus(h *Hub, c *Session, clientRequest []byte) {

	defer app.E().CatchPanic("EventMessageStatus")

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

	defer app.E().CatchPanic("EventOpponentStatus")

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

		if c.account.Id != subscribe.AccountId && !system.Uint8ToBool(subscribe.SystemAccount) {

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
				app.E().SetError(system.SysErr(cronRequestError.Error, system.CronResponseErrorCode, append(request, response...)))
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

	defer app.E().CatchPanic("EventJoin")

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

	defer app.E().CatchPanic("EventTyping")

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

	defer app.E().CatchPanic("EventEcho")

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

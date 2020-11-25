package server

import (
	"chats/app"
	r "chats/repository/room"
	"chats/system"
	"encoding/json"
	uuid "github.com/satori/go.uuid"
	"time"
)

const (
	MessageTypeGetUserStatus   = "getUserStatus"
	MessageTypeSendOnlineUsers = "sendOnlineUsers"
)



func (ws *WsServer) consumer() {
	errorChan := make(chan *system.Error, 2048)

	nats := app.GetNats()
	go app.GetNats().
		Subject(nats.CronTopic()).
		CronConsumer(ws.cronMessagesHandler, errorChan)

	ticker := time.NewTicker(time.Second * 5).C

	for {
		select {
		case <-ticker:
			now := time.Now()
			go func() {
				ws.actualAccountsMutex.Lock()
				defer ws.actualAccountsMutex.Unlock()
				for userId, expiredTime := range ws.actualAccounts {
					if now.After(expiredTime) {
						app.L().Debug(userId, "is offline") //	todo
						delete(ws.actualAccounts, userId)
					}
				}
			}()

		case err := <-errorChan:
			app.E().SetError(&system.Error{
				Error:   err.Error,
				Message: err.Message,
				Code:    err.Code,
				Data:    err.Data,
			})
		}
	}
}

func (ws *WsServer) provider() {
	for {
		time.Sleep(4 * time.Second)

		actualAccountIds := make([]uuid.UUID, 0, len(ws.hub.accounts))

		for accountId := range ws.hub.accounts {
			actualAccountIds = append(actualAccountIds, accountId)
		}

		cronMessage := &CronSendOnlineUsers{
			Type: MessageTypeSendOnlineUsers,
			Data: CronSendOnlineAccountsData{
				Accounts: actualAccountIds,
			},
		}

		request, err := json.Marshal(cronMessage)
		if err != nil {
			system.MarshalError1011(err, nil)
		} else {
			nats := app.GetNats()
			go nats.
				Subject(nats.CronTopic()).
				Publish(request)
		}
	}
}

func (ws *WsServer) cronMessagesHandler(request []byte) ([]byte, *system.Error) {
	data, err := ws.cronMessagesHandlerType(request)
	if err != nil {
		return nil, &system.Error{
			Error:   err.Error,
			Message: err.Message,
			Code:    err.Code,
			Data:    err.Data,
		}
	} else {
		return data, nil
	}
}

func (ws *WsServer) cronMessagesHandlerType(data []byte) ([]byte, *system.Error) {
	cronMessage := &CronMessage{}

	err := json.Unmarshal(data, cronMessage)
	if err != nil {
		return nil, system.UnmarshalRequestError1201(err, data)
	}

	switch cronMessage.Type {
	case MessageTypeGetUserStatus:
		cronGetUserStatusRequest := &CronGetUserStatusRequest{}
		err := json.Unmarshal(data, cronGetUserStatusRequest)
		if err != nil {
			return nil, system.UnmarshalRequestError1201(err, data)
		}

		cronGetUserStatusResponse := &CronGetAccountStatusResponse{}

		func() {
			ws.actualAccountsMutex.Lock()
			defer ws.actualAccountsMutex.Unlock()
			if _, ok := ws.actualAccounts[cronGetUserStatusRequest.Data.AccountId]; ok {
				cronGetUserStatusResponse.Online = true
			} else {
				cronGetUserStatusResponse.Online = false
			}
		}()

		response, err := json.Marshal(cronGetUserStatusResponse)
		if err != nil {
			return nil, system.MarshalError1011(err, data)
		}

		return response, nil
	case MessageTypeSendOnlineUsers:
		cronSendOnlineUsers := &CronSendOnlineUsers{}
		err := json.Unmarshal(data, cronSendOnlineUsers)
		if err != nil {
			return nil, system.UnmarshalRequestError1201(err, data)
		}

		for _, userId := range cronSendOnlineUsers.Data.Accounts {
			duration := time.Duration(5) * time.Second

			func() {
				ws.actualAccountsMutex.Lock()
				defer ws.actualAccountsMutex.Unlock()
				ws.actualAccounts[userId] = time.Now().Add(duration)
			}()
		}
		return nil, nil
	default:
		return nil, system.UnmarshalRequestTypeError1204(err, data)
	}
}

func (ws *WsServer) userServiceMessageManager() {
	diff := app.Instance.Env.CronStep()
	diffTime := time.Now().Add(-diff)

	rep := r.CreateRepository(app.GetDB())

	for {
		items := rep.RecdUsers(diffTime)

		for _, item := range items {

			item.Id = item.Id
			// TODO: send to bus

			/*
			var title, text string
			chat := app.GetDB().GetRoom(item.RoomId)

			switch item.Role {
			case database.UserTypeOperator:
				title = "Сообщение от МК"
				text = "У вас новое сообщение!"
				break
			case database.UserTypeDoctor:
				title = "Сообщение от врача"
				text = "У вас новое сообщение!"
				break
			}

			if len(text) > 0 {
				message := &ApiUserPushRequest{
					Recipient: RecipientWithOrder{
						AccountId:         item.AccountId,
						ConsultationId: 0,//chat.ReferenceId,
					},
					Message: ApiUserPushRequestMessage{
						Title: title,
						Body:  text,
					},
				}
				ws.hub.inf.Nats.UserPush(message)
			}

			 */
		}

		diffTime = time.Now()
		time.Sleep(diff)
	}
}

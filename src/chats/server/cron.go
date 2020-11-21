package server

import (
	"chats/system"
	"chats/models"
	"chats/sdk"
	"encoding/json"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"time"
)

const (
	MessageTypeGetUserStatus   = "getUserStatus"
	MessageTypeSendOnlineUsers = "sendOnlineUsers"
)

//	Deprecated
func (ws *WsServer) cronManager() {
	if system.Cron() {
		go ws.userServiceMessageManager()
		ws.consumer()
	} else {
		ws.provider()
	}
}

func (ws *WsServer) consumer() {
	errorChan := make(chan *sdk.Error, 2048)
	go ws.hub.app.Sdk.
		Subject(system.CronTopic()).
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
						fmt.Println(userId, "is offline") //	todo
						delete(ws.actualAccounts, userId)
					}
				}
			}()

		case err := <-errorChan:
			system.ErrHandler.SetError(&system.Error{
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

		cronMessage := &models.CronSendOnlineUsers{
			Type: MessageTypeSendOnlineUsers,
			Data: models.CronSendOnlineAccountsData{
				Accounts: actualAccountIds,
			},
		}

		request, err := json.Marshal(cronMessage)
		if err != nil {
			system.MarshalError1011(err, nil)
		} else {
			go ws.hub.app.Sdk.
				Subject(system.CronTopic()).
				Publish(request)
		}
	}
}

func (ws *WsServer) cronMessagesHandler(request []byte) ([]byte, *sdk.Error) {
	data, err := ws.cronMessagesHandlerType(request)
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

func (ws *WsServer) cronMessagesHandlerType(data []byte) ([]byte, *system.Error) {
	cronMessage := &models.CronMessage{}

	err := json.Unmarshal(data, cronMessage)
	if err != nil {
		return nil, system.UnmarshalRequestError1201(err, data)
	}

	switch cronMessage.Type {
	case MessageTypeGetUserStatus:
		cronGetUserStatusRequest := &models.CronGetUserStatusRequest{}
		err := json.Unmarshal(data, cronGetUserStatusRequest)
		if err != nil {
			return nil, system.UnmarshalRequestError1201(err, data)
		}

		cronGetUserStatusResponse := &models.CronGetAccountStatusResponse{}

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
		cronSendOnlineUsers := &models.CronSendOnlineUsers{}
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
	diff := system.CronStep()
	diffTime := time.Now().Add(-diff)

	for {
		items := ws.hub.app.DB.RecdUsers(diffTime)

		for _, item := range items {

			item.Id = item.Id
			// TODO: send to bus

			/*
			var title, text string
			chat := ws.hub.app.DB.GetRoom(item.RoomId)

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
				message := &sdk.ApiUserPushRequest{
					Recipient: sdk.RecipientWithOrder{
						AccountId:         item.AccountId,
						ConsultationId: 0,//chat.ReferenceId,
					},
					Message: sdk.ApiUserPushRequestMessage{
						Title: title,
						Body:  text,
					},
				}
				ws.hub.app.Sdk.UserPush(message)
			}

			 */
		}

		diffTime = time.Now()
		time.Sleep(diff)
	}
}

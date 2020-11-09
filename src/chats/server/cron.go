package server

import (
	"chats/database"
	"chats/models"
	"chats/service"
	"encoding/json"
	"fmt"
	"gitlab.medzdrav.ru/health-service/go-sdk"
	"time"
)

const (
	MessageTypeGetUserStatus   = "getUserStatus"
	MessageTypeSendOnlineUsers = "sendOnlineUsers"
)

//	Deprecated
func (ws *WsServer) cronManager() {
	if service.Cron() {
		go ws.userServiceMessageManager()
		ws.consumer()
	} else {
		ws.provider()
	}
}

func (ws *WsServer) consumer() {
	errorChan := make(chan *sdk.Error, 2048)
	go ws.hub.app.Sdk.
		Subject(service.CronTopic()).
		CronConsumer(ws.cronMessagesHandler, errorChan)

	ticker := time.NewTicker(time.Second * 5).C

	for {
		select {
		case <-ticker:
			now := time.Now()
			go func() {
				ws.actualUsersMutex.Lock()
				defer ws.actualUsersMutex.Unlock()
				for userId, expiredTime := range ws.actualUsers {
					if now.After(expiredTime) {
						fmt.Println(userId, "is offline") //	todo
						delete(ws.actualUsers, userId)
					}
				}
			}()

		case err := <-errorChan:
			service.SetError(&models.SystemError{
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

		actualUserIds := make([]uint, 0, len(ws.hub.users))

		for userId := range ws.hub.users {
			actualUserIds = append(actualUserIds, userId)
		}

		cronMessage := &models.CronSendOnlineUsers{
			Type: MessageTypeSendOnlineUsers,
			Data: models.CronSendOnlineUsersData{
				Users: actualUserIds,
			},
		}

		request, err := json.Marshal(cronMessage)
		if err != nil {
			service.MarshalError1011(err, nil)
		} else {
			go ws.hub.app.Sdk.
				Subject(service.CronTopic()).
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

func (ws *WsServer) cronMessagesHandlerType(data []byte) ([]byte, *models.SystemError) {
	cronMessage := &models.CronMessage{}

	err := json.Unmarshal(data, cronMessage)
	if err != nil {
		return nil, service.UnmarshalRequestError1201(err, data)
	}

	switch cronMessage.Type {
	case MessageTypeGetUserStatus:
		cronGetUserStatusRequest := &models.CronGetUserStatusRequest{}
		err := json.Unmarshal(data, cronGetUserStatusRequest)
		if err != nil {
			return nil, service.UnmarshalRequestError1201(err, data)
		}

		cronGetUserStatusResponse := &models.CronGetUserStatusResponse{}

		func() {
			ws.actualUsersMutex.Lock()
			defer ws.actualUsersMutex.Unlock()
			if _, ok := ws.actualUsers[cronGetUserStatusRequest.Data.UserId]; ok {
				cronGetUserStatusResponse.Online = true
			} else {
				cronGetUserStatusResponse.Online = false
			}
		}()

		response, err := json.Marshal(cronGetUserStatusResponse)
		if err != nil {
			return nil, service.MarshalError1011(err, data)
		}

		return response, nil
	case MessageTypeSendOnlineUsers:
		cronSendOnlineUsers := &models.CronSendOnlineUsers{}
		err := json.Unmarshal(data, cronSendOnlineUsers)
		if err != nil {
			return nil, service.UnmarshalRequestError1201(err, data)
		}

		for _, userId := range cronSendOnlineUsers.Data.Users {
			duration := time.Duration(5) * time.Second

			func() {
				ws.actualUsersMutex.Lock()
				defer ws.actualUsersMutex.Unlock()
				ws.actualUsers[userId] = time.Now().Add(duration)
			}()
		}
		return nil, nil
	default:
		return nil, service.UnmarshalRequestTypeError1204(err, data)
	}
}

func (ws *WsServer) userServiceMessageManager() {
	diff := service.CronStep()
	diffTime := time.Now().Add(-diff)

	for {
		items := ws.hub.app.DB.RecdUsers(diffTime)

		for _, item := range items {
			var title, text string
			consultation := ws.hub.app.DB.Chat(item.ChatId)

			switch item.UserType {
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
						UserId:         item.UserId,
						ConsultationId: consultation.OrderId,
					},
					Message: sdk.ApiUserPushRequestMessage{
						Title: title,
						Body:  text,
					},
				}
				ws.hub.app.Sdk.UserPush(message)
			}
		}

		diffTime = time.Now()
		time.Sleep(diff)
	}
}

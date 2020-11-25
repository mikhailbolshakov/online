package server

import (
	"chats/app"
	r "chats/repository/room"
	"chats/system"
	"encoding/json"
	uuid "github.com/satori/go.uuid"
)

func (ws *WsServer) internalConsumer() {

	dataChan := make(chan []byte, 1024)
	rep := r.CreateRepository(app.GetDB())

	go app.GetNats().
		Subject(app.GetNats().InsideTopic()).
		Consumer(dataChan)

	for {
		data := <-dataChan
		app.L().Debugf("Consumer data: %s", string(data))
		message := &RoomMessage{}
		err := json.Unmarshal(data, message)
		if err != nil {
			app.E().SetError(system.UnmarshalError1010(err, data))
			continue
		}

		if message.RoomId != uuid.Nil {
			app.L().Debugf("Message to room %s", message.RoomId.String())
			if room, ok := ws.hub.rooms[message.RoomId]; ok {
				app.L().Debugf("Room found: %s \n", room.roomId)
				answer, err := json.Marshal(message.Message)
				if err != nil {
					app.E().SetError(system.SysErr(err, system.WsCreateClientResponseCode, nil))
					continue
				} else {
					sessionIds := room.getRoomSessionIds()
					app.L().Debug("- Consumer sessionIds cnt:", len(sessionIds))
					app.L().Debug("- GetRoom sessionIds cnt:", len(room.subscribers))
					for _, uniqueId := range sessionIds {
						if session, ok := ws.hub.sessions[uniqueId]; ok {
							go ws.hub.sendMessage(session, answer)
						}
					}
				}
			}

		} else if message.RoomId == uuid.Nil && message.AccountId != uuid.Nil {
			app.L().Debug("- Consumer message to client")
			answer, err := json.Marshal(message.Message)
			if err != nil {
				app.E().SetError(system.SysErr(err, system.WsCreateClientResponseCode, nil))
				continue
			} else {
				if client, ok := ws.hub.accountSessions[message.AccountId]; ok {
					go ws.hub.sendMessage(client, answer)
					app.L().Debug("NATS: Send WS Message") //	TODO
				} else if message.SendPush {
					/*pushMessage := &ApiUserPushResponse{
						Type: message.Message.Type,
						Data: message.Message.Data,
					}
					go ws.hub.inf.Nats.UserPush(message.AccountId, pushMessage)
					app.L().Debug("NATS: Send Push Message")*///	TODO
				}
			}
		} else if message.RoomId == uuid.Nil && message.AccountId == uuid.Nil {

			switch message.Message.Type {
			case system.SystemMsgTypeUserSubscribe:
				app.L().Debug("System message (user subscribing): ", *message.Message)
				messageData := &WSSystemUserSubscribeRequest{}
				err := json.Unmarshal(data, messageData)
				if err != nil {
					app.E().SetError(system.UnmarshalError1010(err, data))
					continue
				} else {
					if session, ok := ws.hub.accountSessions[messageData.Message.Data.AccountId]; ok {

						app.L().Debugf("Session %s found by account %s", session.sessionId, messageData.Message.Data.AccountId.String())

						// search for subscribers by session account
						subscribers := rep.GetAccountSubscribers(messageData.Message.Data.AccountId)
						app.L().Debugf("Subscribers found: %v", subscribers)

						ws.hub.accountSessions[messageData.Message.Data.AccountId].SetSubscribers(subscribers)

						//	update room
						room := ws.hub.LoadRoomIfNotExists(messageData.Message.Data.RoomId)
						room.AddSession(session.sessionId)
						session.rooms[messageData.Message.Data.RoomId] = room

						app.L().Debug("account " + messageData.Message.Data.AccountId.String() + " added to room")

					} else if room, ok := ws.hub.rooms[messageData.Message.Data.RoomId]; ok {
						subscribers := rep.GetRoomAccountSubscribers(messageData.Message.Data.RoomId)
						room.UpdateSubscribers(subscribers)
						app.L().Debug("room " + messageData.Message.Data.RoomId.String() + " updated subscribers")

					}
				}
				break
			case system.SystemMsgTypeUserUnsubscribe:
				app.L().Debug("WS system message: user unsubscribed", *message.Message)
				messageData := &WSSystemUserUnsubscribeRequest{}
				err := json.Unmarshal(data, messageData)
				if err != nil {
					app.E().SetError(system.UnmarshalError1010(err, data))
					continue
				} else {
					if cli, ok := ws.hub.accountSessions[messageData.Message.Data.AccountId]; ok {
						if room, ok := cli.rooms[messageData.Message.Data.RoomId]; ok {
							room.removeSession(cli)
						}
					}
				}
				break
			}
		}
	}
}

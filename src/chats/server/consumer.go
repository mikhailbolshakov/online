package server

import (
	"chats/app"
	r "chats/repository/room"
	"chats/system"
	"encoding/json"
	uuid "github.com/satori/go.uuid"
)

func (ws *WsServer) messageToRoom(message *RoomMessage) *system.Error {

	defer app.E().CatchPanic("consumer.messageToRoom")

	app.L().Debugf("Message to room %s %s", message.RoomId, message)

	if room, ok := ws.hub.rooms[message.RoomId]; ok {
		app.L().Debugf("Room found: %s \n", room.roomId)
		answer, err := json.Marshal(message.Message)

		if err != nil {
			return system.SysErr(err, system.WsCreateClientResponseCode, nil)
		}

		sessionIds := room.getRoomSessionIds()
		app.L().Debugf("Sessions for room %s count %d", message.RoomId.String(), len(sessionIds))
		app.L().Debugf("Subscribers for room %s count %d", message.RoomId.String(), len(room.subscribers))

		for _, sessionId := range sessionIds {
			if session, ok := ws.hub.sessions[sessionId]; ok {
				go ws.hub.sendMessage(session, answer)
			}
		}

	}

	return nil
}

func (ws *WsServer) messageToAccount(message *RoomMessage) *system.Error {

	defer app.E().CatchPanic("consumer.messageToAccount")

	app.L().Debugf("Message to account %s %s", message.AccountId, message)
	answer, err := json.Marshal(message.Message)
	if err != nil {
		return system.SysErr(err, system.WsCreateClientResponseCode, nil)
	}

	if session, ok := ws.hub.accountSessions[message.AccountId]; ok {
		app.L().Debugf("Session for accountId: %s sessionId: %s", message.AccountId.String(), session.sessionId.String())
		go ws.hub.sendMessage(session, answer)
	} else {
		app.L().Debugf("Session for accountId %s not found", message.AccountId.String())

		/*pushMessage := &ApiUserPushResponse{
			Type: message.Message.Type,
			Data: message.Message.Data,
		}
		go ws.hub.inf.Nats.UserPush(message.AccountId, pushMessage)
		app.L().Debug("NATS: Send Push Message")*///	TODO
	}

	return nil

}

func (ws *WsServer) userSubscribe(data []byte) *system.Error {

	defer app.E().CatchPanic("consumer.userSubscribe")

	message := &WSSystemUserSubscribeRequest{}
	err := json.Unmarshal(data, message)
	if err != nil {
		return system.UnmarshalError1010(err, data)
	}

	app.L().Debugf("User subscribe message %s", *message)

	rep := r.CreateRepository(app.GetDB())

	if session, ok := ws.hub.accountSessions[message.Message.Data.AccountId]; ok {

		rep := r.CreateRepository(app.GetDB())
		app.L().Debugf("Session %s found by account %s", session.sessionId, message.Message.Data.AccountId)

		// search for subscribers by session account
		subscribers := rep.GetAccountSubscribers(message.Message.Data.AccountId)
		app.L().Debugf("Subscribers found: %s", subscribers)

		ws.hub.accountSessions[message.Message.Data.AccountId].SetSubscribers(subscribers)

		//	update room
		room := ws.hub.LoadRoomIfNotExists(message.Message.Data.RoomId)
		room.AddSession(session.sessionId)
		session.rooms[message.Message.Data.RoomId] = room

		app.L().Debug("account " + message.Message.Data.AccountId.String() + " added to room")

	} else if room, ok := ws.hub.rooms[message.Message.Data.RoomId]; ok {
		subscribers := rep.GetRoomAccountSubscribers(message.Message.Data.RoomId)
		room.UpdateSubscribers(subscribers)
		app.L().Debug("room " + message.Message.Data.RoomId.String() + " updated subscribers")

	}

	return nil

}

func (ws *WsServer) userUnSubscribe(data []byte) *system.Error {

	defer app.E().CatchPanic("consumer.userUnSubscribe")

	message := &WSSystemUserUnsubscribeRequest{}
	err := json.Unmarshal(data, message)
	if err != nil {
		return system.UnmarshalError1010(err, data)
	}

	app.L().Debugf("User unsubscribe message %s", message)

	if cli, ok := ws.hub.accountSessions[message.Message.Data.AccountId]; ok {
		if room, ok := cli.rooms[message.Message.Data.RoomId]; ok {
			room.removeSession(cli)
		}
	}
	return nil
}

func (ws *WsServer) internalConsumer() {

	dataChan := make(chan []byte, 1024)

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

			err := ws.messageToRoom(message)
			if err != nil {
				app.E().SetError(err)
			}

		} else if message.RoomId == uuid.Nil && message.AccountId != uuid.Nil {

			err := ws.messageToAccount(message)
			if err != nil {
				app.E().SetError(err)
			}

		} else if message.RoomId == uuid.Nil && message.AccountId == uuid.Nil {

			switch message.Message.Type {

				case system.SystemMsgTypeUserSubscribe:
					err := ws.userSubscribe(data)
					if err != nil {
						app.E().SetError(err)
					}
					break

				case system.SystemMsgTypeUserUnsubscribe:
					err := ws.userUnSubscribe(data)
					if err != nil {
						app.E().SetError(err)
					}
					break
			}
		}
	}
}

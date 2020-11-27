package server

import (
	"chats/app"
	"chats/repository"
	a "chats/repository/account"
	r "chats/repository/room"
	"chats/system"
	"encoding/json"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"time"
)

func (ws *WsServer) sendRoomSubscribeMessage(roomId uuid.UUID, accountId uuid.UUID, role string) {

	//	subscribe websocket hub
	roomMessage := &RoomMessage{
		SendPush:  false,
		AccountId: uuid.Nil,
		RoomId:    uuid.Nil,
		Message: &WSChatResponse{
			Type: system.SystemMsgTypeUserSubscribe,
			Data: &RoomMessageAccountSubscribeRequest{
				AccountId: accountId,
				RoomId:    roomId,
				Role:      role,
			},
		},
	}

	ws.hub.SendMessageToRoom(roomMessage)
}

func (ws *WsServer) sendRoomUnsubscribeMessage(roomId uuid.UUID, accountId uuid.UUID) {

	//	subscribe websocket hub
	roomMessage := &RoomMessage{
		Message: &WSChatResponse{
			Type: system.SystemMsgTypeUserUnsubscribe,
			Data: &RoomMessageAccountUnsubscribeRequest{
				AccountId: accountId,
				RoomId:    roomId,
			},
		},
	}

	ws.hub.SendMessageToRoom(roomMessage)
}

func (ws *WsServer) CreateRoom(request *CreateRoomRequest) (*CreateRoomResponse, *system.Error) {

	defer app.E().CatchPanic("CreateRoom")

	accRep := a.CreateRepository(app.GetDB())
	roomRep := r.CreateRepository(app.GetDB())

	roomModel := &r.Room{
		Id:          system.Uuid(),
		ReferenceId: request.Room.ReferenceId,
		// TODO: generate hash
		Hash:        "",
		Audio:       system.BoolToUint8(request.Room.Audio),
		Video:       system.BoolToUint8(request.Room.Video),
		Chat:        system.BoolToUint8(request.Room.Chat),
		Subscribers: []r.RoomSubscriber{},
	}

	var accountIds []uuid.UUID
	for _, s := range request.Room.Subscribers {

		// TODO: create a method GetAccounts([]accountId)
		account, err := accRep.GetAccount(s.Account.AccountId, s.Account.ExternalId)
		if err != nil {
			return nil, err
		}

		if account.Status != AccountStatusActive {
			return nil, &system.Error{
				// TODO: const && code
				Message: fmt.Sprintf("Account %s isn't active", account.Id.String()),
				Code:    0,
			}
		}

		roomModel.Subscribers = append(roomModel.Subscribers, r.RoomSubscriber{
			Id:            system.Uuid(),
			RoomId:        roomModel.Id,
			AccountId:     account.Id,
			Role:          s.Role,
			SystemAccount: system.BoolToUint8(s.AsSystemAccount),
		})

		accountIds = append(accountIds, account.Id)

	}

	// close all opened rooms for accounts
	err := ws.CloseRoomsByAccounts(accountIds)
	if err != nil {
		return nil, err
	}

	// create a new open room
	roomId, err := roomRep.CreateRoom(roomModel)
	if err != nil {
		return nil, err
	}

	for _, s := range roomModel.Subscribers {
		go ws.sendRoomSubscribeMessage(roomId, s.AccountId, s.Role)
	}

	response := &CreateRoomResponse{
		Result: &RoomResponse{
			Id:   roomModel.Id,
			Hash: roomModel.Hash,
		},
		Errors: []ErrorResponse{},
	}

	return response, nil

}

func (ws *WsServer) CloseRoomsByAccounts(accountIds []uuid.UUID) *system.Error {

	defer app.E().CatchPanic("CloseRoomsByAccounts")

	roomRep := r.CreateRepository(app.GetDB())

	roomIds, err := roomRep.CloseRoomsByAccounts(accountIds)
	if err != nil {
		return err
	}

	ws.hub.roomMutex.Lock()
	defer ws.hub.roomMutex.Unlock()

	for _, roomId := range roomIds {
		delete(ws.hub.rooms, roomId)
	}

	return nil

}

func (ws *WsServer) CloseRoom(request *CloseRoomRequest) (*CloseRoomResponse, *system.Error) {

	defer app.E().CatchPanic("CloseRoom")

	roomRep := r.CreateRepository(app.GetDB())

	response := &CloseRoomResponse{
		Errors: []ErrorResponse{},
	}

	if request.ReferenceId == "" && request.RoomId == uuid.Nil {
		return nil, &system.Error{
			Message: "Invalid request",
			Code:    0,
		}
	}

	rooms, err := roomRep.GetRooms(&r.GetRoomCriteria{ReferenceId: request.ReferenceId, RoomId: request.RoomId})
	if err != nil {
		return nil, err
	}

	for _, room := range rooms {
		err := roomRep.CloseRoom(room.Id)
		if err != nil {
			return nil, err
		}
	}

	ws.hub.roomMutex.Lock()
	defer ws.hub.roomMutex.Unlock()

	for _, room := range rooms {
		delete(ws.hub.rooms, room.Id)
	}

	return response, nil

}

func (ws *WsServer) RoomSubscribe(request *RoomSubscribeRequest) (*RoomSubscribeResponse, *system.Error) {

	defer app.E().CatchPanic("RoomSubscribe")

	roomRep := r.CreateRepository(app.GetDB())
	accRep := a.CreateRepository(app.GetDB())

	// get the room
	var room = &r.Room{}
	var err = &system.Error{}
	if request.RoomId != uuid.Nil {
		// search bu Id if provided
		room, err = roomRep.GetRoom(request.RoomId)
		if err != nil {
			return nil, err
		}
	} else if request.ReferenceId != "" {
		// search by Reference Id if provided
		rs, err := roomRep.GetRooms(&r.GetRoomCriteria{ReferenceId: request.ReferenceId, WithSubscribers: true})
		if err != nil {
			return nil, err
		}
		// set if found
		if len(rs) > 0 {
			room = &rs[0]
		}
	} else {
		return nil, &system.Error{
			Message: "Identifiers of a room isn't provided",
			Code:    0,
		}
	}

	// check if room found
	if room.Id == uuid.Nil {
		return nil, &system.Error{
			Message: fmt.Sprintf("GetRoom %s isn't found", request.RoomId.String()),
			Code:    0,
		}
	}

	// check if room isn't closed
	if room.ClosedAt != nil {
		return nil, &system.Error{
			Message: fmt.Sprintf("GetRoom %s closed", request.RoomId.String()),
			Code:    0,
		}
	}

	// TODO: max number of subscribers isn't exceeded

	// go through requested subscribers
	for _, subscribeRq := range request.Subscribers {

		// get account
		// TODO: GetAccounts
		account, err := accRep.GetAccount(subscribeRq.Account.AccountId, subscribeRq.Account.ExternalId)

		if err != nil {
			return nil, err
		}

		// check if account isn't locked
		if account.Status != AccountStatusActive {
			return nil, &system.Error{
				// TODO: const && code
				Message: fmt.Sprintf("Account %s isn't active", account.Id.String()),
				Code:    0,
			}
		}

		// check if account is already subscribed
		accountFound := false
		for _, s := range room.Subscribers {
			if s.AccountId == account.Id {
				accountFound = true
				break
			}
		}

		if !accountFound {

			// close previous rooms for the account
			err := ws.CloseRoomsByAccounts([]uuid.UUID{account.Id})
			if err != nil {
				return nil, err
			}

			// add subscriber to DB
			subscriber := r.RoomSubscriber{
				Id:            system.Uuid(),
				RoomId:        room.Id,
				AccountId:     account.Id,
				Role:          subscribeRq.Role,
				SystemAccount: system.BoolToUint8(subscribeRq.AsSystemAccount),
			}

			room.Subscribers = append(room.Subscribers, subscriber)
			_, err = roomRep.RoomSubscribeAccount(room, &subscriber)
			if err != nil {
				return nil, err
			}

			go ws.sendRoomSubscribeMessage(room.Id, account.Id, subscribeRq.Role)

		}

	}

	response := &RoomSubscribeResponse{
		Errors: []ErrorResponse{},
	}

	return response, nil
}

func (ws *WsServer) RoomUnsubscribe(request *RoomUnsubscribeRequest) (*RoomUnsubscribeResponse, *system.Error) {

	defer app.E().CatchPanic("RoomUnsubscribe")

	response := &RoomUnsubscribeResponse{
		Errors: []ErrorResponse{},
	}

	app.L().Debugf("Room unsubscribe. Request %v:", *request)

	roomRep := r.CreateRepository(app.GetDB())
	accRep := a.CreateRepository(app.GetDB())

	// get the room
	var rooms []r.Room
	var err = &system.Error{}
	if request.RoomId != uuid.Nil {

		// search bu Id if provided
		room, err := roomRep.GetRoom(request.RoomId)
		if err != nil {
			return nil, err
		}

		if room != nil {
			rooms = append(rooms, *room)
			app.L().Debugf("Room unsubscribe. Room found %s by roomId.", room.Id.String())
		} else {
			return nil, system.SysErrf(nil, system.NoRoomFoundByIdCode, nil, request.RoomId.String())
		}

	} else if request.ReferenceId != "" {
		// search by Reference Id if provided
		rooms, err = roomRep.GetRooms(&r.GetRoomCriteria{ReferenceId: request.ReferenceId, WithSubscribers: true})
		if err != nil {
			return nil, err
		}

		if len(rooms) == 0 {
			return nil, system.SysErrf(nil, system.NoRoomFoundByReferenceCode, nil, request.ReferenceId)
		}

		app.L().Debugf("Room unsubscribe. %d room(s) found by referenceId %s", len(rooms), request.ReferenceId)

	} else {
		return nil, system.SysErr(nil, system.IncorrectRequestCode, nil)
	}

	// get account
	account, err := accRep.GetAccount(request.AccountId.AccountId, request.AccountId.ExternalId)
	if err != nil {
		return nil, err
	}

	// go through all rooms
	for _, room := range rooms {

		// check if room isn't closed
		if room.ClosedAt != nil {
			return nil, system.SysErr(nil, system.RoomAlreadyClosedCode, []byte(request.RoomId.String()))
		}

		accountFound := false

		// go through requested subscribers
		for _, subscriber := range room.Subscribers {

			if subscriber.AccountId == account.Id {

				accountFound = true

				err := roomRep.RoomUnsubscribeAccount(room.Id, account.Id)
				if err != nil {
					return nil, err
				}

				go ws.sendRoomUnsubscribeMessage(room.Id, account.Id)

			}

		}

		if !accountFound {
			return nil, system.SysErrf(nil, system.NotSubscribedAccountCode, nil, room.Id.String(), account.Id.String())
		}
	}

	return response, nil

}

func (ws *WsServer) GetRoomsByCriteria(request *GetRoomsByCriteriaRequest) (*GetRoomsByCriteriaResponse, *system.Error) {

	defer app.E().CatchPanic("GetRoomsByCriteria")

	roomRep := r.CreateRepository(app.GetDB())

	criteriaModel := &r.GetRoomCriteria{
		AccountId:         request.AccountId.AccountId,
		ExternalAccountId: request.AccountId.ExternalId,
		ReferenceId:       request.ReferenceId,
		RoomId:            request.RoomId,
		WithClosed:        request.WithClosed,
		WithSubscribers:   request.WithSubscribers,
	}

	result, err := roomRep.GetRooms(criteriaModel)
	if err != nil {
		return nil, err
	}

	response := &GetRoomsByCriteriaResponse{
		Rooms:  []GetRoomResponse{},
		Errors: []ErrorResponse{},
	}

	for _, item := range result {

		room := GetRoomResponse{
			Id:          item.Id,
			Hash:        item.Hash,
			ReferenceId: item.ReferenceId,
			Chat:        system.Uint8ToBool(item.Chat),
			Video:       system.Uint8ToBool(item.Video),
			Audio:       system.Uint8ToBool(item.Audio),
			ClosedAt:    item.ClosedAt,
			Subscribers: []GetSubscriberResponse{},
		}
		if request.WithSubscribers {
			for _, s := range item.Subscribers {
				room.Subscribers = append(room.Subscribers, GetSubscriberResponse{
					Id:            s.Id,
					AccountId:     s.AccountId,
					Role:          s.Role,
					UnSubscribeAt: s.UnsubscribeAt,
				})
			}
		}
		response.Rooms = append(response.Rooms, room)

	}

	return response, nil

}

func (ws *WsServer) GetMessageHistory(request *GetMessageHistoryRequest) (*GetMessageHistoryResponse, *system.Error) {

	defer app.E().CatchPanic("GetMessageHistory")

	roomRep := r.CreateRepository(app.GetDB())

	if request.Criteria.AccountId.AccountId == uuid.Nil &&
		request.Criteria.AccountId.ExternalId == "" &&
		request.Criteria.RoomId == uuid.Nil &&
		request.Criteria.ReferenceId == "" {
		return nil, &system.Error{
			Message: "Query parameters must be more selective",
			Code:    0,
		}
	}

	criteria := request.Criteria
	criteriaModel := &r.GetMessageHistoryCriteria{
		AccountId:         criteria.AccountId.AccountId,
		AccountExternalId: criteria.AccountId.ExternalId,
		ReferenceId:       criteria.ReferenceId,
		RoomId:            criteria.RoomId,
		Statuses:          criteria.Statuses,
		CreatedBefore:     criteria.CreatedBefore,
		CreatedAfter:      criteria.CreatedAfter,
		WithStatuses:      criteria.WithStatuses,
		SentOnly:          criteria.SentOnly,
		ReceivedOnly:      criteria.ReceivedOnly,
		WithAccounts:      criteria.WithAccounts,
	}

	var pagingRqModel = &repository.PagingRequest{
		SortBy: []repository.SortRequest{},
	}
	if request.PagingRequest != nil {
		pagingRqModel.Index = request.PagingRequest.Index
		pagingRqModel.Size = request.PagingRequest.Size

		for _, s := range request.PagingRequest.SortBy {
			pagingRqModel.SortBy = append(pagingRqModel.SortBy, repository.SortRequest{
				Field:     s.Field,
				Direction: s.Direction,
			})
		}

	}

	if pagingRqModel.Index <= 0 {
		pagingRqModel.Index = 1
	}

	if pagingRqModel.Size <= 1 {
		pagingRqModel.Size = 100
	}

	items, pagingRs, accountsRs, err := roomRep.GetMessageHistory(criteriaModel, pagingRqModel)
	if err != nil {
		return nil, err
	}

	response := &GetMessageHistoryResponse{
		Messages: []MessageHistoryItem{},
		Accounts: []MessageAccount{},
		Paging:   &PagingResponse{},
		Errors:   []ErrorResponse{},
	}

	for _, item := range items {
		message := MessageHistoryItem{
			Id:                 item.Id,
			ClientMessageId:    item.ClientMessageId,
			ReferenceId:        item.ReferenceId,
			RoomId:             item.RoomId,
			Type:               item.Type,
			Message:            item.Message,
			FileId:             item.FileId,
			Params:             item.Params,
			SenderAccountId:    item.SenderAccountId,
			RecipientAccountId: item.RecipientAccountId,
			Statuses:           []MessageStatus{},
		}

		for _, s := range item.Statuses {
			message.Statuses = append(message.Statuses, MessageStatus{
				AccountId:  s.AccountId,
				Status:     s.Status,
				StatusDate: s.StatusDate,
			})
		}

		response.Messages = append(response.Messages, message)
	}

	for _, a := range accountsRs {
		response.Accounts = append(response.Accounts, MessageAccount{
			Id:         a.Id,
			Type:       a.Type,
			Status:     a.Status,
			Account:    a.Account,
			ExternalId: a.ExternalId,
			FirstName:  a.FirstName,
			MiddleName: a.MiddleName,
			LastName:   a.LastName,
			Email:      a.Email,
			Phone:      a.Phone,
			AvatarUrl:  a.AvatarUrl,
		})
	}

	response.Paging.Index = pagingRs.Index
	response.Paging.Total = pagingRs.Total

	return response, nil
}

func (ws *WsServer) SendChatMessages(request *SendChatMessagesRequest) (*SendChatMessageResponse, *system.Error) {

	defer app.E().CatchPanic("SendChatMessages")

	response := &SendChatMessageResponse{}

	app.L().Debugf("Chat message sending. Request: %s", request)
	rqJson, _ := json.Marshal(request)

	roomRepository := r.CreateRepository(app.GetDB())
	accountRepository := a.CreateRepository(app.GetDB())

	loc, err := app.Instance.GetLocation()
	if err != nil {
		return nil, system.SysErr(err, system.LoadLocationErrorCode, nil)
	}

	var roomId uuid.UUID
	recipients := make(map[uuid.UUID]Account)
	var subscribers []r.RoomSubscriber
	var sysErr = &system.Error{}

	for _, item := range request.Data.Messages {
		if len(item.Text) > maxMessageSize {
			return nil, system.SysErr(err, system.MessageTooLongErrorCode, rqJson)
		}
		if item.RoomId == uuid.Nil {
			return nil, system.SysErr(err, system.MysqlChatIdIncorrectCode, rqJson)
		}

		if roomId == uuid.Nil {
			roomId = item.RoomId
			subscribers, sysErr = roomRepository.GetRoomSubscribers(roomId)
			if sysErr != nil {
				return nil, sysErr
			}

			if len(subscribers) == 0 {
				return nil, system.SysErr(err, system.MysqlChatSubscribeEmptyCode, rqJson)
			}
		}

		var senderSubscriberId = uuid.Nil
		var senderAccountId = uuid.Nil
		var senderSubscriberType string
		var opponents []r.ChatOpponent

		for _, s := range subscribers {

			// add to list recipients all the accounts (including sender) except system account (bot)
			if _, ok := recipients[s.AccountId]; !ok && !system.Uint8ToBool(s.SystemAccount) {

				account, err := accountRepository.GetAccount(s.AccountId, "")
				if err != nil {
					return nil, err
				}
				recipients[s.AccountId] = *ConvertAccountFromModel(account)
			}

			if s.AccountId == request.SenderAccountId {
				senderAccountId = s.AccountId
				senderSubscriberId = s.Id
				senderSubscriberType = s.Role
			} else {
				opponents = append(opponents, r.ChatOpponent{
					SubscriberId:  s.Id,
					AccountId:     s.AccountId,
					SystemAccount: system.Uint8ToBool(s.SystemAccount),
				})
			}
		}

		if senderSubscriberId == uuid.Nil {
			return nil, system.SysErr(err, system.MysqlChatAccessDeniedCode, rqJson)
		}

		if item.RecipientAccountId != uuid.Nil {
			if _, ok := recipients[item.RecipientAccountId]; !ok {
				return nil, system.SysErr(err, system.PrivateChatRecipientNotFoundAmongSubscribersCode, rqJson)
			}
		}

		paramsJson, err := json.Marshal(item.Params)
		if err != nil {
			return nil, system.SysErr(err, system.UnmarshallingErrorCode, nil)
		}

		dbMessage := &r.ChatMessage{
			Id:              system.Uuid(),
			ClientMessageId: item.ClientMessageId,
			RoomId:          roomId,
			AccountId:       senderAccountId,
			Type:            item.Type,
			SubscribeId:     senderSubscriberId,
			Message:         item.Text,
			Params:          string(paramsJson),
		}
		if item.RecipientAccountId != uuid.Nil {
			dbMessage.RecipientAccountId = &item.RecipientAccountId
		}

		sysErr = roomRepository.CreateMessage(dbMessage, opponents)
		if sysErr != nil {
			return nil, sysErr
		}

		messageResponse := &WSChatMessagesDataMessageResponse{
			Id:                 dbMessage.Id,
			ClientMessageId:    item.ClientMessageId,
			InsertDate:         dbMessage.CreatedAt.In(loc).Format(time.RFC3339),
			ChatId:             roomId,
			AccountId:          request.SenderAccountId,
			Sender:             senderSubscriberType,
			Status:             r.MessageStatusRecd,
			Type:               item.Type,
			Text:               item.Text,
			RecipientAccountId: item.RecipientAccountId,
			Params:             item.Params,
		}

		//if len(dbMessage.FileId) > 0 {
		//	//file := &FileModel{Id: dbMessage.FileId}
		//	//sdkErr := h.inf.Nats.File(file, roomId, c.account.Id)
		//	//if sdkErr != nil {
		//	//	app.E().SetError(&system.Error{
		//	//		Error:   sdkErr.Error,
		//	//		Message: sdkErr.Message,
		//	//		Code:    sdkErr.Code,
		//	//		Data:    sdkErr.Data,
		//	//	})
		//	//	return
		//	//}
		//	messageResponseData := &WSChatMessagesDataMessageFileResponse{
		//		WSChatMessagesDataMessageResponse: *messageResponse,
		//		//File:                              nil,
		//	}
		//	messages = append(messages, messageResponseData)
		//} else {
		//	messages = append(messages, messageResponse)
		//}

		if messageResponse.RecipientAccountId != uuid.Nil {

			rsMsg := &RoomMessage{
				RoomId:    uuid.Nil,
				AccountId: messageResponse.RecipientAccountId,
				Message: &WSChatResponse{
					Type: EventMessage,
					Data: WSChatMessagesDataResponse{
						Messages: []interface{}{messageResponse},
						Accounts: []Account{recipients[messageResponse.RecipientAccountId]},
					},
				},
			}

			// send to internal NATS topic for balancing
			ws.hub.SendMessageToRoom(rsMsg)

		} else {

			var recipientAccounts []Account
			for _, item := range recipients {
				recipientAccounts = append(recipientAccounts, item)
			}

			rsMsg := &RoomMessage{
				RoomId: roomId,
				Message: &WSChatResponse{
					Type: EventMessage,
					Data: WSChatMessagesDataResponse{
						Messages: []interface{}{messageResponse},
						Accounts: recipientAccounts,
					},
				},
			}

			// send to internal NATS topic for balancing
			ws.hub.SendMessageToRoom(rsMsg)

		}

	}

	return response, nil
}

func (ws *WsServer) resendRecdMessagesToSession(session *Session, roomId uuid.UUID) {

	defer app.E().CatchPanic("resendRecdMessagesToSession")

	loc, e := app.Instance.GetLocation()
	if e != nil {
		system.SysErr(e, system.LoadLocationErrorCode, nil)
	}

	roomRepository := r.CreateRepository(app.GetDB())

	messages, err := roomRepository.GetAccountRecdMessages(session.account.Id, roomId)
	if err != nil {
		app.E().SetError(err)
	}

	if len(messages) > 0 {
		app.L().Debugf("Messages to resend found: %s", len(messages))
		for _, m := range messages {

			jsonParams := make(map[string]string)
			if m.Params != "" {
				err := json.Unmarshal([]byte(m.Params), &jsonParams)
				if err != nil {
					app.E().SetError(system.E(err))
				}
			}

			var recipientAccountId uuid.UUID
			if m.RecipientAccountId != nil {
				recipientAccountId = *m.RecipientAccountId
			}

			msg := &WSChatResponse{
				Type: EventMessage,
				Data: WSChatMessagesDataResponse{
					Messages: []interface{}{
						&WSChatMessagesDataMessageResponse{
							Id:                 m.Id,
							ClientMessageId:    m.ClientMessageId,
							InsertDate:         m.CreatedAt.In(loc).Format(time.RFC3339),
							ChatId:             roomId,
							AccountId:          m.AccountId,
							Status:             r.MessageStatusRecd,
							Type:               m.Type,
							Text:               m.Message,
							RecipientAccountId: recipientAccountId,
							Params:             jsonParams,
						}},
				},
			}

			msgMar, e := json.Marshal(msg)
			if e != nil {
				app.E().SetError(err)
			}

			app.L().Debugf("Resending message to session. accountId: %s, msg: %s", session.account.Id, string(msgMar))
			go ws.hub.sendMessage(session, msgMar)
		}
	}

}

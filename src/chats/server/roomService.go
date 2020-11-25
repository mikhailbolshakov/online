package server

import (
	"chats/models"
	"chats/sdk"
	"chats/system"
	"fmt"
	uuid "github.com/satori/go.uuid"
)

func (ws *WsServer) sendRoomSubscribeMessage(roomId uuid.UUID, accountId uuid.UUID, role string) {

	//	subscribe websocket hub
	roomMessage := &RoomMessage{
		SendPush:  false,
		AccountId: uuid.Nil,
		RoomId:    uuid.Nil,
		Message: &models.WSChatResponse{
			Type: system.SystemMsgTypeUserSubscribe,
			Data: &sdk.RoomMessageAccountSubscribeRequest{
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
		Message: &models.WSChatResponse{
			Type: system.SystemMsgTypeUserUnsubscribe,
			Data: &sdk.RoomMessageAccountUnsubscribeRequest{
				AccountId: accountId,
				RoomId:    roomId,
			},
		},
	}

	ws.hub.SendMessageToRoom(roomMessage)
}

func (ws *WsServer) CreateRoom(request *sdk.CreateRoomRequest) (*sdk.CreateRoomResponse, *system.Error) {

	roomModel := &models.Room{
		Id:          system.Uuid(),
		ReferenceId: request.Room.ReferenceId,
		// TODO: generate hash
		Hash:        "",
		Audio:       system.BoolToUint8(request.Room.Audio),
		Video:       system.BoolToUint8(request.Room.Video),
		Chat:        system.BoolToUint8(request.Room.Chat),
		Subscribers: []models.RoomSubscriber{},
	}

	var accountIds []uuid.UUID
	for _, s := range request.Room.Subscribers {

		// TODO: create a method GetAccounts([]accountId)
		account, err := ws.hub.app.DB.GetAccount(s.Account.AccountId, s.Account.ExternalId)
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

		roomModel.Subscribers = append(roomModel.Subscribers, models.RoomSubscriber{
			Id:            system.Uuid(),
			RoomId:        roomModel.Id,
			AccountId:     account.Id,
			Role:          s.Role,
			SystemAccount: system.BoolToUint8(false),
		})

		accountIds = append(accountIds, account.Id)

	}

	// close all opened rooms for accounts
	err := ws.CloseRoomsByAccounts(accountIds)
	if err != nil {
		return nil, err
	}

	// create a new open room
	roomId, err := ws.hub.app.DB.CreateRoom(roomModel)
	if err != nil {
		return nil, err
	}

	for _, s := range roomModel.Subscribers {
		go ws.sendRoomSubscribeMessage(roomId, s.AccountId, s.Role)
	}

	response := &sdk.CreateRoomResponse{
		Result: &sdk.RoomResponse{
			Id:   roomModel.Id,
			Hash: roomModel.Hash,
		},
		Errors: []sdk.ErrorResponse{},
	}

	return response, nil

}

func (ws *WsServer) CloseRoomsByAccounts(accountIds []uuid.UUID) *system.Error {

	roomIds, err := ws.hub.app.DB.CloseRoomsByAccounts(accountIds)
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

func (ws *WsServer) CloseRoom(request *sdk.CloseRoomRequest) (*sdk.CloseRoomResponse, *system.Error) {

	response := &sdk.CloseRoomResponse{
		Errors: []sdk.ErrorResponse{},
	}

	if request.ReferenceId == "" && request.RoomId == uuid.Nil {
		return nil, &system.Error{
			Message: "Invalid request",
			Code:    0,
		}
	}

	rooms, err := ws.hub.app.DB.GetRooms(&models.GetRoomCriteria{ReferenceId: request.ReferenceId, RoomId: request.RoomId})
	if err != nil {
		return nil, err
	}

	for _, room := range rooms {
		err := ws.hub.app.DB.CloseRoom(room.Id)
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

func (ws *WsServer) RoomSubscribe(request *sdk.RoomSubscribeRequest) (*sdk.RoomSubscribeResponse, *system.Error) {

	// get the room
	var room = &models.Room{}
	var err = &system.Error{}
	if request.RoomId != uuid.Nil {
		// search bu Id if provided
		room, err = ws.hub.app.DB.GetRoom(request.RoomId)
		if err != nil {
			return nil, err
		}
	} else if request.ReferenceId != "" {
		// search by Reference Id if provided
		rs, err := ws.hub.app.DB.GetRooms(&models.GetRoomCriteria{ReferenceId: request.ReferenceId, WithSubscribers: true})
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
		account, err := ws.hub.app.DB.GetAccount(subscribeRq.Account.AccountId, subscribeRq.Account.ExternalId)

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
			// add subscriber to DB
			subscriber := models.RoomSubscriber{
				Id:        system.Uuid(),
				RoomId:    room.Id,
				AccountId: account.Id,
				Role:      subscribeRq.Role,
				// TODO: add attr SystemAccount to account
				SystemAccount: system.BoolToUint8(false),
			}

			room.Subscribers = append(room.Subscribers, subscriber)
			_, err = ws.hub.app.DB.RoomSubscribeAccount(room, &subscriber)
			if err != nil {
				return nil, err
			}

			go ws.sendRoomSubscribeMessage(room.Id, account.Id, subscribeRq.Role)

		}

	}

	response := &sdk.RoomSubscribeResponse{
		Errors: []sdk.ErrorResponse{},
	}

	return response, nil
}

func (ws *WsServer) RoomUnsubscribe(request *sdk.RoomUnsubscribeRequest) (*sdk.RoomUnsubscribeResponse, *system.Error) {

	// get the room
	var room = &models.Room{}
	var err = &system.Error{}
	if request.RoomId != uuid.Nil {
		// search bu Id if provided
		room, err = ws.hub.app.DB.GetRoom(request.RoomId)
		if err != nil {
			return nil, err
		}
	} else if request.ReferenceId != "" {
		// search by Reference Id if provided
		rs, err := ws.hub.app.DB.GetRooms(&models.GetRoomCriteria{ReferenceId: request.ReferenceId, WithSubscribers: true})
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

	account, err := ws.hub.app.DB.GetAccount(request.AccountId.AccountId, request.AccountId.ExternalId)
	if err != nil {
		return nil, err
	}

	accountFound := false

	// go through requested subscribers
	for _, subscriber := range room.Subscribers {

		if subscriber.AccountId == account.Id {

			accountFound = true

			err := ws.hub.app.DB.RoomUnsubscribeAccount(room.Id, account.Id)
			if err != nil {
				return nil, err
			}

			go ws.sendRoomUnsubscribeMessage(room.Id, account.Id)

		}

	}

	if !accountFound {
		return nil, &system.Error{
			Message: fmt.Sprintf("No account %s found for room %s to unsubscribe", room.Id.String(), account.Id.String()),
			Code:    0,
		}
	}

	response := &sdk.RoomUnsubscribeResponse{
		Errors: []sdk.ErrorResponse{},
	}

	return response, nil

}

func (ws *WsServer) GetRoomsByCriteria(request *sdk.GetRoomsByCriteriaRequest) (*sdk.GetRoomsByCriteriaResponse, *system.Error) {

	criteriaModel := &models.GetRoomCriteria{
		AccountId:         request.AccountId.AccountId,
		ExternalAccountId: request.AccountId.ExternalId,
		ReferenceId:       request.ReferenceId,
		RoomId:            request.RoomId,
		WithClosed:        request.WithClosed,
		WithSubscribers:   request.WithSubscribers,
	}

	result, err := ws.hub.app.DB.GetRooms(criteriaModel)
	if err != nil {
		return nil, err
	}

	response := &sdk.GetRoomsByCriteriaResponse{
		Rooms:  []sdk.GetRoomResponse{},
		Errors: []sdk.ErrorResponse{},
	}

	for _, item := range result {

		room := sdk.GetRoomResponse{
			Id:          item.Id,
			Hash:        item.Hash,
			ReferenceId: item.ReferenceId,
			Chat:        system.Uint8ToBool(item.Chat),
			Video:       system.Uint8ToBool(item.Video),
			Audio:       system.Uint8ToBool(item.Audio),
			ClosedAt:    item.ClosedAt,
			Subscribers: []sdk.GetSubscriberResponse{},
		}
		if request.WithSubscribers {
			for _, s := range item.Subscribers {
				room.Subscribers = append(room.Subscribers, sdk.GetSubscriberResponse{
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

func (ws *WsServer) GetMessageHistory(request *sdk.GetMessageHistoryRequest) (*sdk.GetMessageHistoryResponse, *system.Error) {

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
	criteriaModel := &models.GetMessageHistoryCriteria{
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

	var pagingRqModel = &models.PagingRequest{
		SortBy: []models.SortRequest{},
	}
	if request.PagingRequest != nil {
		pagingRqModel.Index = request.PagingRequest.Index
		pagingRqModel.Size = request.PagingRequest.Size

		for _, s := range request.PagingRequest.SortBy {
			pagingRqModel.SortBy = append(pagingRqModel.SortBy, models.SortRequest{
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

	items, pagingRs, accountsRs, err := ws.hub.app.DB.GetMessageHistory(criteriaModel, pagingRqModel)
	if err != nil {
		return nil, err
	}

	response := &sdk.GetMessageHistoryResponse{
		Messages: []sdk.MessageHistoryItem{},
		Accounts: []sdk.MessageAccount{},
		Paging:   &sdk.PagingResponse{},
		Errors:   []sdk.ErrorResponse{},
	}

	for _, item := range items {
		message := sdk.MessageHistoryItem{
			Id:               item.Id,
			ClientMessageId:  item.ClientMessageId,
			ReferenceId:      item.ReferenceId,
			RoomId:           item.RoomId,
			Type:             item.Type,
			Message:          item.Message,
			FileId:           item.FileId,
			Params:           item.Params,
			SenderAccountId:  item.SenderAccountId,
			SenderExternalId: item.SenderExternalId,
			Statuses:         []sdk.MessageStatus{},
		}

		for _, s := range item.Statuses {
			message.Statuses = append(message.Statuses, sdk.MessageStatus{
				AccountId:  s.AccountId,
				Status:     s.Status,
				StatusDate: s.StatusDate,
			})
		}

		response.Messages = append(response.Messages, message)
	}

	for _, a := range accountsRs {
		response.Accounts = append(response.Accounts, sdk.MessageAccount{
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

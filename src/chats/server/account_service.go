package server

import (
	"chats/database"
	"chats/infrastructure"
	"chats/models"
	"chats/sdk"
	"chats/sentry"
	"encoding/json"
	"github.com/satori/go.uuid"
)

const (
	AccountStatusActive = "active"
	AccountStatusLocked = "locked"
)

const (
	AccountTypeBot = "bot"
	AccountTypeAnonymousUser = "anonymous_user"
	AccountTypeUser = "account"
)

const (
	OnlineStatusOffline = "offline"
	OnlineStatusOnline = "online"
	OnlineStatusBusy = "busy"
	OnlineStatusAway = "away"
)

func (ws *WsServer) createAccount(params []byte) ([]byte, *sentry.SystemError) {

	request := &sdk.CreateAccountRequest{}
	err := json.Unmarshal(params, request)
	if err != nil {
		return nil, infrastructure.UnmarshalRequestError1201(err, params)
	}

	id, _ := uuid.NewV4()
	model := &models.Account{
		Id:         id,
		Type:       request.Body.Type,
		Status:		AccountStatusActive,
		Account:    request.Body.Account,
		ExternalId: request.Body.ExternalId,
		FirstName:  request.Body.FirstName,
		MiddleName: request.Body.MiddleName,
		LastName:   request.Body.LastName,
		Email:      request.Body.Email,
		Phone:      request.Body.Phone,
		AvatarUrl:  request.Body.AvatarUrl,
	}

	accountId, err := ws.hub.app.DB.CreateAccount(model)
	if err != nil {
		return nil, &sentry.SystemError{
			Error:   err,
			Message: database.MysqlChatCreateError,
			Code:    database.MysqlChatCreateErrorCode,
			Data:    params,
		}
	}

	id, _ = uuid.NewV4()
	onlineStatusModel := &models.OnlineStatus{
		Id:        id,
		AccountId: accountId,
		Status:    OnlineStatusOffline,
	}

	_, err = ws.hub.app.DB.CreateOnlineStatus(onlineStatusModel)
	if err != nil {
		return nil, &sentry.SystemError{
			Error:   err,
			Message: database.MysqlChatCreateError,
			Code:    database.MysqlChatCreateErrorCode,
			Data:    params,
		}
	}

	response := &sdk.CreateAccountResponse{AccountId: accountId}

	result, err := json.Marshal(response)
	if err != nil {
		return nil, infrastructure.MarshalError1011(err, params)
	}

	return result, nil
}

func (ws *WsServer) setOnlineStatus(accountId uuid.UUID, onlineStatus string) () {



}

func (ws *WsServer) updateAccount(params []byte) ([]byte, *sentry.SystemError) {
	return nil, nil
}
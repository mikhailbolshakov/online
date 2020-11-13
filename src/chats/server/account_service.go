package server

import (
	"chats/converter"
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
	AccountTypeBot           = "bot"
	AccountTypeAnonymousUser = "anonymous_user"
	AccountTypeUser          = "user"
)

const (
	OnlineStatusOffline = "offline"
	OnlineStatusOnline  = "online"
	OnlineStatusBusy    = "busy"
	OnlineStatusAway    = "away"
)

func validateCreateAccount(account *sdk.Account) (bool, *sentry.SystemError) {

	if account.Type == "" || (account.Type != AccountTypeAnonymousUser && account.Type != AccountTypeUser) {
		return false, &sentry.SystemError {
			Message: WsCreateAccountInvalidTypeErrorMessage,
			Code:    WsCreateAccountInvalidTypeErrorCode,
		}
	}

	if account.Account == "" {
		return false, &sentry.SystemError {
			Message: WsCreateAccountEmptyAccountErrorMessage,
			Code:    WsCreateAccountEmptyAccountErrorCode,
		}
	}

	return true, nil
}

func (ws *WsServer) createAccount(params []byte) ([]byte, *sentry.SystemError) {

	request := &sdk.CreateAccountRequest{}
	err := json.Unmarshal(params, request)
	if err != nil {
		return nil, infrastructure.UnmarshalRequestError1201(err, params)
	}

	if ok, err := validateCreateAccount(&request.Body); !ok {
		return nil, err
	}

	id, _ := uuid.NewV4()
	model := &models.Account{
		Id:         id,
		Type:       request.Body.Type,
		Status:     AccountStatusActive,
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

func (ws *WsServer) setOnlineStatus(params []byte) ([]byte, *sentry.SystemError) {

	request := &sdk.UpdateAccountOnlineStatusRequest{}
	err := json.Unmarshal(params, request)
	if err != nil {
		return nil, infrastructure.UnmarshalRequestError1201(err, params)
	}

	// get account
	account, sentryErr := ws.hub.app.DB.GetAccount(request.Body.Account.AccountId, request.Body.Account.ExternalId)
	if sentryErr != nil {
		return nil, sentryErr
	}

	// get online status
	onlineStatusObj, sentryErr := ws.hub.app.DB.GetOnlineStatus(account.Id)
	if sentryErr != nil {
		return nil, sentryErr
	}

	onlineStatusObj.Status = request.Body.Status

	// update online status
	sentryErr = ws.hub.app.DB.UpdateOnlineStatus(onlineStatusObj)
	if sentryErr != nil {
		return nil, sentryErr
	}

	response := &sdk.BoolResponseData{Result: true}
	result, err := json.Marshal(response)
	if err != nil {
		return nil, infrastructure.MarshalError1011(err, params)
	}

	return result, nil

}

func (ws *WsServer) updateAccount(params []byte) ([]byte, *sentry.SystemError) {
	return nil, nil
}



func (ws *WsServer) getAccountById(params []byte) ([]byte, *sentry.SystemError) {

	request := &sdk.GetAccountByIdRequest{}
	err := json.Unmarshal(params, request)
	if err != nil {
		return nil, infrastructure.UnmarshalRequestError1201(err, params)
	}

	accountModel, sentryErr := ws.hub.app.DB.GetAccount(request.Body.Account.AccountId, request.Body.Account.ExternalId)
	if sentryErr != nil {
		return nil, sentryErr
	}

	response := &sdk.GetAccountResponse{Data: *converter.ConvertAccountFromModel(accountModel)}
	result, err := json.Marshal(response)

	if err != nil {
		return nil, infrastructure.MarshalError1011(err, params)
	}

	return result, nil
}

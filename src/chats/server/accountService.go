package server

import (
	"chats/models"
	"chats/sdk"
	"chats/system"
	"fmt"
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

func validateCreateAccount(account *sdk.CreateAccountRequest) (bool, *system.Error) {

	if account.Type == "" || (account.Type != AccountTypeAnonymousUser && account.Type != AccountTypeUser) {
		return false, &system.Error{
			Message: WsCreateAccountInvalidTypeErrorMessage,
			Code:    WsCreateAccountInvalidTypeErrorCode,
		}
	}

	if account.Account == "" {
		return false, &system.Error{
			Message: WsCreateAccountEmptyAccountErrorMessage,
			Code:    WsCreateAccountEmptyAccountErrorCode,
		}
	}

	return true, nil
}

func (ws *WsServer) createAccount(request *sdk.CreateAccountRequest) (*sdk.CreateAccountResponse, *system.Error) {

	response := &sdk.CreateAccountResponse{
		Errors: []sdk.ErrorResponse{},
	}

	if ok, err := validateCreateAccount(request); !ok {
		response.Errors = []sdk.ErrorResponse{
			{
				Message: err.Message,
				Code:    err.Code,
				Stack:   err.Stack,
			},
		}
		return response, nil
	}

	model := &models.Account{
		Id:         system.Uuid(),
		Type:       request.Type,
		Status:     AccountStatusActive,
		Account:    request.Account,
		ExternalId: request.ExternalId,
		FirstName:  request.FirstName,
		MiddleName: request.MiddleName,
		LastName:   request.LastName,
		Email:      request.Email,
		Phone:      request.Phone,
		AvatarUrl:  request.AvatarUrl,
	}

	accountId, err := ws.hub.app.DB.CreateAccount(model)
	if err != nil {
		return nil, err
	}

	onlineStatusModel := &models.OnlineStatus{
		Id:        system.Uuid(),
		AccountId: accountId,
		Status:    OnlineStatusOffline,
	}

	_, err = ws.hub.app.DB.CreateOnlineStatus(onlineStatusModel)
	if err != nil {
		return nil, err
	}

	response.AccountId = accountId

	return response, nil
}

func (ws *WsServer) setOnlineStatus(request *sdk.SetAccountOnlineStatusRequest) (*sdk.SetAccountOnlineStatusResponse, *system.Error) {

	ok := false
	for _, s := range []string{OnlineStatusOffline, OnlineStatusAway, OnlineStatusBusy, OnlineStatusOnline} {
		if request.Status == s {
			ok = true
			break
		}
	}
	if !ok {
		return nil, &system.Error{
			Message: fmt.Sprintf("Status %s is invalid", request.Status),
			Code:    0,
		}
	}

	// get account
	account, err := ws.hub.app.DB.GetAccount(request.Account.AccountId, request.Account.ExternalId)
	if err != nil {
		return nil, err
	}

	if account == nil {
		return nil, &system.Error{
			Message: fmt.Sprintf("Account not found (id: %s, extId: %s)", request.Account.AccountId.String(), request.Account.ExternalId),
			Code:    0,
		}
	}

	// update online status
	err = ws.hub.app.DB.UpdateOnlineStatus(account.Id, request.Status)
	if err != nil {
		return nil, err
	}

	response := &sdk.SetAccountOnlineStatusResponse{Errors: []sdk.ErrorResponse{}}

	return response, nil

}

func (ws *WsServer) getOnlineStatus(request *sdk.GetAccountOnlineStatusRequest) (*sdk.GetAccountOnlineStatusResponse, *system.Error) {

	// get account
	account, err := ws.hub.app.DB.GetAccount(request.Account.AccountId, request.Account.ExternalId)
	if err != nil {
		return nil, err
	}

	if account == nil {
		return nil, &system.Error{
			Message: fmt.Sprintf("Account not found (id: %s, extId: %s)", request.Account.AccountId.String(), request.Account.ExternalId),
			Code:    0,
		}
	}

	status, err := ws.hub.app.DB.GetOnlineStatus(account.Id)
	if err != nil {
		return nil, err
	}

	response := &sdk.GetAccountOnlineStatusResponse{
		Errors: []sdk.ErrorResponse{},
		Status: status,
	}

	return response, nil

}

func (ws *WsServer) updateAccount(request *sdk.UpdateAccountRequest) (*sdk.UpdateAccountResponse, *system.Error) {

	account, err := ws.hub.app.DB.GetAccount(request.AccountId.AccountId, request.AccountId.ExternalId)
	if err != nil {
		return nil, err
	}

	if account == nil {
		return nil, &system.Error{
			Message: fmt.Sprintf("Account not found (id: %s, extId: %s)", request.AccountId.AccountId.String(), request.AccountId.ExternalId),
			Code:    0,
		}
	}

	account.LastName = request.LastName
	account.MiddleName = request.MiddleName
	account.FirstName = request.FirstName
	account.Phone = request.Phone
	account.Email = request.Email
	account.AvatarUrl = request.AvatarUrl

	if err := ws.hub.app.DB.UpdateAccount(account); err != nil {
		return nil, err
	}

	response := &sdk.UpdateAccountResponse{Errors: []sdk.ErrorResponse{}}
	return response, nil

}

func (ws *WsServer) getAccountsByCriteria(criteria *sdk.GetAccountsByCriteriaRequest) (*sdk.GetAccountsByCriteriaResponse, *system.Error) {

	items, err := ws.hub.app.DB.GetAccountsByCriteria(&models.GetAccountsCriteria{
		AccountId:  criteria.AccountId.AccountId,
		ExternalId: criteria.AccountId.ExternalId,
		Email:      criteria.Email,
		Phone:      criteria.Phone,
	})
	if err != nil {
		return nil, err
	}

	response := &sdk.GetAccountsByCriteriaResponse{Accounts: []sdk.Account{}, Errors: []sdk.ErrorResponse{}}
	for _, i := range items {
		response.Accounts = append(response.Accounts, sdk.Account{
			Id:         i.Id,
			Account:    i.Account,
			Type:       i.Type,
			ExternalId: i.ExternalId,
			FirstName:  i.FirstName,
			MiddleName: i.MiddleName,
			LastName:   i.LastName,
			Email:      i.Email,
			Phone:      i.Phone,
			AvatarUrl:  i.AvatarUrl,
		})
	}

	return response, nil

}

func (ws *WsServer) getAccountById(request *sdk.AccountIdRequest) (*sdk.Account, *system.Error) {

	accountModel, err := ws.hub.app.DB.GetAccount(request.AccountId, request.ExternalId)
	if err != nil {
		return nil, err
	}

	return ConvertAccountFromModel(accountModel), nil
}

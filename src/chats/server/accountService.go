package server

import (
	"chats/app"
	a "chats/repository/account"
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

func validateCreateAccount(account *CreateAccountRequest) (bool, *system.Error) {

	if account.Type == "" || (account.Type != AccountTypeAnonymousUser && account.Type != AccountTypeUser) {
		return false, system.SysErr(nil, system.WsCreateAccountInvalidTypeErrorCode, nil)
	}

	if account.Account == "" {
		return false, system.SysErr(nil, system.WsCreateAccountEmptyAccountErrorCode, nil)
	}

	return true, nil
}

func (ws *WsServer) createAccount(request *CreateAccountRequest) (*CreateAccountResponse, *system.Error) {

	response := &CreateAccountResponse{
		Errors: []ErrorResponse{},
	}

	if ok, err := validateCreateAccount(request); !ok {
		response.Errors = []ErrorResponse{
			{
				Message: err.Message,
				Code:    err.Code,
				Stack:   err.Stack,
			},
		}
		return response, nil
	}

	model := &a.Account{
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

	rep := a.CreateRepository(app.Instance.Inf.DB)
	accountId, err := rep.CreateAccount(model)
	if err != nil {
		return nil, err
	}

	onlineStatusModel := &a.OnlineStatus{
		Id:        system.Uuid(),
		AccountId: accountId,
		Status:    OnlineStatusOffline,
	}

	_, err = rep.CreateOnlineStatus(onlineStatusModel)
	if err != nil {
		return nil, err
	}

	response.AccountId = accountId

	return response, nil
}

func (ws *WsServer) setOnlineStatus(request *SetAccountOnlineStatusRequest) (*SetAccountOnlineStatusResponse, *system.Error) {

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

	rep := a.CreateRepository(app.GetDB())

	// get account
	account, err := rep.GetAccount(request.Account.AccountId, request.Account.ExternalId)
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
	err = rep.UpdateOnlineStatus(account.Id, request.Status)
	if err != nil {
		return nil, err
	}

	response := &SetAccountOnlineStatusResponse{Errors: []ErrorResponse{}}

	return response, nil

}

func (ws *WsServer) getOnlineStatus(request *GetAccountOnlineStatusRequest) (*GetAccountOnlineStatusResponse, *system.Error) {

	rep := a.CreateRepository(app.GetDB())

	// get account
	account, err := rep.GetAccount(request.Account.AccountId, request.Account.ExternalId)
	if err != nil {
		return nil, err
	}

	if account == nil {
		return nil, &system.Error{
			Message: fmt.Sprintf("Account not found (id: %s, extId: %s)", request.Account.AccountId.String(), request.Account.ExternalId),
			Code:    0,
		}
	}

	status, err := rep.GetOnlineStatus(account.Id)
	if err != nil {
		return nil, err
	}

	response := &GetAccountOnlineStatusResponse{
		Errors: []ErrorResponse{},
		Status: status,
	}

	return response, nil

}

func (ws *WsServer) updateAccount(request *UpdateAccountRequest) (*UpdateAccountResponse, *system.Error) {

	rep := a.CreateRepository(app.GetDB())

	account, err := rep.GetAccount(request.AccountId.AccountId, request.AccountId.ExternalId)
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

	if err := rep.UpdateAccount(account); err != nil {
		return nil, err
	}

	response := &UpdateAccountResponse{Errors: []ErrorResponse{}}
	return response, nil

}

func (ws *WsServer) getAccountsByCriteria(criteria *GetAccountsByCriteriaRequest) (*GetAccountsByCriteriaResponse, *system.Error) {

	rep := a.CreateRepository(app.GetDB())

	items, err := rep.GetAccountsByCriteria(&a.GetAccountsCriteria{
		AccountId:  criteria.AccountId.AccountId,
		ExternalId: criteria.AccountId.ExternalId,
		Email:      criteria.Email,
		Phone:      criteria.Phone,
	})
	if err != nil {
		return nil, err
	}

	response := &GetAccountsByCriteriaResponse{Accounts: []Account{}, Errors: []ErrorResponse{}}
	for _, i := range items {
		response.Accounts = append(response.Accounts, Account{
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

func (ws *WsServer) getAccountById(request *AccountIdRequest) (*Account, *system.Error) {

	rep := a.CreateRepository(app.GetDB())

	accountModel, err := rep.GetAccount(request.AccountId, request.ExternalId)
	if err != nil {
		return nil, err
	}

	return ConvertAccountFromModel(accountModel), nil
}

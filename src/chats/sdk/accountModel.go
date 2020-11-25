package sdk

import uuid "github.com/satori/go.uuid"

type AccountIdRequest struct {
	AccountId  uuid.UUID `json:"accountId"`
	ExternalId string    `json:"externalId"`
}

type CreateAccountRequest struct {
	Account    string `json:"account"`
	Type       string `json:"type"`
	ExternalId string `json:"externalId"`
	FirstName  string `json:"firstName"`
	MiddleName string `json:"middleName"`
	LastName   string `json:"lastName"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	AvatarUrl  string `json:"avatarUrl"`
}

type UpdateAccountRequest struct {
	AccountId  AccountIdRequest `json:"accountId"`
	FirstName  string           `json:"firstName"`
	MiddleName string           `json:"middleName"`
	LastName   string           `json:"lastName"`
	Email      string           `json:"email"`
	Phone      string           `json:"phone"`
	AvatarUrl  string           `json:"avatarUrl"`
}

type UpdateAccountResponse struct {
	Errors []ErrorResponse `json:"errors"`
}

type CreateAccountResponse struct {
	AccountId uuid.UUID       `json:"accountId"`
	Errors    []ErrorResponse `json:"errors"`
}

type GetAccountsByCriteriaRequest struct {
	AccountId AccountIdRequest `json:"accountId"`
	Email     string           `json:"email"`
	Phone     string           `json:"phone"`
}

type GetAccountsByCriteriaResponse struct {
	Accounts []Account `json:"accounts"`
	Errors []ErrorResponse `json:"errors"`
}

type Account struct {
	Id         uuid.UUID `json:"id"`
	Account    string    `json:"account"`
	Type       string    `json:"type"`
	ExternalId string    `json:"externalId"`
	FirstName  string    `json:"firstName"`
	MiddleName string    `json:"middleName"`
	LastName   string    `json:"lastName"`
	Email      string    `json:"email"`
	Phone      string    `json:"phone"`
	AvatarUrl  string    `json:"avatarUrl"`
}

type SetAccountOnlineStatusRequest struct {
	Account *AccountIdRequest `json:"account"`
	Status  string            `json:"status"`
}

type SetAccountOnlineStatusResponse struct {
	Errors []ErrorResponse `json:"errors"`
}

type GetAccountOnlineStatusRequest struct {
	Account *AccountIdRequest `json:"account"`
}

type GetAccountOnlineStatusResponse struct {
	Errors []ErrorResponse `json:"errors"`
	Status  string            `json:"status"`
}

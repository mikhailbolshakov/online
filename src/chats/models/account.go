package models

import (
	"github.com/satori/go.uuid"
)

type Account struct {
	Id         uuid.UUID
	Type       string `gorm:"column:account_type"`
	Status     string `gorm:"column:status"`
	Account    string `gorm:"column:account"`
	ExternalId string `gorm:"column:external_id"`
	FirstName  string `gorm:"column:first_name"`
	MiddleName string `gorm:"column:middle_name"`
	LastName   string `gorm:"column:last_name"`
	Email      string `gorm:"column:email"`
	Phone      string `gorm:"column:phone"`
	AvatarUrl  string `gorm:"column:avatar_url"`
	BaseModel
}

type OnlineStatus struct {
	Id        uuid.UUID
	AccountId uuid.UUID `gorm:"column:account_id"`
	Status    string    `gorm:"column:status"`
	BaseModel
}

type GetAccountsCriteria struct {
	AccountId uuid.UUID
	ExternalId string
	Email string
	Phone string
}

package tests

import (
	"chats/sdk"
	uuid "github.com/satori/go.uuid"
	"testing"
	"chats/tests/test_helper"
)

func TestCreateAndGetById_Success(t *testing.T) {

	sdkService, err := test_helper.InitSdk()
	if err != nil {
		t.Error(err.Error(), sdkService)
	}

	externalId, _ := uuid.NewV4()
	accountId, err := test_helper.CreateAccount(sdkService, externalId.String())

	if err != nil {
		t.Fatal(err.Error())
	}

	if accountId == uuid.Nil {
		t.Fatal("Account creation failed. AccountId is empty")
	}

	accountRs, err := test_helper.GetAccountById(sdkService, accountId)
	if err != nil {
		t.Fatal(err.Error())
	}

	if accountRs.Data.Id == uuid.Nil {
		t.Fatal("Account not found by Id")
	}

}

func TestCreateEmptyType_Error(t *testing.T) {

	sdkService, err := test_helper.InitSdk()
	if err != nil {
		t.Error(err.Error(), sdkService)
	}

	accountId, err := test_helper.CreateAccountExt(sdkService, &sdk.Account{
		Account:    "test",
		Type:       "",
		ExternalId: "111",
	})

	if accountId != uuid.Nil {
		t.Fatal("Test failed")
	}

}

func TestCreateAndGetByExternalId_Success(t *testing.T) {

	sdkService, err := test_helper.InitSdk()
	if err != nil {
		t.Error(err.Error(), sdkService)
	}

	externalId, _ := uuid.NewV4()
	accountId, err := test_helper.CreateAccount(sdkService, externalId.String())

	if err != nil {
		t.Fatal(err.Error())
	}

	if accountId == uuid.Nil {
		t.Fatal("Account creation failed. AccountId is empty")
	}

	accountRs, err := test_helper.GetAccountByExternalId(sdkService, externalId.String())
	if err != nil {
		t.Fatal(err.Error())
	}

	if accountRs.Data.Id == uuid.Nil {
		t.Fatal("Account not found by external Id")
	}
}

func TestGetByEmpty_Error(t *testing.T) {

	sdkService, err := test_helper.InitSdk()
	if err != nil {
		t.Error(err.Error(), sdkService)
	}

	emptyAccount, err := test_helper.GetAccountById(sdkService, uuid.Nil)
	if emptyAccount.Data.Id != uuid.Nil {
		t.Fatal("Test failed")
	}

}

func TestOnlineStatus_Success(t *testing.T) {

	sdkService, err := test_helper.InitSdk()
	if err != nil {
		t.Error(err.Error(), sdkService)
	}

	accountId, err := test_helper.CreateAccount(sdkService, "")

	if err != nil {
		t.Fatal(err.Error())
	}

	if accountId == uuid.Nil {
		t.Fatal("Account creation failed. AccountId is empty")
	}

	rs, err := test_helper.SetAccountOnlineStatus(sdkService, accountId, "online")
	if err != nil {
		t.Fatal(err.Error())
	}

	if !rs.Result {
		t.Fatal(err.Error())
	}

}

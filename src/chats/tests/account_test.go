package tests

import (
	pb "chats/proto"
	"chats/tests/helper"
	"testing"
)

func TestCreateDefault_Success(t *testing.T) {

	conn, err := helper.GrpcConnection()
	if err != nil {
		t.Fatal(err.Error())
	}
	defer conn.Close()

	accountId, externalId, err := helper.CreateDefaultAccount(conn)
	if err != nil {
		t.Fatal(err.Error())
	}

	items, err := helper.GetAccountsByCriteria(conn, &pb.GetAccountsByCriteriaRequest{
		AccountId: &pb.AccountIdRequest{
			AccountId:  pb.FromUUID(accountId),
			ExternalId: "",
		},
		Email:     "",
		Phone:     "",
	})
	if err != nil {
		t.Fatal(err.Error())
	}
	if items == nil || len(items) == 0 {
		t.Fatal("No items found")
	}

	items, err = helper.GetAccountsByCriteria(conn, &pb.GetAccountsByCriteriaRequest{
		AccountId: &pb.AccountIdRequest{
			ExternalId: externalId,
		},
		Email:     "",
		Phone:     "",
	})
	if err != nil {
		t.Fatal(err.Error())
	}
	if items == nil || len(items) == 0 {
		t.Fatal("No items found")
	}

}

func TestUpdate_Success(t *testing.T) {

	conn, err := helper.GrpcConnection()
	if err != nil {
		t.Fatal(err.Error())
	}
	defer conn.Close()

	accountId, externalId, err := helper.CreateDefaultAccount(conn)
	if err != nil {
		t.Fatal(err.Error())
	}

	err = helper.UpdateAccount(conn, &pb.UpdateAccountRequest{
		AccountId:  &pb.AccountIdRequest{
			AccountId:  pb.FromUUID(accountId),
			ExternalId: "",
		},
		FirstName:  "Петр",
		MiddleName: "Петрович",
		LastName:   "Петров",
		Email:      "petr@gmail.com",
		Phone:      "+7888999000",
		AvatarUrl:  "http://...",
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	err = helper.UpdateAccount(conn, &pb.UpdateAccountRequest{
		AccountId:  &pb.AccountIdRequest{
			AccountId:  nil,
			ExternalId: externalId,
		},
		FirstName:  "Глеб",
		MiddleName: "Глебович",
		LastName:   "Глебов",
		Email:      "petr@gmail.com",
		Phone:      "+7888999000",
		AvatarUrl:  "http://...",
	})
	if err != nil {
		t.Fatal(err.Error())
	}

}

func TestOnlineStatus_Success(t *testing.T) {
	conn, err := helper.GrpcConnection()
	if err != nil {
		t.Fatal(err.Error())
	}
	defer conn.Close()

	accountId, _, err := helper.CreateDefaultAccount(conn)
	if err != nil {
		t.Fatal(err.Error())
	}

	_, err = helper.GetAccountOnlineStatus(conn, accountId)
	if err != nil {
		t.Fatal(err.Error())
	}

	err = helper.SetAccountOnlineStatus(conn, accountId, "online")
	if err != nil {
		t.Fatal(err.Error())
	}

	_, err = helper.GetAccountOnlineStatus(conn, accountId)
	if err != nil {
		t.Fatal(err.Error())
	}

	err = helper.SetAccountOnlineStatus(conn, accountId, "away")
	if err != nil {
		t.Fatal(err.Error())
	}

	_, err = helper.GetAccountOnlineStatus(conn, accountId)
	if err != nil {
		t.Fatal(err.Error())
	}

}

//func TestCreateEmptyType_Error(t *testing.T) {
//
//	sdkService, err := helper.InitSdk()
//	if err != nil {
//		t.Error(err.Error(), sdkService)
//	}
//
//	accountId, err := helper.CreateAccountExt(sdkService, &sdk.Account{
//		Account:    "test",
//		Type:       "",
//		ExternalId: "111",
//	})
//
//	if accountId != uuid.Nil {
//		t.Fatal("Test failed")
//	}
//
//}
//
//func TestCreateAndGetByExternalId_Success(t *testing.T) {
//
//	sdkService, err := helper.InitSdk()
//	if err != nil {
//		t.Error(err.Error(), sdkService)
//	}
//
//	externalId, _ := uuid.NewV4()
//	accountId, err := helper.CreateDefaultAccount(sdkService, externalId.String())
//
//	if err != nil {
//		t.Fatal(err.Error())
//	}
//
//	if accountId == uuid.Nil {
//		t.Fatal("Account creation failed. AccountId is empty")
//	}
//
//	accountRs, err := helper.GetAccountByExternalId(sdkService, externalId.String())
//	if err != nil {
//		t.Fatal(err.Error())
//	}
//
//	if accountRs.Data.Id == uuid.Nil {
//		t.Fatal("Account not found by external Id")
//	}
//}
//
//func TestGetByEmpty_Error(t *testing.T) {
//
//	sdkService, err := helper.InitSdk()
//	if err != nil {
//		t.Error(err.Error(), sdkService)
//	}
//
//	emptyAccount, err := helper.GetAccountById(sdkService, uuid.Nil)
//	if emptyAccount.Data.Id != uuid.Nil {
//		t.Fatal("Test failed")
//	}
//
//}
//
//func TestOnlineStatus_Success(t *testing.T) {
//
//	sdkService, err := helper.InitSdk()
//	if err != nil {
//		t.Error(err.Error(), sdkService)
//	}
//
//	accountId, err := helper.CreateDefaultAccount(sdkService, "")
//
//	if err != nil {
//		t.Fatal(err.Error())
//	}
//
//	if accountId == uuid.Nil {
//		t.Fatal("Account creation failed. AccountId is empty")
//	}
//
//	rs, err := helper.SetAccountOnlineStatus(sdkService, accountId, "online")
//	if err != nil {
//		t.Fatal(err.Error())
//	}
//
//	if !rs.Result {
//		t.Fatal(err.Error())
//	}
//
//}

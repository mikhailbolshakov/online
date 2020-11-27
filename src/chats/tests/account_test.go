package tests

import (
	pb "chats/proto"
	"chats/tests/helper"
	"testing"
	"time"
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

func TestOnlineStatusWhenWSConnectingAndDisconnecting_Success(t *testing.T) {

	conn, err := helper.GrpcConnection()
	if err != nil {
		t.Fatal(err.Error())
	}
	defer conn.Close()

	accountId, _, err := helper.CreateDefaultAccount(conn)
	if err != nil {
		t.Fatal(err)
	}

	_, err = helper.GetAccountOnlineStatus(conn, accountId)
	if err != nil {
		t.Fatal(err)
	}

	wsFirst, _, err := helper.AccountWebSocket(accountId)

	time.Sleep(time.Second)

	_, err = helper.GetAccountOnlineStatus(conn, accountId)
	if err != nil {
		t.Fatal(err)
	}

	wsFirst.Close()

	time.Sleep(time.Second)

	_, err = helper.GetAccountOnlineStatus(conn, accountId)
	if err != nil {
		t.Fatal(err)
	}

	wsFirst, _, err = helper.AccountWebSocket(accountId)

	time.Sleep(time.Second)

	_, err = helper.GetAccountOnlineStatus(conn, accountId)
	if err != nil {
		t.Fatal(err)
	}

}


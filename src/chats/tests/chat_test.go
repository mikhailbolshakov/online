package tests

import (
	"chats/models"
	pb "chats/proto"
	"chats/server"
	"chats/system"
	"chats/tests/helper"
	"context"
	uuid "github.com/satori/go.uuid"
	"log"
	"testing"
	"time"
)

func TestNewRoomMessageExchange_Success(t *testing.T) {

	conn, err := helper.GrpcConnection()
	if err != nil {
		t.Fatal(err.Error())
	}
	defer conn.Close()

	accountIdFirst, _, err := helper.CreateDefaultAccount(conn)
	accountIdSecond, externalIdSecond, err := helper.CreateDefaultAccount(conn)

	wsFirst, msgChanFirst, err := helper.AccountWebSocket(accountIdFirst)
	wsSecond, msgChanSecond, err := helper.AccountWebSocket(accountIdSecond)

	done := make(chan interface{})
	receivedChan := make(chan interface{})

	//defer wsFirst.Close()
	//defer wsSecond.Close()
	defer close(msgChanFirst)
	defer close(msgChanSecond)

	roomService := pb.NewRoomClient(conn)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	referenceId := system.Uuid().String()
	r, err := roomService.Create(ctx, &pb.CreateRoomRequest{
		ReferenceId: referenceId,
		Chat:        true,
		Video:       false,
		Audio:       false,
		Subscribers: []*pb.SubscriberRequest{
			{
				Account: &pb.AccountIdRequest{
					AccountId: pb.FromUUID(accountIdFirst),
				},
				Role: "client",
			},
			{
				Account: &pb.AccountIdRequest{
					ExternalId: externalIdSecond,
				},
				Role: "operator",
			},
		},
	})
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	roomId := r.Result.Id.ToUUID()

	go helper.ReadMessages(wsFirst, msgChanFirst, roomId, accountIdFirst, receivedChan, done, true)
	go helper.ReadMessages(wsSecond, msgChanSecond, roomId, accountIdSecond, receivedChan, done, true)

	err = helper.SendMessage(wsFirst, accountIdFirst, server.EventMessage, &models.WSChatMessageDataRequest{
		RoomId: roomId,
		Type:   "message",
		Text:   "привет второй",
		Params: map[string]string{"param1": "value1", "param2": "value2"},
	})
	if err != nil {
		t.Fatal("Failed")
	}

	err = helper.SendMessage(wsSecond, accountIdSecond, server.EventMessage, &models.WSChatMessageDataRequest{
		RoomId: roomId,
		Type:   "message",
		Text:   "привет первый",
		Params: map[string]string{"param1": "value1", "param2": "value2"},
	})
	if err != nil {
		t.Fatal("Failed")
	}

	var received = 0
	for {
		select {
			case <-receivedChan:
				received++
				if received == 2 {
					done <- true
					return
				}
			case <-time.After(10 * time.Second):
				done <- true
				t.Fatal("Test failed. Timeout")
		}
	}

}

func TestGetChatInfo_Success(t *testing.T) {

	sdkService, err := helper.InitSdk()
	if err != nil {
		t.Fatal(err.Error(), sdkService)
	}

	chatInfo, err := helper.GetChatInfo(sdkService,
							uuid.FromStringOrNil("c7dc4c3c-7d88-4a4e-ae90-2b1775158405"),
							uuid.FromStringOrNil("093d596a-4299-4ad1-9f77-4677adb3ce96"),
							"")

	if err != nil {
		t.Fatal(err)
	}

	for _, item := range chatInfo.Data {
		log.Println(item)
	}

}

func TestGetChatsByAccount_Success(t *testing.T) {

	sdkService, err := helper.InitSdk()
	if err != nil {
		t.Fatal(err.Error(), sdkService)
	}

	chats, err := helper.GetChatsByAccount(sdkService,
		uuid.FromStringOrNil("093d596a-4299-4ad1-9f77-4677adb3ce96"),
		"")

	if err != nil {
		t.Fatal(err)
	}

	for _, item := range chats.Data {
		log.Println(item)
	}

}

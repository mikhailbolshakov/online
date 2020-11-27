package tests

import (
	pb "chats/proto"
	"chats/server"
	"chats/system"
	"chats/tests/helper"
	"context"
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

	err = helper.SendMessage(wsFirst, accountIdFirst, server.EventMessage, &server.WSChatMessageDataRequest{
		RoomId: roomId,
		Type:   "message",
		Text:   "привет второй",
		Params: map[string]string{"param1": "value1", "param2": "value2"},
	})
	if err != nil {
		t.Fatal("Failed")
	}

	err = helper.SendMessage(wsSecond, accountIdSecond, server.EventMessage, &server.WSChatMessageDataRequest{
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

func TestSystemAccountSendPrivateMessage_Success(t *testing.T) {

	conn, err := helper.GrpcConnection()
	if err != nil {
		t.Fatal(err.Error())
	}
	defer conn.Close()

	accountIdFirst, _, err := helper.CreateDefaultAccount(conn)
	botAccountId, _, err := helper.CreateBotAccount(conn)

	wsFirst, msgChanFirst, err := helper.AccountWebSocket(accountIdFirst)

	done := make(chan interface{})
	receivedChan := make(chan interface{})

	//defer wsFirst.Close()
	//defer wsSecond.Close()
	defer close(msgChanFirst)

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
					AccountId: pb.FromUUID(botAccountId),
				},
				Role: "bot",
				AsSystemAccount: true,
			},
		},
	})
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	roomId := r.Result.Id.ToUUID()

	go helper.ReadMessages(wsFirst, msgChanFirst, roomId, accountIdFirst, receivedChan, done, true)

	_, err = roomService.SendChatMessages(ctx, &pb.SendChatMessagesRequest{
		SenderAccountId: pb.FromUUID(botAccountId),
		Type:            server.EventMessage,
		Data:            &pb.SendChatMessagesDataRequest{Messages: []*pb.SendChatMessageDataRequest {
			{
				ClientMessageId:    "",
				RoomId:             pb.FromUUID(roomId),
				Type:               "message",
				Text:               "добро пожаловать",
				Params:             map[string]string{"param1": "value1", "param2": "value2"},
				RecipientAccountId: pb.FromUUID(accountIdFirst),
			},
		} },
	})
	if err != nil {
		t.Fatal("Failed")
	}

	var received = 0
	for {
		select {
		case <-receivedChan:
			received++
			if received == 1 {
				done <- true
				return
			}
		case <-time.After(10 * time.Second):
			done <- true
			t.Fatal("Test failed. Timeout")
		}
	}

}

func TestTwoSubscribersBotSendsPrivateMessage_Success(t *testing.T) {

	conn, err := helper.GrpcConnection()
	if err != nil {
		t.Fatal(err.Error())
	}
	defer conn.Close()

	accountIdFirst, _, err := helper.CreateDefaultAccount(conn)
	accountIdSecond, _, err := helper.CreateDefaultAccount(conn)
	botAccountId, _, err := helper.CreateBotAccount(conn)

	wsFirst, msgChanFirst, err := helper.AccountWebSocket(accountIdFirst)
	wsSecond, msgChanSecond, err := helper.AccountWebSocket(accountIdSecond)

	done := make(chan interface{})
	receivedChan := make(chan interface{})

	//defer wsFirst.Close()
	//defer wsSecond.Close()
	defer close(msgChanFirst)

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
					AccountId: pb.FromUUID(accountIdSecond),
				},
				Role: "operator",
			},
			{
				Account: &pb.AccountIdRequest{
					AccountId: pb.FromUUID(botAccountId),
				},
				Role: "bot",
				AsSystemAccount: true,
			},
		},
	})
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	roomId := r.Result.Id.ToUUID()
	log.Println("RoomId:", roomId.String())

	go helper.ReadMessages(wsFirst, msgChanFirst, roomId, accountIdFirst, receivedChan, done, true)
	go helper.ReadMessages(wsSecond, msgChanSecond, roomId, accountIdSecond, receivedChan, done, true)

	_, err = roomService.SendChatMessages(ctx, &pb.SendChatMessagesRequest{
		SenderAccountId: pb.FromUUID(botAccountId),
		Type:            server.EventMessage,
		Data:            &pb.SendChatMessagesDataRequest{Messages: []*pb.SendChatMessageDataRequest {
			{
				ClientMessageId:    "",
				RoomId:             pb.FromUUID(roomId),
				Type:               "message",
				Text:               "добро пожаловать Первый",
				Params:             map[string]string{"param1": "value1", "param2": "value2"},
				RecipientAccountId: pb.FromUUID(accountIdFirst),
			},
			{
				ClientMessageId:    "",
				RoomId:             pb.FromUUID(roomId),
				Type:               "message",
				Text:               "добро пожаловать Второй",
				Params:             map[string]string{"param1": "value1", "param2": "value2"},
				RecipientAccountId: pb.FromUUID(accountIdSecond),
			},
		} },
	})
	if err != nil {
		t.Fatal("Failed")
	}

	err = helper.SendMessage(wsFirst, accountIdFirst, server.EventMessage, &server.WSChatMessageDataRequest{
		RoomId: roomId,
		Type:   "message",
		Text:   "привет второй",
		Params: map[string]string{"param1": "value1", "param2": "value2"},
	})
	if err != nil {
		t.Fatal("Failed")
	}

	err = helper.SendMessage(wsSecond, accountIdSecond, server.EventMessage, &server.WSChatMessageDataRequest{
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
			if received == 4 {
				done <- true
				return
			}
		case <-time.After(10 * time.Second):
			done <- true
			t.Fatal("Test failed. Timeout")
		}
	}

}


func TestMessagingAfterReconnection_Success(t *testing.T) {

	conn, err := helper.GrpcConnection()
	if err != nil {
		t.Fatal(err.Error())
	}
	defer conn.Close()

	accountIdFirst, _, err := helper.CreateDefaultAccount(conn)
	accountIdSecond, _, err := helper.CreateDefaultAccount(conn)

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
					AccountId: pb.FromUUID(accountIdSecond),
				},
				Role: "operator",
			},
		},
	})
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	roomId := r.Result.Id.ToUUID()
	log.Println("RoomId:", roomId.String())

	//go helper.ReadMessages(wsFirst, msgChanFirst, roomId, accountIdFirst, receivedChan, done, true)
	//go helper.ReadMessages(wsSecond, msgChanSecond, roomId, accountIdSecond, receivedChan, done, true)

	err = helper.SendMessage(wsFirst, accountIdFirst, server.EventMessage, &server.WSChatMessageDataRequest{
		RoomId: roomId,
		Type:   "message",
		Text:   "привет второй",
		Params: map[string]string{"param1": "value1", "param2": "value2"},
	})
	if err != nil {
		t.Fatal("Failed")
	}

	err = helper.SendMessage(wsSecond, accountIdSecond, server.EventMessage, &server.WSChatMessageDataRequest{
		RoomId: roomId,
		Type:   "message",
		Text:   "привет первый",
		Params: map[string]string{"param1": "value1", "param2": "value2"},
	})
	if err != nil {
		t.Fatal("Failed")
	}

	time.Sleep(time.Second)

	//close connection
	//close(msgChanFirst)
	//close(msgChanSecond)
	wsFirst.Close()
	wsSecond.Close()

	wsFirst, msgChanFirst, err = helper.AccountWebSocket(accountIdFirst)
	wsSecond, msgChanSecond, err = helper.AccountWebSocket(accountIdSecond)

	go helper.ReadMessages(wsFirst, msgChanFirst, roomId, accountIdFirst, receivedChan, done, true)
	go helper.ReadMessages(wsSecond, msgChanSecond, roomId, accountIdSecond, receivedChan, done, true)

	err = helper.SendMessage(wsFirst, accountIdFirst, server.EventMessage, &server.WSChatMessageDataRequest{
		RoomId: roomId,
		Type:   "message",
		Text:   "привет второй, я жив",
		Params: map[string]string{"param1": "value1", "param2": "value2"},
	})
	if err != nil {
		t.Fatal("Failed")
	}

	err = helper.SendMessage(wsSecond, accountIdSecond, server.EventMessage, &server.WSChatMessageDataRequest{
		RoomId: roomId,
		Type:   "message",
		Text:   "привет первый, я жив",
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
			if received == 4 {
				done <- true
				return
			}
		case <-time.After(10 * time.Second):
			done <- true
			t.Fatal("Test failed. Timeout")
		}
	}

}

func TestResendMessages_Success(t *testing.T) {

	conn, err := helper.GrpcConnection()
	if err != nil {
		t.Fatal(err.Error())
	}
	defer conn.Close()

	accountIdFirst, _, err := helper.CreateDefaultAccount(conn)
	accountIdSecond, _, err := helper.CreateDefaultAccount(conn)

	wsFirst, msgChanFirst, err := helper.AccountWebSocket(accountIdFirst)

	done := make(chan interface{})
	receivedChan := make(chan interface{})

	//defer wsFirst.Close()
	//defer wsSecond.Close()
	defer close(msgChanFirst)

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
					AccountId: pb.FromUUID(accountIdSecond),
				},
				Role: "operator",
			},
		},
	})
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	roomId := r.Result.Id.ToUUID()
	log.Println("RoomId:", roomId.String())

	err = helper.SendMessage(wsFirst, accountIdFirst, server.EventMessage, &server.WSChatMessageDataRequest{
		RoomId: roomId,
		Type:   "message",
		Text:   "привет второй",
		Params: map[string]string{"param1": "value1", "param2": "value2"},
	})
	if err != nil {
		t.Fatal("Failed")
	}

	//close connection
	//close(msgChanFirst)
	//close(msgChanSecond)
	wsFirst.Close()

	time.Sleep(time.Second)

	wsSecond, msgChanSecond, err := helper.AccountWebSocket(accountIdSecond)
	defer close(msgChanSecond)

	go helper.ReadMessages(wsSecond, msgChanSecond, roomId, accountIdSecond, receivedChan, done, true)

	var received = 0
	for {
		select {
		case <-receivedChan:
			received++
			if received == 1 {
				done <- true
				return
			}
		case <-time.After(10 * time.Second):
			done <- true
			t.Fatal("Test failed. Timeout")
		}
	}

}


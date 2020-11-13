package tests

import (
	"chats/models"
	"chats/server"
	"chats/tests/test_helper"
	"encoding/json"
	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
	"log"
	"testing"
	"time"
)

func TestNewChatAndSubscribeByAccountId_Success(t *testing.T) {

	var chatId = uuid.Nil

	sdkService, err := test_helper.InitSdk()
	if err != nil {
		t.Error(err.Error(), sdkService)
	}

	accountId_First, err := test_helper.CreateAccount(sdkService, "111")
	accountId_Second, err := test_helper.CreateAccount(sdkService, "222")

	conn_first, msgChan_First, err := test_helper.AccountWebSocket(accountId_First)
	conn_second, msgChan_Second, err := test_helper.AccountWebSocket(accountId_Second)

	done := make(chan struct{})

	defer close(msgChan_First)
	defer close(msgChan_Second)
	defer conn_first.Close()
	defer conn_second.Close()

	read := func(conn *websocket.Conn, readChan chan []byte, accountId uuid.UUID) {

		for {
			select {
			case msg := <-readChan:
				log.Println("[First]: " + string(msg))
				message := &test_helper.WSChatResponse{}
				_ = json.Unmarshal(msg, message)
				for _, m := range message.Data.Messages {
					var direct string
					if m.AccountId == accountId {
						direct = "send"
					} else {
						direct = "received"

						if chatId != uuid.Nil {
							err := test_helper.SendReadStatus(conn, chatId, m.Id)
							if err != nil {
								t.Fatal(err)
							} else {
								log.Printf("%s read sent \n", m.Id)
							}
						}

					}

					log.Printf("%s [%s]: %s (id = %s)\n", direct, accountId, m.Text, m.Id)
				}

			case <-done:
				return
			}
		}
	}

	go read(conn_first, msgChan_First, accountId_First)
	go read(conn_second, msgChan_Second, accountId_Second)

	newChatRs, err := test_helper.NewChatAndSubscribe(sdkService, accountId_First, "", "ref1", "client")
	if err != nil || newChatRs.Data.ChatId == uuid.Nil {
		t.Fatal("New chat creation failed")
	}
	chatId = newChatRs.Data.ChatId

	subscribeRs, err := test_helper.ChatSubscribe(sdkService, chatId, accountId_Second, "ref1", "client")
	if err != nil || !subscribeRs.Data.Result {
		t.Fatal("Subscription failed")
	}

	err = test_helper.SendMessage(conn_first, accountId_First, server.EventMessage, &models.WSChatMessageDataRequest{
		ChatId: newChatRs.Data.ChatId,
		Type:   "message",
		Text:   "привет второй",
		Params: map[string]string{"param1": "value1", "param2": "value2"},
	})

	err = test_helper.SendMessage(conn_second, accountId_Second, server.EventMessage, &models.WSChatMessageDataRequest{
		ChatId: newChatRs.Data.ChatId,
		Type:   "message",
		Text:   "привет первый",
		Params: map[string]string{"param1": "value1", "param2": "value2"},
	})

	if err != nil {
		t.Fatal("Failed")
	}

	for {
		time.Sleep(time.Second)
	}

}

func TestGetChatInfo_Success(t *testing.T) {

	sdkService, err := test_helper.InitSdk()
	if err != nil {
		t.Fatal(err.Error(), sdkService)
	}

	chatInfo, err := test_helper.GetChatInfo(sdkService,
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

	sdkService, err := test_helper.InitSdk()
	if err != nil {
		t.Fatal(err.Error(), sdkService)
	}

	chats, err := test_helper.GetChatsByAccount(sdkService,
		uuid.FromStringOrNil("093d596a-4299-4ad1-9f77-4677adb3ce96"),
		"")

	if err != nil {
		t.Fatal(err)
	}

	for _, item := range chats.Data {
		log.Println(item)
	}

}

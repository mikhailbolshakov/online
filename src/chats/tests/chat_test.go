package tests

//func TestNewChatAndSubscribeByAccountId_Success(t *testing.T) {
//
//	var chatId = uuid.Nil
//
//	sdkService, err := helper.InitSdk()
//	if err != nil {
//		t.Error(err.Error(), sdkService)
//	}
//
//	accountId_First, err := helper.CreateAccount(sdkService, "111")
//	accountId_Second, err := helper.CreateAccount(sdkService, "222")
//
//	conn_first, msgChan_First, err := helper.AccountWebSocket(accountId_First)
//	conn_second, msgChan_Second, err := helper.AccountWebSocket(accountId_Second)
//
//	done := make(chan struct{})
//	receivedChan := make(chan interface{})
//
//	//defer conn_first.Close()
//	//defer conn_second.Close()
//	defer close(msgChan_First)
//	defer close(msgChan_Second)
//
//	read := func(conn *websocket.Conn, readChan chan []byte, accountId uuid.UUID) {
//
//		for {
//			select {
//			case msg := <-readChan:
//				//log.Println("[First]: " + string(msg))
//				message := &helper.WSChatResponse{}
//				_ = json.Unmarshal(msg, message)
//				for _, m := range message.Data.Messages {
//					var direct string
//					if m.AccountId == accountId {
//						direct = "send"
//					} else {
//						direct = "received"
//
//						if chatId != uuid.Nil {
//							err := helper.SendReadStatus(conn, chatId, m.Id)
//							receivedChan <- 0
//							if err != nil {
//								t.Fatal(err)
//							} else {
//								log.Printf("%s read sent \n", m.Id)
//							}
//						}
//
//					}
//
//					log.Printf("%s [%s]: %s (id = %s)\n", direct, accountId, m.Text, m.Id)
//				}
//
//			case <-done:
//				return
//			}
//		}
//	}
//
//	go read(conn_first, msgChan_First, accountId_First)
//	go read(conn_second, msgChan_Second, accountId_Second)
//
//	newChatRs, err := helper.NewChatAndSubscribe(sdkService, accountId_First, "", "ref1", "client")
//	if err != nil || newChatRs.Data.RoomId == uuid.Nil {
//		t.Fatal("New chat creation failed")
//	}
//	chatId = newChatRs.Data.RoomId
//
//	subscribeRs, err := helper.ChatSubscribe(sdkService, chatId, accountId_Second, "ref1", "client")
//	if err != nil || !subscribeRs.Data.Result {
//		t.Fatal("Subscription failed")
//	}
//
//	err = helper.SendMessage(conn_first, accountId_First, server.EventMessage, &models.WSChatMessageDataRequest{
//		RoomId: newChatRs.Data.RoomId,
//		Type:   "message",
//		Text:   "привет второй",
//		Params: map[string]string{"param1": "value1", "param2": "value2"},
//	})
//
//	err = helper.SendMessage(conn_second, accountId_Second, server.EventMessage, &models.WSChatMessageDataRequest{
//		RoomId: newChatRs.Data.RoomId,
//		Type:   "message",
//		Text:   "привет первый",
//		Params: map[string]string{"param1": "value1", "param2": "value2"},
//	})
//
//	if err != nil {
//		t.Fatal("Failed")
//	}
//
//	var received = 0
//	for {
//		select {
//			case <-receivedChan:
//				received++
//				if received == 2 {
//					return
//				}
//			case <-time.After(10 * time.Second):
//				t.Fatal("Test failed. Timeout")
//		}
//	}
//
//}
//
//func TestGetChatInfo_Success(t *testing.T) {
//
//	sdkService, err := helper.InitSdk()
//	if err != nil {
//		t.Fatal(err.Error(), sdkService)
//	}
//
//	chatInfo, err := helper.GetChatInfo(sdkService,
//							uuid.FromStringOrNil("c7dc4c3c-7d88-4a4e-ae90-2b1775158405"),
//							uuid.FromStringOrNil("093d596a-4299-4ad1-9f77-4677adb3ce96"),
//							"")
//
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	for _, item := range chatInfo.Data {
//		log.Println(item)
//	}
//
//}
//
//func TestGetChatsByAccount_Success(t *testing.T) {
//
//	sdkService, err := helper.InitSdk()
//	if err != nil {
//		t.Fatal(err.Error(), sdkService)
//	}
//
//	chats, err := helper.GetChatsByAccount(sdkService,
//		uuid.FromStringOrNil("093d596a-4299-4ad1-9f77-4677adb3ce96"),
//		"")
//
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	for _, item := range chats.Data {
//		log.Println(item)
//	}
//
//}

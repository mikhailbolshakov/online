package helper

import (
	r "chats/repository/room"
	"chats/server"
	"chats/system"
	"encoding/json"
	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
	"log"
)

type WSChatResponse struct {
	Type string                     `json:"type"`
	Data WSChatMessagesDataResponse `json:"data"`
}
type WSChatMessagesDataResponse struct {
	Messages []server.WSChatMessagesDataMessageResponse `json:"messages"`
	Accounts []server.Account                              `json:"accounts"`
}


func SendMessage(socket *websocket.Conn, accountId uuid.UUID, messageType string, message *server.WSChatMessageDataRequest) error {

	if message.ClientMessageId == "" {
		message.ClientMessageId = system.Uuid().String()
	}

	msgRq := &server.WSChatMessagesRequest{
		AccountId: accountId,
		Type:      messageType,
		Data: server.WSChatMessagesDataRequest{
			Messages: []server.WSChatMessageDataRequest{*message},
		},
	}

	request, err := json.Marshal(msgRq)
	if err != nil {
		return err
	}

	err = socket.WriteMessage(websocket.TextMessage, request)
	if err != nil {
		return err
	}
	return nil
}

func SendReadStatus(socket *websocket.Conn, roomId uuid.UUID, messageId uuid.UUID) error {

	msgRq := &server.WSChatMessageStatusRequest{
		Type: server.EventMessageStatus,
		Data: server.WSChatMessageStatusDataRequest{
			Status:    r.MessageStatusRead,
			RoomId:    roomId,
			MessageId: messageId,
		},
	}

	request, err := json.Marshal(msgRq)
	if err != nil {
		return err
	}

	err = socket.WriteMessage(websocket.TextMessage, request)
	if err != nil {
		return err
	}
	return nil
}

func ReadMessages(conn *websocket.Conn,
	readChan <-chan []byte,
	roomId uuid.UUID,
	accountId uuid.UUID,
	receivedChan chan interface{},
	doneChan <-chan interface{},
	sendReadStatus bool) {

	for {
		select {
		case msg := <-readChan:
			//log.Println("[First]: " + string(msg))
			message := &WSChatResponse{}
			_ = json.Unmarshal(msg, message)
			for _, m := range message.Data.Messages {
				var direct string
				if m.AccountId == accountId {
					direct = "send"
				} else {
					direct = "received"

					if roomId != uuid.Nil {
						if sendReadStatus {
							err := SendReadStatus(conn, roomId, m.Id)
							if err != nil {
								log.Fatalf("Error: %v", err)
								return
							} else {
								log.Printf("%s read sent \n", m.Id)
							}
						}
						receivedChan <- true
					}

				}

				log.Printf("%s [%s]: %s (id = %s)\n", direct, accountId, m.Text, m.Id)
			}

		case <-doneChan:
			return
		}
	}
}


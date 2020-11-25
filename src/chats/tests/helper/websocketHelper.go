package helper

import (
	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
	"log"
	"net/http"
)

func AccountWebSocket(accountId uuid.UUID) (*websocket.Conn, chan []byte, error) {

	header := http.Header{}
	c, _, err := websocket.DefaultDialer.Dial( "ws://localhost:8000/ws/?token=" + accountId.String(), header)
	if err != nil {
		return nil, nil, err
	}

	done := make(chan struct{})
	readMessageChan := make(chan []byte)

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read error:", err)
				return
			}
			readMessageChan <- message
		}
	}()

	return c, readMessageChan, nil

	//ticker := time.NewTicker(time.Second * 5)
	//defer ticker.Stop()
	//
	//interrupt := make(chan os.Signal, 1)
	//signal.Notify(interrupt, os.Interrupt)
	//
	//for {
	//	select {
	//	case <-done:
	//		return
	//	case t := <-ticker.C:
	//		err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
	//		if err != nil {
	//			log.Println("write:", err)
	//			return
	//		}
	//	case <-interrupt:
	//		log.Println("interrupt")
	//
	//		// Cleanly close the connection by sending a close message and then
	//		// waiting (with timeout) for the server to close the connection.
	//		err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	//		if err != nil {
	//			log.Println("write close:", err)
	//			return
	//		}
	//		select {
	//		case <-done:
	//			//case <-time.After(time.Second):
	//		}
	//		return
	//	}
	//}

}

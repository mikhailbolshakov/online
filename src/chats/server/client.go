package server

import (
	"chats/database"
	"chats/models"
	"chats/service"
	"fmt"
	"github.com/gorilla/websocket"
	"gitlab.medzdrav.ru/health-service/go-sdk"
	"log"
	"strconv"
	"sync"
	"time"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4608
)

type Client struct {
	hub             *Hub
	conn            *websocket.Conn
	sendChan        chan []byte
	uniqId          string
	rooms           map[uint]*Room
	subscribes      map[uint]database.SubscribeUserModel
	subscribesMutex sync.Mutex
	user            *sdk.UserModel
}

func (c *Client) send(message []byte) {
	defer func() {
		if r := recover(); r != nil {
			_, ok := r.(error)
			if !ok {
				err := fmt.Errorf("%v", r)
				service.SetPanic(&models.SystemError{
					Error:   err,
					Message: WsSendMessageError,
					Code:    WsSendMessageErrorCode,
					Data:    message,
				})
			}
		}
	}()

	c.sendChan <- message
}

func (c *Client) Write() {
	pingTicker := time.NewTicker(pingPeriod)
	defer func() {
		pingTicker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.sendChan:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			if !ok {
				log.Println("Хаб закрыл канал", c.user.Id) //	TODO
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			writer, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			log.Println("message to client: ", string(message)) //	TODO

			writer.Write(message)
			if err := writer.Close(); err != nil {
				return
			}
		case <-pingTicker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

func (c *Client) Read() {
	defer func() {
		log.Println("Отключаем клиента", c.user.Id) //	TODO
		c.hub.unregisterChan <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		messageType, message, err := c.conn.ReadMessage()
		if err != nil {
			service.SetError(&models.SystemError{
				Error:   err,
				Message: WsConnReadMessageError + "; messageType: " + strconv.Itoa(messageType),
				Code:    WsConnReadMessageErrorCode,
				Data:    []byte(message),
			})
			log.Println(">>>>>> ReadMessageError:", err) //	TODO
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("> > > Read sentry: %v", err) //	TODO
			}
			break
		}

		go c.hub.onClientMessage(message, c)
	}
}

func (c *Client) SetSubscribers(data map[uint]database.SubscribeUserModel) {
	c.subscribesMutex.Lock()
	c.subscribes = data
	c.subscribesMutex.Unlock()
}

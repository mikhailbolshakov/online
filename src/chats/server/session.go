package server

import (
	"chats/models"
	"chats/sdk"
	"chats/system"
	"fmt"
	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
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

type Session struct {
	hub             *Hub
	conn            *websocket.Conn
	sendChan        chan []byte
	sessionId       uuid.UUID
	rooms           map[uuid.UUID]*Room
	subscribers     map[uuid.UUID]models.AccountSubscriber
	subscribesMutex sync.Mutex
	account         *sdk.Account
}

func InitSession(h *Hub, conn *websocket.Conn) *Session {
	return 	&Session{
		hub:        h,
		conn:       conn,
		sendChan:   make(chan []byte, 256),
		sessionId:  system.Uuid(),
	}
}

func (c *Session) send(message []byte) {
	defer func() {
		if r := recover(); r != nil {
			_, ok := r.(error)
			if !ok {
				err := fmt.Errorf("%v", r)
				system.ErrHandler.SetError(&system.Error{
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

func (c *Session) Write() {
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
				log.Println("Хаб закрыл канал", c.account.Id) //	TODO
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

func (c *Session) Read() {
	defer func() {
		log.Println("Отключаем клиента", c.account.Id) //	TODO
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
		log.Printf("Message from socket: %v \n", string(message))

		if err != nil {
			system.ErrHandler.SetError(&system.Error{
				Error:   err,
				Message: WsConnReadMessageError + "; messageType: " + strconv.Itoa(messageType),
				Code:    WsConnReadMessageErrorCode,
				Data:    []byte(message),
			})
			log.Println(">>>>>> ReadMessageError:", err) //	TODO
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("> > > Read sentry: %v", err) 	 //	TODO
			}
			break
		}

		go c.hub.onMessage(message, c)
	}
}

func (c *Session) SetSubscribers(data map[uuid.UUID]models.AccountSubscriber) {
	c.subscribesMutex.Lock()
	defer c.subscribesMutex.Unlock()
	c.subscribers = data
}

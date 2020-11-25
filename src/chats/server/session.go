package server

import (
	"chats/app"
	r "chats/repository/room"
	"chats/system"
	"fmt"
	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
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
	// TODO: remove link to repository
	subscribers     map[uuid.UUID]r.AccountSubscriber
	subscribesMutex sync.Mutex
	account         *Account
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
				app.E().SetError(system.SysErr(err, system.WsSendMessageErrorCode, message))
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
				app.L().Debug("Хаб закрыл канал", c.account.Id) //	TODO
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			writer, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			app.L().Debug("message to client: ", string(message)) //	TODO

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
		app.L().Debug("Отключаем клиента", c.account.Id) //	TODO
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
		_, message, err := c.conn.ReadMessage()
		app.L().Debugf("Message from socket: %v \n", string(message))

		if err != nil {
			app.E().SetError(system.SysErr(err, system.WsConnReadMessageErrorCode, message))
			app.L().Debug(">>>>>> ReadMessageError:", err) //	TODO
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				app.L().Debugf("> > > Read sentry: %v", err) 	 //	TODO
			}
			break
		}

		go c.hub.onMessage(message, c)
	}
}

func (c *Session) SetSubscribers(data map[uuid.UUID]r.AccountSubscriber) {
	c.subscribesMutex.Lock()
	defer c.subscribesMutex.Unlock()
	c.subscribers = data
}

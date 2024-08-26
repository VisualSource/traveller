package socket

import (
	"bytes"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Client struct {
	hub       *Hub
	conn      *websocket.Conn
	send      chan []byte
	UserId    string
	SessionId string
}

type clientBoardcastMessage struct {
	Message string
	Target  string
}

func (c *Client) GetId() string {
	return strings.Join([]string{c.UserId, c.SessionId}, ":")
}

func (c *Client) readPump() {
	defer func() {
		c.conn.Close()
		c.hub.unregister <- c.GetId()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(appData string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) || websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				break
			}
			log.Printf("error: %v\n", err)
			continue
		}
		msg = bytes.TrimSpace(bytes.Replace(msg, newline, space, -1))

		var data ClientMessage
		err = json.Unmarshal(msg, &data)
		if err != nil {
			log.Printf("JSON unmarshal error: %v\n", err)
			continue
		}

		switch data.ContentType {
		case "BroadcastMessage":
			var d clientBoardcastMessage
			err = data.parsePayload(&d)
			if err != nil {
				log.Println(err)
				continue
			}

			c.hub.broadcast <- NewBroadcastMessage(c.SessionId, d.Message, d.Target, c.UserId)
		case "PrivateMessage":
			var d clientBoardcastMessage
			err = data.parsePayload(&d)
			if err != nil {
				log.Println(err)
				continue
			}
			c.hub.message <- NewPrivateMessage(c.UserId, d.Target, c.SessionId, d.Message)
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(msg)
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func ServeWs(hub *Hub, c echo.Context) error {
	sess, err := session.Get("session", c)
	if err != nil {
		return err
	}

	conn, err := upgrader.Upgrade(c.Response().Writer, c.Request(), nil)
	if err != nil {
		log.Println(err)
		return err
	}

	sessionId := c.Param("sessionId")
	userId := sess.Values["sub"].(string)

	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256), SessionId: sessionId, UserId: userId}
	client.hub.register <- client

	go client.writePump()
	go client.readPump()
	return nil
}

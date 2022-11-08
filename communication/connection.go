package communication

import (
	"github.com/gorilla/websocket"
	"time"
)

const writeWait = 10 * time.Second

type Connection struct {
	Id       string
	Name     string
	Ws       *websocket.Conn
	Send     chan []byte
	IsAdmin  bool
	Language string
}

func (c *Connection) Write(mt int, payload []byte) error {
	c.Ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.Ws.WriteMessage(mt, payload)
}

func (c Connection) GetId() string {
	return c.Id
}

func (c Connection) GetName() string {
	return c.Name
}

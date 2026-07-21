package transport

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/zmhuanf/feng/internal/core"
	"github.com/zmhuanf/feng/internal/protocol"
)

var Upgrader = websocket.Upgrader{
	CheckOrigin: func(*http.Request) bool { return true },
}

type Conn struct {
	conn        *websocket.Conn
	codec       core.Codec
	messageType int
	lock        sync.Mutex
}

func NewConn(conn *websocket.Conn, codec core.Codec) *Conn {
	return &Conn{conn: conn, codec: codec, messageType: codec.MessageType()}
}

func Dial(url string, codec core.Codec) (*Conn, error) {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	return NewConn(conn, codec), nil
}

func (c *Conn) Read() (*protocol.Message, error) {
	for {
		messageType, data, err := c.conn.ReadMessage()
		if err != nil {
			return nil, err
		}
		if messageType != c.messageType {
			continue
		}
		var msg protocol.Message
		if err := c.codec.Unmarshal(data, &msg); err != nil {
			return nil, err
		}
		return &msg, nil
	}
}

func (c *Conn) Send(msg *protocol.Message) error {
	data, err := c.codec.Marshal(msg)
	if err != nil {
		return err
	}
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.conn.WriteMessage(c.messageType, data)
}

func (c *Conn) Close() error {
	return c.conn.Close()
}

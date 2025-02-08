package fibril

import (
	"github.com/gofiber/contrib/websocket"
	"github.com/google/uuid"
	"github.com/lishank0119/pubsub"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type Client struct {
	UUID string
	hub  *Hub
	conn *websocket.Conn
	send chan box
	open atomic.Bool
	opt  *option
	exit chan bool
	keys sync.Map
	once sync.Once
	sub  *pubsub.Subscriber
}

func (c *Client) Subscribe(topic string, handler pubsub.HandlerFunc) {
	c.sub.Subscribe(topic, handler)
}

func (c *Client) isOpen() bool {
	return c.open.Load()
}

func (c *Client) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *Client) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *Client) GetWsConnect() *websocket.Conn { return c.conn }

func (c *Client) writePump() {
	ticker := time.NewTicker(c.opt.pingPeriod)
	defer ticker.Stop()

loop:
	for {
		select {
		case b, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(c.opt.writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if b.t == websocket.CloseMessage {
				if err := c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, string(b.msg))); err != nil {
					c.opt.errorHandler(c, err)
				}

				time.Sleep(c.opt.disconnectDelayClose)
				break loop
			}

			err := c.conn.WriteMessage(b.t, b.msg)
			if err != nil {
				c.opt.errorHandler(c, err)
				return
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(c.opt.writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case _, ok := <-c.exit:
			if !ok {
				break loop
			}
		}
	}

	c.close()
}

func (c *Client) readPump() {
	defer c.destroy()

	c.conn.SetReadLimit(c.opt.maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(c.opt.pongWait))

	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(c.opt.pongWait))
		c.opt.pongHandler(c)
		return nil
	})

	if c.opt.closeHandler != nil {
		c.conn.SetCloseHandler(func(code int, text string) error {
			return c.opt.closeHandler(c, code, text)
		})
	}

	for {
		t, message, err := c.conn.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNormalClosure) {
				c.opt.errorHandler(c, err)
			}
			break
		}

		switch t {
		case websocket.TextMessage:
			c.opt.textMessageHandler(c, string(message))
		case websocket.BinaryMessage:
			c.opt.binaryMessageHandler(c, message)
		}
	}
}

func (c *Client) Disconnect(closeMsg string) {
	message := box{t: websocket.CloseMessage, msg: []byte(closeMsg)}
	c.writeMessage(message)
}

func (c *Client) destroy() {
	c.sub.UnsubscribeAll()
	c.hub.unregisterClient(c)
	c.close()
}

func (c *Client) close() {
	c.once.Do(func() {
		c.open.Store(false)
		_ = c.conn.SetReadDeadline(time.Now())
		err := c.conn.Close()
		if err != nil {
			c.opt.errorHandler(c, err)
			return
		}
		close(c.exit)
	})
}

func (c *Client) StoreKey(key any, value any) {
	c.keys.Store(key, value)
}

func (c *Client) DeleteKey(key any) {
	c.keys.Delete(key)
}

func (c *Client) GetKey(key any) (any, bool) {
	return c.keys.Load(key)
}

func (c *Client) writeMessage(message box) {
	if !c.isOpen() {
		c.opt.errorHandler(c, ErrWriteClosed)
		return
	}

	select {
	case c.send <- message:
	default:
		c.opt.errorHandler(c, ErrMessageBufferFull)
	}
}

func (c *Client) WriteText(msg string) error {
	if !c.isOpen() {
		return ErrClientClosed
	}

	c.writeMessage(box{t: websocket.TextMessage, msg: []byte(msg)})

	return nil
}

func (c *Client) WriteBinary(msg []byte) error {
	if !c.isOpen() {
		return ErrClientClosed
	}

	c.writeMessage(box{t: websocket.BinaryMessage, msg: msg})

	return nil
}

func newClient(hub *Hub, conn *websocket.Conn, option *option, keys map[any]any) *Client {
	client := &Client{
		UUID: uuid.New().String(),
		hub:  hub,
		opt:  option,
		conn: conn,
		send: make(chan box, option.messageBufferSize),
		sub:  hub.pubSub.NewSubscriber(option.messageBufferSize),
		exit: make(chan bool),
	}

	if keys != nil {
		for k, v := range keys {
			client.StoreKey(k, v)
		}
	}

	client.hub.registerClient(client)

	client.open.Store(true)

	option.connectHandler(client)

	go client.writePump()

	client.readPump()

	return client
}

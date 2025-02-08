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

// Client represents a WebSocket client connection.
type Client struct {
	UUID string             // Unique identifier for the client
	hub  *Hub               // Reference to the Hub managing this client
	conn *websocket.Conn    // The WebSocket connection
	send chan box           // Channel for sending messages to the client
	open atomic.Bool        // Indicates if the connection is open
	opt  *option            // Configuration options for the client
	exit chan bool          // Channel to signal the client to exit
	keys sync.Map           // Key-value store for custom client data
	once sync.Once          // Ensures the close operation is performed only once
	sub  *pubsub.Subscriber // Subscriber for Pub/Sub messages
}

// Subscribe subscribes the client to a specific topic with a handler function.
func (c *Client) Subscribe(topic string, handler pubsub.HandlerFunc) {
	c.sub.Subscribe(topic, handler)
}

// isOpen checks if the client's WebSocket connection is open.
func (c *Client) isOpen() bool {
	return c.open.Load()
}

// LocalAddr returns the local network address of the client.
func (c *Client) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

// RemoteAddr returns the remote network address of the client.
func (c *Client) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// GetWsConnect returns the WebSocket connection of the client.
func (c *Client) GetWsConnect() *websocket.Conn { return c.conn }

// writePump handles outgoing messages to the client and manages keep-alive pings.
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

// readPump reads incoming messages from the WebSocket connection.
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

// Disconnect initiates a graceful disconnection from the client with a close message.
func (c *Client) Disconnect(closeMsg string) {
	message := box{t: websocket.CloseMessage, msg: []byte(closeMsg)}
	c.writeMessage(message)
}

// destroy performs cleanup operations when the client is disconnected.
func (c *Client) destroy() {
	c.sub.UnsubscribeAll()
	c.hub.unregisterClient(c)
	c.close()
}

// close safely closes the client's WebSocket connection and signals the exit channel.
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

// StoreKey stores a key-value pair associated with the client.
func (c *Client) StoreKey(key any, value any) {
	c.keys.Store(key, value)
}

// DeleteKey removes a key-value pair associated with the client.
func (c *Client) DeleteKey(key any) {
	c.keys.Delete(key)
}

// GetKey retrieves a value by key from the client's key-value store.
func (c *Client) GetKey(key any) (any, bool) {
	return c.keys.Load(key)
}

// writeMessage sends a message to the client's send channel.
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

// WriteText sends a text message to the client.
func (c *Client) WriteText(msg string) error {
	if !c.isOpen() {
		return ErrClientClosed
	}

	c.writeMessage(box{t: websocket.TextMessage, msg: []byte(msg)})
	return nil
}

// WriteBinary sends a binary message to the client.
func (c *Client) WriteBinary(msg []byte) error {
	if !c.isOpen() {
		return ErrClientClosed
	}

	c.writeMessage(box{t: websocket.BinaryMessage, msg: msg})
	return nil
}

// newClient initializes a new WebSocket client and starts its read and write loops.
func newClient(hub *Hub, conn *websocket.Conn, option *option, keys map[any]any) *Client {
	client := &Client{
		UUID: uuid.New().String(),
		hub:  hub,
		opt:  option,
		conn: conn,
		send: make(chan box, option.messageBufferSize),
		sub:  hub.pubSub.NewSubscriber(),
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

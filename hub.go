package fibril

import (
	"github.com/gofiber/contrib/websocket"
	"github.com/lishank0119/shardingmap"
)

type Hub struct {
	opt       *option
	clientMap *shardingmap.ShardingMap[string, *Client]
	broadcast chan box
}

func (h *Hub) run() {
	for {
		select {
		case b := <-h.broadcast:
			h.clientMap.ForEach(func(uuid string, client *Client) {
				if b.filter == nil {
					client.send <- b
				} else if b.filter(client) {
					client.send <- b
				}
			})
		}
	}
}

func (h *Hub) registerClient(client *Client) {
	h.clientMap.Set(client.UUID, client)
}

func (h *Hub) unregisterClient(client *Client) {
	h.clientMap.Delete(client.UUID)
	h.opt.disconnectHandler(client)
}

func (h *Hub) disconnectClientFilter(closeMsg string, fn filterFunc) {
	h.clientMap.ForEach(func(uuid string, client *Client) {
		if fn == nil {
			client.Disconnect(closeMsg)
		} else if fn(client) {
			client.Disconnect(closeMsg)
		}
	})
}

func (h *Hub) disconnectClient(closeMsg string, uuid string) error {
	if client, ok := h.clientMap.Get(uuid); ok {
		client.Disconnect(closeMsg)
		return nil
	}

	return ErrClientNotFound
}

func (h *Hub) broadcastText(msg string, fn func(*Client) bool) {
	message := box{t: websocket.TextMessage, msg: []byte(msg), filter: fn}
	h.broadcast <- message
}

func (h *Hub) broadcastBinary(msg []byte, fn func(*Client) bool) {
	message := box{t: websocket.TextMessage, msg: msg, filter: fn}
	h.broadcast <- message
}

func (h *Hub) sendTextToClient(uuid string, msg string) error {
	if client, ok := h.clientMap.Get(uuid); ok {
		client.writeMessage(box{t: websocket.TextMessage, msg: []byte(msg)})
		return nil
	}

	return ErrClientNotFound
}

func (h *Hub) sendBinaryToClient(uuid string, msg []byte) error {
	if client, ok := h.clientMap.Get(uuid); ok {
		client.writeMessage(box{t: websocket.BinaryMessage, msg: msg})
		return nil
	}

	return ErrClientNotFound
}

func newHub(opt *option) *Hub {
	m := shardingmap.New[string, *Client](
		shardingmap.WithShardCount[string, *Client](opt.shardCount),
	)

	return &Hub{
		opt:       opt,
		clientMap: m,
		broadcast: make(chan box),
	}
}

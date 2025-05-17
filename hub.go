package fibril

import (
	"context"
	"github.com/gofiber/contrib/websocket"
	"github.com/lishank0119/pubsub"
	"github.com/lishank0119/shardingmap"
)

// Hub manages WebSocket clients, broadcasting messages, and Pub/Sub communications.
type Hub struct {
	opt       *option                                   // Configuration options for the Hub
	clientMap *shardingmap.ShardingMap[string, *Client] // A sharded map to efficiently manage connected clients
	broadcast chan box                                  // Channel for broadcasting messages to clients
	pubSub    *pubsub.PubSub                            // Internal Pub/Sub system for message distribution
}

// subscriberCount returns the number of subscribers for a given topic.
func (h *Hub) subscriberCount(topic string) int {
	return h.pubSub.SubscriberCount(topic)
}

// listTopics returns a list of all active topics currently subscribed to.
func (h *Hub) listTopics() []string {
	return h.pubSub.ListTopics()
}

// getClient returns the client associated with the given UUID.
func (h *Hub) getClient(uuid string) (*Client, bool) {
	return h.clientMap.Get(uuid)
}

// forEachClientWithContext iterates over all clients and stops if context is cancelled.
func (h *Hub) forEachClientWithContext(ctx context.Context, callback func(uuid string, client *Client)) {
	h.clientMap.ForEach(func(uuid string, client *Client) {
		select {
		case <-ctx.Done():
			return
		default:
			callback(uuid, client)
		}
	})
}

// forEachClient calls the Hub’s ForEachClient to iterate over all clients.
func (h *Hub) forEachClient(callback func(uuid string, client *Client)) {
	h.clientMap.ForEach(func(uuid string, client *Client) {
		callback(uuid, client)
	})
}

func (h *Hub) clientLen() int {
	return h.clientMap.Len()
}

// publish sends a message to a specific topic using the Pub/Sub system.
func (h *Hub) publish(topic string, msg []byte) error {
	return h.pubSub.Publish(topic, msg)
}

// run continuously listens for messages on the broadcast channel and dispatches them to clients.
func (h *Hub) run() {
	for {
		select {
		case b := <-h.broadcast:
			h.clientMap.ForEach(func(uuid string, client *Client) {
				if b.filter == nil {
					client.writeMessage(b) // Send message to all clients if no filter is set
				} else if b.filter(client) {
					client.writeMessage(b) // Send message only to clients that pass the filter
				}
			})
		}
	}
}

// registerClient adds a new client to the client map.
func (h *Hub) registerClient(client *Client) {
	h.clientMap.Set(client.GetUUID(), client)
}

// unregisterClient removes a client from the client map and triggers the disconnect handler.
func (h *Hub) unregisterClient(client *Client) {
	h.clientMap.Delete(client.GetUUID())
	h.opt.disconnectHandler(client)
}

// disconnectAll disconnects all clients with the given close message.
func (h *Hub) disconnectAll(closeMsg string) {
	h.disconnectClientFilter(closeMsg, nil)
}

// disconnectClientFilter disconnects clients based on a provided filter function.
// If no filter is provided, all clients will be disconnected with the given close message.
func (h *Hub) disconnectClientFilter(closeMsg string, fn filterFunc) {
	h.clientMap.ForEach(func(uuid string, client *Client) {
		if fn == nil {
			client.Disconnect(closeMsg) // Disconnect all clients if no filter is provided
		} else if fn(client) {
			client.Disconnect(closeMsg) // Disconnect only clients that match the filter
		}
	})
}

// disconnectClient disconnects a specific client identified by its UUID.
func (h *Hub) disconnectClient(closeMsg string, uuid string) error {
	if client, ok := h.clientMap.Get(uuid); ok {
		client.Disconnect(closeMsg)
		return nil
	}
	return ErrClientNotFound
}

// broadcastText sends a text message to all clients that match the filter function.
// If no filter is provided, the message will be sent to all clients.
func (h *Hub) broadcastText(msg string, fn func(*Client) bool) {
	message := box{t: websocket.TextMessage, msg: []byte(msg), filter: fn}
	h.broadcast <- message
}

// broadcastBinary sends a binary message to all clients that match the filter function.
// If no filter is provided, the message will be sent to all clients.
func (h *Hub) broadcastBinary(msg []byte, fn func(*Client) bool) {
	message := box{t: websocket.BinaryMessage, msg: msg, filter: fn}
	h.broadcast <- message
}

// sendTextToClient sends a text message to a specific client identified by its UUID.
func (h *Hub) sendTextToClient(uuid string, msg string) error {
	if client, ok := h.clientMap.Get(uuid); ok {
		client.writeMessage(box{t: websocket.TextMessage, msg: []byte(msg)})
		return nil
	}
	return ErrClientNotFound
}

// sendBinaryToClient sends a binary message to a specific client identified by its UUID.
func (h *Hub) sendBinaryToClient(uuid string, msg []byte) error {
	if client, ok := h.clientMap.Get(uuid); ok {
		client.writeMessage(box{t: websocket.BinaryMessage, msg: msg})
		return nil
	}
	return ErrClientNotFound
}

// newHub initializes and returns a new Hub instance with the provided options.
func newHub(opt *option) *Hub {
	m := shardingmap.New[string, *Client](
		shardingmap.WithShardCount[string, *Client](opt.shardCount),
	)

	return &Hub{
		opt:       opt,
		clientMap: m,
		pubSub: pubsub.NewPubSub(&pubsub.Config{
			BucketNum:           opt.shardCount,            // Number of buckets for sharding Pub/Sub messages
			BucketMessageBuffer: opt.messageBufferSize * 2, // Buffer size for each Pub/Sub bucket
		}),
		broadcast: make(chan box), // Channel for broadcasting messages
	}
}

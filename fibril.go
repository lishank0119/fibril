package fibril

import "github.com/gofiber/contrib/websocket"

// Fibril represents the core WebSocket server, managing clients and message broadcasting.
type Fibril struct {
	option *option // Configuration options for the Fibril instance
	hub    *Hub    // Central hub responsible for managing clients and messages
}

// GetClient proxies the Hub's getClient to retrieve a connected client by UUID.
func (f *Fibril) GetClient(uuid string) (*Client, bool) {
	return f.hub.getClient(uuid)
}

// ForEachClient iterates over all connected clients and applies the callback.
func (f *Fibril) ForEachClient(callback func(uuid string, client *Client)) {
	f.hub.forEachClient(callback)
}

// ClientLen returns the number of active WebSocket clients connected to the hub.
func (f *Fibril) ClientLen() int {
	return f.hub.clientMap.Len()
}

// Publish sends a message to all subscribers of the specified topic.
func (f *Fibril) Publish(topic string, msg []byte) error {
	return f.hub.publish(topic, msg)
}

// SendTextToClient sends a text message to a specific client identified by UUID.
func (f *Fibril) SendTextToClient(uuid string, msg string) error {
	return f.hub.sendTextToClient(uuid, msg)
}

// SendBinaryToClient sends a binary message to a specific client identified by UUID.
func (f *Fibril) SendBinaryToClient(uuid string, msg []byte) error {
	return f.hub.sendBinaryToClient(uuid, msg)
}

// BroadcastText broadcasts a text message to all connected clients.
func (f *Fibril) BroadcastText(msg string) {
	f.hub.broadcastText(msg, nil)
}

// BroadcastTextFilter broadcasts a text message to clients that meet the specified filter condition.
func (f *Fibril) BroadcastTextFilter(msg string, fn func(*Client) bool) {
	f.hub.broadcastText(msg, fn)
}

// BroadcastBinary broadcasts a binary message to all connected clients.
func (f *Fibril) BroadcastBinary(msg []byte) {
	f.hub.broadcastBinary(msg, nil)
}

// BroadcastBinaryFilter broadcasts a binary message to clients that meet the specified filter condition.
func (f *Fibril) BroadcastBinaryFilter(msg []byte, fn func(*Client) bool) {
	f.hub.broadcastBinary(msg, fn)
}

// RegisterClient registers a new WebSocket client without additional metadata.
func (f *Fibril) RegisterClient(conn *websocket.Conn) {
	newClient(f.hub, conn, f.option, nil)
}

// RegisterClientWithKeys registers a new WebSocket client with custom key-value pairs.
func (f *Fibril) RegisterClientWithKeys(conn *websocket.Conn, keys map[any]any) {
	newClient(f.hub, conn, f.option, keys)
}

// TextMessageHandler sets the handler function for incoming text messages from clients.
func (f *Fibril) TextMessageHandler(handler func(*Client, string)) {
	f.option.textMessageHandler = handler
}

// BinaryMessageHandler sets the handler function for incoming binary messages from clients.
func (f *Fibril) BinaryMessageHandler(handler func(*Client, []byte)) {
	f.option.binaryMessageHandler = handler
}

// ErrorHandler sets the handler function for managing errors.
func (f *Fibril) ErrorHandler(handler handleErrorFunc) {
	f.option.errorHandler = handler
}

// CloseHandler sets the handler function for client connection closures.
func (f *Fibril) CloseHandler(handler handleCloseFunc) {
	f.option.closeHandler = handler
}

// ConnectHandler sets the handler function triggered when a client connects.
func (f *Fibril) ConnectHandler(handler handleClientFunc) {
	f.option.connectHandler = handler
}

// DisconnectHandler sets the handler function triggered when a client disconnects.
func (f *Fibril) DisconnectHandler(handler handleClientFunc) {
	f.option.disconnectHandler = handler
}

// PongHandler sets the handler function for managing WebSocket pong messages.
func (f *Fibril) PongHandler(handler handleClientFunc) {
	f.option.pongHandler = handler
}

// DisconnectClient disconnects a specific client identified by UUID with an optional close message.
func (f *Fibril) DisconnectClient(closeMsg string, uuid string) error {
	return f.hub.disconnectClient(closeMsg, uuid)
}

// DisconnectClientFilter disconnects clients that meet the specified filter condition with an optional close message.
func (f *Fibril) DisconnectClientFilter(closeMsg string, fn filterFunc) {
	f.hub.disconnectClientFilter(closeMsg, fn)
}

// New initializes a new Fibril instance with optional configuration functions.
func New(args ...OptFunc) *Fibril {
	opt := defaultOption() // Apply default options
	for _, optFunc := range args {
		optFunc(opt) // Apply custom options
	}

	hub := newHub(opt) // Create the central hub
	go hub.run()       // Start the hub event loop

	return &Fibril{
		option: opt,
		hub:    hub,
	}
}

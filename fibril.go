package fibril

import "github.com/gofiber/contrib/websocket"

type Fibril struct {
	option *option
	hub    *Hub
}

func (f *Fibril) Publish(topic string, msg []byte) {
	f.hub.publish(topic, msg)
}

func (f *Fibril) SendTextToClient(uuid string, msg string) error {
	return f.hub.sendTextToClient(uuid, msg)
}

func (f *Fibril) SendBinaryToClient(uuid string, msg []byte) error {
	return f.hub.sendBinaryToClient(uuid, msg)
}

func (f *Fibril) BroadcastText(msg string) {
	f.hub.broadcastText(msg, nil)
}

func (f *Fibril) BroadcastTextFilter(msg string, fn func(*Client) bool) {
	f.hub.broadcastText(msg, fn)
}

func (f *Fibril) BroadcastBinary(msg []byte) {
	f.hub.broadcastBinary(msg, nil)
}

func (f *Fibril) BroadcastBinaryFilter(msg []byte, fn func(*Client) bool) {
	f.hub.broadcastBinary(msg, fn)
}

func (f *Fibril) RegisterClient(conn *websocket.Conn) {
	newClient(f.hub, conn, f.option, nil)
}

func (f *Fibril) RegisterClientWithKeys(conn *websocket.Conn, keys map[any]any) {
	newClient(f.hub, conn, f.option, keys)
}

func (f *Fibril) TextMessageHandler(handler func(*Client, string)) {
	f.option.textMessageHandler = handler
}

func (f *Fibril) BinaryMessageHandler(handler func(*Client, []byte)) {
	f.option.binaryMessageHandler = handler
}

func (f *Fibril) ErrorHandler(handler handleErrorFunc) {
	f.option.errorHandler = handler
}

func (f *Fibril) CloseHandler(handler handleCloseFunc) {
	f.option.closeHandler = handler
}

func (f *Fibril) ConnectHandler(handler handleClientFunc) {
	f.option.connectHandler = handler
}

func (f *Fibril) DisconnectHandler(handler handleClientFunc) {
	f.option.disconnectHandler = handler
}

func (f *Fibril) PongHandler(handler handleClientFunc) {
	f.option.pongHandler = handler
}

func (f *Fibril) DisconnectClient(closeMsg string, uuid string) error {
	return f.hub.disconnectClient(closeMsg, uuid)
}

func (f *Fibril) DisconnectClientFilter(closeMsg string, fn filterFunc) {
	f.hub.disconnectClientFilter(closeMsg, fn)
}

func New(args ...OptFunc) *Fibril {
	opt := defaultOption()
	for _, optFunc := range args {
		optFunc(opt)
	}

	hub := newHub(opt)

	go hub.run()

	return &Fibril{
		option: opt,
		hub:    hub,
	}
}

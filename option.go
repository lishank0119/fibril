package fibril

import "time"

type handleErrorFunc func(*Client, error)
type handleCloseFunc func(*Client, int, string) error
type handleClientFunc func(*Client)

type option struct {
	shardCount           int
	maxMessageSize       int64
	messageBufferSize    int
	writeWait            time.Duration
	pongWait             time.Duration
	pingPeriod           time.Duration
	disconnectDelayClose time.Duration
	textMessageHandler   func(*Client, string)
	binaryMessageHandler func(*Client, []byte)
	errorHandler         handleErrorFunc
	closeHandler         handleCloseFunc
	connectHandler       handleClientFunc
	disconnectHandler    handleClientFunc
	pongHandler          handleClientFunc
}

type OptFunc func(*option)

func WithShardCount(count int) OptFunc {
	return func(o *option) {
		o.shardCount = count
	}
}

func WithMaxMessageSize(size int64) OptFunc {
	return func(o *option) {
		o.maxMessageSize = size
	}
}

func WithMessageBufferSize(size int) OptFunc {
	return func(o *option) {
		o.messageBufferSize = size
	}
}

func WithWriteWait(wait time.Duration) OptFunc {
	return func(o *option) {
		o.writeWait = wait
	}
}

func WithPongWait(wait time.Duration) OptFunc {
	return func(o *option) {
		o.pongWait = wait
	}
}

func WithPingPeriod(pingPeriod time.Duration) OptFunc {
	return func(o *option) {
		o.pingPeriod = pingPeriod
	}
}

func defaultOption() *option {
	return &option{
		shardCount:           16,
		maxMessageSize:       512,
		messageBufferSize:    256,
		writeWait:            10 * time.Second,
		pongWait:             60 * time.Second,
		pingPeriod:           54 * time.Second,
		disconnectDelayClose: 100 * time.Millisecond,
		textMessageHandler:   func(*Client, string) {},
		binaryMessageHandler: func(*Client, []byte) {},
		errorHandler:         func(*Client, error) {},
		closeHandler:         nil,
		connectHandler:       func(*Client) {},
		disconnectHandler:    func(*Client) {},
		pongHandler:          func(*Client) {},
	}
}

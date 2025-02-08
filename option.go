package fibril

import "time"

// handleErrorFunc defines a function type for handling errors from a client.
type handleErrorFunc func(*Client, error)

// handleCloseFunc defines a function type for handling client connection closure.
// It returns an error if the close operation encounters an issue.
type handleCloseFunc func(*Client, int, string) error

// handleClientFunc defines a function type for handling client-related events
// such as connection, disconnection, or pong responses.
type handleClientFunc func(*Client)

// option holds configuration settings for the WebSocket server behavior.
type option struct {
	shardCount           int                   // Number of shards for load distribution
	maxMessageSize       int64                 // Maximum size of a message (in bytes)
	messageBufferSize    int                   // Buffer size for message channels
	writeWait            time.Duration         // Maximum duration to wait for a write operation to complete
	pongWait             time.Duration         // Duration to wait for a pong response before considering the connection dead
	pingPeriod           time.Duration         // Interval for sending ping messages to clients
	disconnectDelayClose time.Duration         // Delay before closing a disconnected client
	textMessageHandler   func(*Client, string) // Handler for processing text messages
	binaryMessageHandler func(*Client, []byte) // Handler for processing binary messages
	errorHandler         handleErrorFunc       // Handler for processing errors
	closeHandler         handleCloseFunc       // Handler for client connection closure
	connectHandler       handleClientFunc      // Handler triggered when a client connects
	disconnectHandler    handleClientFunc      // Handler triggered when a client disconnects
	pongHandler          handleClientFunc      // Handler triggered when a pong message is received
}

// OptFunc represents a functional option pattern for modifying the option struct.
type OptFunc func(*option)

// WithShardCount sets the number of shards for the WebSocket server.
// If the provided count is less than 2, it defaults to 2 to ensure proper load distribution.
func WithShardCount(count int) OptFunc {
	return func(o *option) {
		if count < 2 {
			o.shardCount = 2
		} else {
			o.shardCount = count
		}
	}
}

// WithMaxMessageSize sets the maximum size of incoming messages (in bytes).
// If the provided size is less than 256 bytes, it defaults to 256 to ensure reliable message processing.
func WithMaxMessageSize(size int64) OptFunc {
	return func(o *option) {
		if size < 256 {
			o.maxMessageSize = 256
		} else {
			o.maxMessageSize = size
		}
	}
}

// WithMessageBufferSize sets the buffer size for storing messages before processing.
func WithMessageBufferSize(size int) OptFunc {
	return func(o *option) {
		o.messageBufferSize = size
	}
}

// WithWriteWait sets the maximum duration to wait for a write operation to complete.
func WithWriteWait(wait time.Duration) OptFunc {
	return func(o *option) {
		o.writeWait = wait
	}
}

// WithPongWait sets the maximum duration to wait for a pong response from the client.
func WithPongWait(wait time.Duration) OptFunc {
	return func(o *option) {
		o.pongWait = wait
	}
}

// WithPingPeriod sets the interval for sending periodic ping messages to clients.
func WithPingPeriod(pingPeriod time.Duration) OptFunc {
	return func(o *option) {
		o.pingPeriod = pingPeriod
	}
}

// defaultOption returns a new option instance with default configuration settings.
func defaultOption() *option {
	return &option{
		shardCount:           16,                       // Default number of shards
		maxMessageSize:       512,                      // Default maximum message size (in bytes)
		messageBufferSize:    256,                      // Default buffer size for messages
		writeWait:            10 * time.Second,         // Default write timeout duration
		pongWait:             60 * time.Second,         // Default pong wait duration
		pingPeriod:           54 * time.Second,         // Default ping period
		disconnectDelayClose: 100 * time.Millisecond,   // Default delay before closing disconnected clients
		textMessageHandler:   func(*Client, string) {}, // Default no-op handler for text messages
		binaryMessageHandler: func(*Client, []byte) {}, // Default no-op handler for binary messages
		errorHandler:         func(*Client, error) {},  // Default no-op handler for errors
		closeHandler:         nil,                      // No default close handler
		connectHandler:       func(*Client) {},         // Default no-op handler for client connections
		disconnectHandler:    func(*Client) {},         // Default no-op handler for client disconnections
		pongHandler:          func(*Client) {},         // Default no-op handler for pong messages
	}
}

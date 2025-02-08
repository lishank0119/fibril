package fibril

// filterFunc defines a function type that takes a *Client as input
// and returns a boolean indicating whether the client meets specific criteria.
type filterFunc func(*Client) bool

// box represents a message container used in the WebSocket system.
// It holds the message type, the actual message data, and an optional filter
// to determine which clients should receive the message.
type box struct {
	t      int        // WebSocket message type (e.g., text or binary)
	msg    []byte     // Actual message content
	filter filterFunc // Optional filter to determine target clients
}

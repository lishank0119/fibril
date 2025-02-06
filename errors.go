package fibril

import "errors"

var (
	ErrClientClosed      = errors.New("client is closed")
	ErrWriteClosed       = errors.New("tried to write to closed a session")
	ErrMessageBufferFull = errors.New("message buffer is full")
	ErrClientNotFound    = errors.New("client not found")
)

[![Go Reference](https://pkg.go.dev/badge/github.com/lishank0119/fibril.svg)](https://pkg.go.dev/github.com/lishank0119/fibril)
[![go.mod](https://img.shields.io/github/go-mod/go-version/lishank0119/fibril)](go.mod)

# Fibril - A WebSocket Library for Go

[中文](README.zh-TW.md)

Fibril is a Go-based WebSocket library built on top of [GoFiber](https://github.com/gofiber/fiber) that provides a
robust and efficient way to handle WebSocket connections, messaging, and subscriptions. It supports features like
message broadcasting, client management, and publishing messages to subscribed topics.

## Features

- WebSocket connection management.
- Sharding support for scalable client management.
- Publish/Subscribe system for topic-based messaging.
- Customizable message handlers (text and binary).
- Graceful error handling and client disconnection.

## Installation

```bash
go get -u github.com/lishank0119/fibril
```

## Usage

### Creating a WebSocket Hub

```go
package main

import (
	"fmt"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/lishank0119/fibril"
	"log"
)

func main() {
	// Create a new Fiber app
	app := fiber.New()

	// WebSocket endpoint
	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	// Initialize Fibril with options
	f := fibril.New(
		fibril.WithShardCount(20),
		fibril.WithMaxMessageSize(1024),
	)

	// Register ConnectHandler
	f.ConnectHandler(func(client *fibril.Client) {
		log.Println("Client connected:", client.UUID)
		if err := client.WriteText("Welcome!"); err != nil {
			log.Println(err)
			return
		}

		// Broadcast welcome message to other clients
		f.BroadcastText(fmt.Sprintf("Welcome! UUID: %s", client.UUID))
	})

	// WebSocket server that handles connections
	app.Get("/ws/:id", websocket.New(func(c *websocket.Conn) {
		f.RegisterClientWithKeys(c, map[any]any{"id": c.Params("id")})
	}))

	// Start the server
	log.Fatal(app.Listen(":3000"))
}
```

### Pub/Sub Example

```go
client.Subscribe("server-time", func (msg []byte) {
err := client.WriteText(string(msg))
if err != nil {
log.Println("Error sending server time:", err, "UUID:", client.UUID)
return
}
})

func publishServerTime(f *fibril.Fibril) {
IntervalTime := 1 * time.Second
ticker := time.NewTicker(IntervalTime)
for {
select {
case <-ticker.C:
err := f.Publish("server-time", []byte(time.Now().Format(time.RFC3339)))
if err != nil {
log.Fatal(err)
return
}
}
}
}
```

### Available Handlers

- **ConnectHandler**: Handles client connection.
- **DisconnectHandler**: Handles client disconnection.
- **ErrorHandler**: Handles errors that occur within a client connection.
- **TextMessageHandler**: Handles text messages.
- **BinaryMessageHandler**: Handles binary messages.
- **PongHandler**: Handles pong responses from clients.

## Configuration Options

You can customize the following options when initializing `Fibril`:

- **ShardCount**: The number of shards for managing clients (default: 16).
- **MaxMessageSize**: The maximum size of incoming messages in bytes (default: 512).
- **MessageBufferSize**: The size of the message buffer (default: 256).
- **WriteWait**: The duration to wait before closing the write connection (default: 10 seconds).
- **PongWait**: The duration to wait for a pong response from the client (default: 60 seconds).
- **PingPeriod**: The interval to send ping messages (default: 54 seconds).

Example:

```go
f := fibril.New(
fibril.WithShardCount(20),
fibril.WithMaxMessageSize(1024),
)
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contributions

Feel free to contribute to the project by forking it, making improvements, or submitting bug fixes via pull requests.


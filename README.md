[![Go Reference](https://pkg.go.dev/badge/github.com/lishank0119/fibril.svg)](https://pkg.go.dev/github.com/lishank0119/fibril)
[![go.mod](https://img.shields.io/github/go-mod/go-version/lishank0119/fibril)](go.mod)
[![Go Report Card](https://goreportcard.com/badge/github.com/lishank0119/fibril)](https://goreportcard.com/report/github.com/lishank0119/fibril)

# Fibril - A WebSocket Library for Go

[中文](README.zh-TW.md)

Fibril is a Go-based WebSocket library built on top of [GoFiber](https://github.com/gofiber/fiber) that provides a
robust and efficient way to handle WebSocket connections, messaging, and subscriptions. It supports features like
message broadcasting, client management, and publishing messages to subscribed topics.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
    - [Basic Example](#basic-example)
    - [Fibril Function Examples](#fibril-function-examples)
    - [Client Function Example](#client-function-example)
- [Configuration Options](#configuration-options)
- [Contributions](#contributions)
- [License](#license)

## Features

- 🚀 High-performance WebSocket server
- 🔗 Built-in Pub/Sub system
- ⚡ Sharding support for better scalability
- 🔒 Client management with UUIDs
- 🗂️ Custom key-value storage per client

## Installation

```bash
go get -u github.com/lishank0119/fibril
```

## Usage

### Basic Example

```go
package main

import (
  "github.com/gofiber/contrib/websocket"
  "github.com/gofiber/fiber/v2"
  "github.com/lishank0119/fibril"
  "log"
)

func main() {
  app := fiber.New()

  app.Use("/ws", func(c *fiber.Ctx) error {
    if websocket.IsWebSocketUpgrade(c) {
      return c.Next()
    }
    return fiber.ErrUpgradeRequired
  })

  f := fibril.New(
    fibril.WithShardCount(4),
    fibril.WithMaxMessageSize(1024),
  )

  f.TextMessageHandler(func(client *fibril.Client, msg string) {
    log.Printf("Received message from %s: %s", client.GetUUID(), msg)
    f.BroadcastText("Echo: " + msg)
  })

  app.Get("/ws", websocket.New(func(c *websocket.Conn) {
    f.RegisterClient(c)
  }))

  log.Fatal(app.Listen(":3000"))
}
```

### Fibril Function Examples

#### Publish

Publishes a message to a specific topic.

```go
err := f.Publish("server-time", []byte("2024-02-09T15:04:05Z"))
if err != nil {
	log.Println("Publish error:", err)
}
```

#### SendTextToClient

Sends a text message to a specific client.

```go
err := f.SendTextToClient("client-uuid", "Hello, Client!")
if err != nil {
	log.Println("Send error:", err)
}
```

#### SendBinaryToClient

Sends binary data to a specific client.

```go
err := f.SendBinaryToClient("client-uuid", []byte{0x01, 0x02, 0x03})
if err != nil {
	log.Println("Send error:", err)
}
```

#### BroadcastText

Broadcasts a text message to all connected clients.

```go
f.BroadcastText("Hello, everyone!")
```

#### BroadcastBinary

Broadcasts binary data to all connected clients.

```go
f.BroadcastBinary([]byte{0x10, 0x20, 0x30})
```

#### RegisterClient

Registers a new WebSocket client.

```go
app.Get("/ws", websocket.New(func(c *websocket.Conn) {
	f.RegisterClient(c)
}))
```

#### RegisterClientWithKeys

Registers a new WebSocket client with custom key-value pairs.

```go
app.Get("/ws/:id", websocket.New(func(c *websocket.Conn) {
	f.RegisterClientWithKeys(c, map[any]any{"id": c.Params("id")})
}))
```

#### DisconnectClient

Disconnects a specific client by UUID.

```go
err := f.DisconnectClient("Goodbye!", "client-uuid")
if err != nil {
	log.Println("Disconnect error:", err)
}
```

#### DisconnectClientFilter

Disconnects clients based on a filter function.

```go
f.DisconnectClientFilter("Maintenance", func(c *fibril.Client) bool {
	return c.GetKey("role") == "guest"
})
```

#### Available Handlers

- **ConnectHandler**: Handles client connection.

```go
f.ConnectHandler(func(client *fibril.Client) {
	log.Println("Client connected:", client.UUID)
})
```

- **DisconnectHandler**: Handles client disconnection.

```go
f.DisconnectHandler(func(client *fibril.Client) {
	log.Println("Client disconnected:", client.UUID)
})
```

- **ErrorHandler**: Handles errors that occur within a client connection.

```go
f.ErrorHandler(func(client *fibril.Client, err error) {
	log.Println("Error for client", client.UUID, ":", err)
})
```

- **TextMessageHandler**: Handles text messages.

```go
f.TextMessageHandler(func(client *fibril.Client, msg string) {
	log.Println("Received text message:", msg)
})
```

- **BinaryMessageHandler**: Handles binary messages.

```go
f.BinaryMessageHandler(func(client *fibril.Client, msg []byte) {
	log.Println("Received binary message:", msg)
})
```

- **PongHandler**: Handles pong responses from clients.

```go
f.PongHandler(func(client *fibril.Client) {
	log.Println("Pong received from:", client.UUID)
})
```

### Client Function Example

#### GetUUID
Gets the unique identifier (UUID) of the client.

```go
uuid := client.GetUUID()
log.Printf("Client UUID: %s", uuid)
```

#### Subscribe

Subscribes the client to a specific topic with a handler function.

```go
client.Subscribe("topic-name", func(msg []byte) {
	log.Printf("Received message for topic: %s", string(msg))
})
```

#### SendText

Sends a text message to the client.

```go
err := client.SendText("Hello, Client!")
if err != nil {
	log.Println("Send error:", err)
}
```

#### SendBinary

Sends binary data to the client.

```go
err := client.SendBinary([]byte{0x01, 0x02, 0x03})
if err != nil {
	log.Println("Send error:", err)
}
```

#### Disconnect

Disconnects the client with a custom message.

```go
client.Disconnect("Goodbye!")
```

#### StoreKey

Stores a custom key-value pair associated with the client.

```go
client.StoreKey("role", "admin")
```

#### DeleteKey

Deletes a custom key-value pair associated with the client.

```go
client.DeleteKey("role")
```

#### GetKey

Retrieves the value of a key associated with the client.

```go
role, ok := client.GetKey("role")
if ok {
	log.Printf("Client role: %v", role)
}
```

## Configuration Options

You can customize the following options when initializing `Fibril`:

- **ShardCount**: The number of shards for managing clients (default: 16).
- **MaxMessageSize**: The maximum size of incoming messages in bytes (default: 512).
- **MessageBufferSize**: The size of the message buffer (default: 256).
- **WriteWait**: The duration to wait before closing the write connection (default: 10 seconds).
- **PongWait**: The duration to wait for a pong response from the client (default: 60 seconds).
- **PingPeriod**: The interval to send ping messages (default: 54 seconds).

### Example:

```go
f := fibril.New(
    fibril.WithShardCount(20),
    fibril.WithMaxMessageSize(1024),
    fibril.WithMessageBufferSize(512),
    fibril.WithWriteWait(15 * time.Second),
    fibril.WithPongWait(30 * time.Second),
    fibril.WithPingPeriod(25 * time.Second),
)
```

## Contributions

Feel free to contribute to the project by forking it, making improvements, or submitting bug fixes via pull requests.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.


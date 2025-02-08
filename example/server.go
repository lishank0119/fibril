package main

import (
	"fmt"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/lishank0119/fibril"
	"log"
	"time"
)

func main() {
	app := fiber.New()

	// Middleware to upgrade connection to WebSocket
	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	// Create a new Fibril instance with custom options
	f := fibril.New(
		fibril.WithShardCount(20),       // Set the number of shards
		fibril.WithMaxMessageSize(1024), // Set the maximum message size in bytes
	)

	// Start a goroutine to periodically publish server time
	go PublishServerTime(f)

	// Set up the WebSocket connection handler
	f.ConnectHandler(func(client *fibril.Client) {
		// Log client connection and send a welcome message
		log.Println("Client connected, UUID:", client.UUID)
		if err := client.WriteText("Welcome!"); err != nil {
			log.Println("Error sending welcome message:", err)
			return
		}

		// Broadcast welcome message to all clients except the one connecting
		f.BroadcastTextFilter(fmt.Sprintf("Welcome! UUID: %s", client.UUID), func(c *fibril.Client) bool {
			return c.UUID != client.UUID
		})

		// Send a personalized message if client has an "id" key
		id, ok := client.GetKey("id")
		if ok {
			if err := f.SendTextToClient(client.UUID, fmt.Sprintf("Hello %v", id)); err != nil {
				log.Println("Error sending personalized message:", err)
				return
			}
		}

		// Disconnect clients with the same "id"
		f.DisconnectClientFilter("Duplicate ID detected", func(c *fibril.Client) bool {
			if fID, ok := c.GetKey("id"); ok {
				return c.UUID != client.UUID && id == fID
			}
			return false
		})

		// Subscribe the client to the "server-time" topic and send server time on update
		client.Subscribe("server-time", func(msg []byte) {
			err := client.WriteText(string(msg))
			if err != nil {
				log.Println("Error sending server time:", err, "UUID:", client.UUID)
				return
			}
		})
	})

	// Handle incoming text messages from clients
	f.TextMessageHandler(func(client *fibril.Client, msg string) {
		log.Println("Received message from UUID:", client.UUID, "Message:", msg)
		f.BroadcastText(msg) // Broadcast the received message to all clients
	})

	// Handle client disconnections
	f.DisconnectHandler(func(client *fibril.Client) {
		log.Println("Client disconnected, UUID:", client.UUID)
	})

	// Handle errors
	f.ErrorHandler(func(client *fibril.Client, err error) {
		log.Fatalf("Error from client UUID:%s, Error: %v \n", client.UUID, err)
	})

	// Set up the WebSocket endpoint with an "id" parameter
	app.Get("/ws/:id", websocket.New(func(c *websocket.Conn) {
		f.RegisterClientWithKeys(c, map[any]any{"id": c.Params("id")}) // Register client with an "id" key
	}))

	log.Fatal(app.Listen(":3000")) // Start the Fiber server on port 3000
}

// PublishServerTime sends the current server time to the "server-time" topic every second
func PublishServerTime(f *fibril.Fibril) {
	IntervalTime := 1 * time.Second
	ticker := time.NewTicker(IntervalTime)
	for {
		select {
		case <-ticker.C:
			// Publish the current server time
			err := f.Publish("server-time", []byte(time.Now().Format(time.RFC3339)))
			if err != nil {
				log.Fatal("Error publishing server time:", err)
				return
			}
		}
	}
}

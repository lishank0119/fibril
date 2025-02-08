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

	app.Use("/ws", func(c *fiber.Ctx) error {
		// IsWebSocketUpgrade returns true if the client
		// requested upgrade to the WebSocket protocol.
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	f := fibril.New()

	go PublishServerTime(f)

	f.ConnectHandler(func(client *fibril.Client) {
		log.Println("connect uuid:", client.UUID)
		if err := client.WriteText("Welcome!"); err != nil {
			log.Println(err)
			return
		}

		f.BroadcastTextFilter(fmt.Sprintf("Welcome! uuid:%s", client.UUID), func(c *fibril.Client) bool {
			return c.UUID != client.UUID
		})

		id, ok := client.GetKey("id")
		if ok {
			if err := f.SendTextToClient(client.UUID, fmt.Sprintf("Hello %v", id)); err != nil {
				return
			}
		}

		// kick same id
		f.DisconnectClientFilter("same id", func(c *fibril.Client) bool {
			if fID, ok := c.GetKey("id"); ok {
				return c.UUID != client.UUID && id == fID
			}

			return false
		})

		client.Subscribe("server-time", func(msg []byte) {
			err := client.WriteText(string(msg))
			if err != nil {
				log.Println(err)
				return
			}
		})
	})

	f.TextMessageHandler(func(client *fibril.Client, msg string) {
		log.Println("get msg from:", client.UUID, msg)
		f.BroadcastText(msg)
	})

	f.DisconnectHandler(func(client *fibril.Client) {
		log.Println("disconnect uuid:", client.UUID)
	})

	f.ErrorHandler(func(client *fibril.Client, err error) {
		log.Printf("uuid:%s error:%v", client.UUID, err)
	})

	app.Get("/ws/:id", websocket.New(func(c *websocket.Conn) {
		f.RegisterClientWithKeys(c, map[any]any{"id": c.Params("id")})
	}))

	log.Fatal(app.Listen(":3000"))
}

func PublishServerTime(f *fibril.Fibril) {
	IntervalTime := 1 * time.Second
	ticker := time.NewTicker(IntervalTime)
	for {
		select {
		case <-ticker.C:
			f.Publish("server-time", []byte(time.Now().Format(time.RFC3339)))
		}
	}
}

package main

import (
	"fmt"
	"github.com/lishank0119/fibril"
	"log"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
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

	f.ConnectHandler(func(client *fibril.Client) {
		log.Println("connect uuid:", client.UUID)
		if err := client.WriteText("Welcome!"); err != nil {
			log.Println(err)
			return
		}

		f.BroadcastTextFilter(fmt.Sprintf("Welcome! uuid:%s", client.UUID), func(c *fibril.Client) bool {
			return c.UUID != client.UUID
		})

		time.AfterFunc(time.Millisecond*100, func() {
			id, ok := client.GetKey("id")
			if ok {
				if err := f.SendTextToClient(client.UUID, fmt.Sprintf("Hello %v", id)); err != nil {
					return
				}
			}

			// kick same id
			f.DisconnectClientFilter("same id", func(c *fibril.Client) bool {
				if fID, ok := client.GetKey("id"); ok {
					return c.UUID != client.UUID && id == fID
				}

				return false
			})
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

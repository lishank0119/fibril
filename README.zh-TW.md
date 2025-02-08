# Fibril - Go 語言 WebSocket 函式庫

[English](README.md)

Fibril 是一個基於 [GoFiber](https://github.com/gofiber/fiber) 的 Go 語言 WebSocket 函式庫，提供了一個穩健且高效的方式來處理
WebSocket 連接、訊息傳遞和訂閱功能。它支援訊息廣播、客戶端管理以及將訊息發佈至訂閱的主題。

## 特點

- WebSocket 連接管理。
- 支援分片（Sharding）來實現可擴展的客戶端管理。
- 支援發佈/訂閱（Pub/Sub）系統，用於基於主題的訊息傳遞。
- 可自定義訊息處理程序（文本和二進位）。
- 優雅的錯誤處理和客戶端斷線處理。

## 安裝

```bash
go get -u github.com/lishank0119/fibril
```

## 使用範例

### 建立 WebSocket Hub

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
	// 建立一個新的 Fiber 應用
	app := fiber.New()

	// WebSocket 端點
	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	// 初始化 Fibril 並設置選項
	f := fibril.New(
		fibril.WithShardCount(20),
		fibril.WithMaxMessageSize(1024),
	)

	// 註冊 ConnectHandler
	f.ConnectHandler(func(client *fibril.Client) {
		log.Println("Client 連接:", client.UUID)
		if err := client.WriteText("歡迎！"); err != nil {
			log.Println(err)
			return
		}

		// 廣播歡迎訊息給其他客戶端
		f.BroadcastText(fmt.Sprintf("歡迎！UUID: %s", client.UUID))
	})

	// WebSocket 伺服器，處理連接
	app.Get("/ws/:id", websocket.New(func(c *websocket.Conn) {
		f.RegisterClientWithKeys(c, map[any]any{"id": c.Params("id")})
	}))

	// 啟動伺服器
	log.Fatal(app.Listen(":3000"))
}
```

### Pub/Sub 範例

```go
client.Subscribe("server-time", func (msg []byte) {
err := client.WriteText(string(msg))
if err != nil {
log.Println("Error sending server time:", err, "UUID:", client.UUID)
return
}
})

func PublishServerTime(f *fibril.Fibril) {
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

### 可用的處理程序

- **ConnectHandler**: 處理客戶端連接。
- **DisconnectHandler**: 處理客戶端斷線。
- **ErrorHandler**: 處理發生錯誤的客戶端連接。
- **TextMessageHandler**: 處理文本訊息。
- **BinaryMessageHandler**: 處理二進位訊息。
- **PongHandler**: 處理來自客戶端的 Pong 回應。

## 配置選項

您可以在初始化 `Fibril` 時自定義以下選項：

- **ShardCount**: 管理客戶端的分片數量（預設：16）。
- **MaxMessageSize**: 每個訊息的最大大小（預設：512 bytes）。
- **MessageBufferSize**: 訊息緩衝區的大小（預設：256）。
- **WriteWait**: 等待關閉寫入連接的時間（預設：10 秒）。
- **PongWait**: 等待來自客戶端的 Pong 回應時間（預設：60 秒）。
- **PingPeriod**: 發送 Ping 訊息的間隔（預設：54 秒）。

範例：

```go
f := fibril.New(
fibril.WithShardCount(20),
fibril.WithMaxMessageSize(1024),
)
```

## 授權許可

此專案採用 MIT 授權條款 - 詳情請參閱 [LICENSE](LICENSE) 檔案。

## 貢獻

如果您有興趣貢獻這個專案，可以透過 Fork 專案、改善功能或提交錯誤修正來參與貢獻。

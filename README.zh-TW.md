# Fibril - Go 語言 WebSocket 函式庫

[English](README.zh-TW.md)

Fibril 是一個基於 [GoFiber](https://github.com/gofiber/fiber) 的 Go 語言 WebSocket 函式庫，提供了一個穩健且高效的方式來處理
WebSocket 連接、訊息傳遞和訂閱功能。它支援訊息廣播、客戶端管理以及將訊息發佈至訂閱的主題。

## 目錄

- [功能](#功能)
- [安裝](#安裝)
- [使用方式](#使用方式)
    - [基本範例](#基本範例)
    - [Fibril 函式範例](#fibril-函式範例)
    - [Client 函式範例](#client-函式範例)
- [配置選項](#配置選項)
- [貢獻](#貢獻)
- [授權](#授權)

## 功能

- 🚀 高效能的 WebSocket 伺服器
- 🔗 內建 Pub/Sub 系統
- ⚡ 支援分片以提高擴展性
- 🔒 使用 UUID 進行客戶端管理
- 🗂️ 為每個客戶端提供自訂的鍵值存儲

## 安裝

```bash
go get -u github.com/lishank0119/fibril
```

## 使用方式

### 基本範例

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
    log.Printf("收到來自 %s 的訊息: %s", client.GetUUID(), msg)
    f.BroadcastText("回音: " + msg)
  })

  app.Get("/ws", websocket.New(func(c *websocket.Conn) {
    f.RegisterClient(c)
  }))

  log.Fatal(app.Listen(":3000"))
}
```

### Fibril 函式範例

### ClientLen

取得連線客戶端的數量

```go
f.ClientLen()
```

#### Publish

發佈一條訊息到指定的主題。

```go
err := f.Publish("server-time", []byte("2024-02-09T15:04:05Z"))
if err != nil {
	log.Println("發佈錯誤:", err)
}
```

#### SendTextToClient

發送文字訊息到指定的客戶端。

```go
err := f.SendTextToClient("client-uuid", "你好，客戶端！")
if err != nil {
	log.Println("發送錯誤:", err)
}
```

#### SendBinaryToClient

發送二進位資料到指定的客戶端。

```go
err := f.SendBinaryToClient("client-uuid", []byte{0x01, 0x02, 0x03})
if err != nil {
	log.Println("發送錯誤:", err)
}
```

#### BroadcastText

向所有連接的客戶端廣播文字訊息。

```go
f.BroadcastText("大家好！")
```

#### BroadcastBinary

向所有連接的客戶端廣播二進位資料。

```go
f.BroadcastBinary([]byte{0x10, 0x20, 0x30})
```

#### RegisterClient

註冊一個新的 WebSocket 客戶端。

```go
app.Get("/ws", websocket.New(func(c *websocket.Conn) {
	f.RegisterClient(c)
}))
```

#### RegisterClientWithKeys

註冊一個帶有自訂鍵值對的新 WebSocket 客戶端。

```go
app.Get("/ws/:id", websocket.New(func(c *websocket.Conn) {
	f.RegisterClientWithKeys(c, map[any]any{"id": c.Params("id")})
}))
```

#### DisconnectClient

根據 UUID 斷開指定的客戶端。

```go
err := f.DisconnectClient("再見!", "client-uuid")
if err != nil {
	log.Println("斷開連接錯誤:", err)
}
```

#### DisconnectClientFilter

根據過濾條件斷開客戶端連接。

```go
f.DisconnectClientFilter("維護中", func(c *fibril.Client) bool {
	return c.GetKey("role") == "guest"
})
```

#### Handler

- **ConnectHandler**: 處理客戶端連接。

```go
f.ConnectHandler(func(client *fibril.Client) {
	log.Println("客戶端已連接:", client.GetUUID())
})
```

- **DisconnectHandler**: 處理客戶端斷開連接。

```go
f.DisconnectHandler(func(client *fibril.Client) {
	log.Println("客戶端已斷開:", client.GetUUID())
})
```

- **ErrorHandler**: 處理客戶端連接中發生的錯誤。

```go
f.ErrorHandler(func(client *fibril.Client, err error) {
	log.Println("客戶端", client.GetUUID(), "錯誤:", err)
})
```

- **TextMessageHandler**: 處理文字訊息。

```go
f.TextMessageHandler(func(client *fibril.Client, msg string) {
	log.Println("收到文字訊息:", msg)
})
```

- **BinaryMessageHandler**: 處理二進位訊息。

```go
f.BinaryMessageHandler(func(client *fibril.Client, msg []byte) {
	log.Println("收到二進位訊息:", msg)
})
```

- **PongHandler**: 處理來自客戶端的 Pong 回應。

```go
f.PongHandler(func(client *fibril.Client) {
	log.Println("收到來自:", client.GetUUID(), "的 Pong")
})
```

### Client 函式範例

#### GetUUID
取得客戶端的唯一編號 (UUID)。

```go
uuid := client.GetUUID()
log.Printf("Client UUID: %s", uuid)
```

#### Subscribe

訂閱客戶端到指定的主題並設置處理函式。

```go
client.Subscribe("topic-name", func(msg []byte) {
	log.Printf("收到主題的訊息: %s", string(msg))
})
```

#### SendText

發送文字訊息給客戶端。

```go
err := client.SendText("你好，客戶端！")
if err != nil {
	log.Println("發送錯誤:", err)
}
```

#### SendBinary

發送二進位資料給客戶端。

```go
err := client.SendBinary([]byte{0x01, 0x02, 0x03})
if err != nil {
	log.Println("發送錯誤:", err)
}
```

#### Disconnect

根據自訂訊息斷開客戶端。

```go
client.Disconnect("再見！")
```

#### StoreKey

為客戶端存儲自訂鍵值對。

```go
client.StoreKey("role", "admin")
```

#### DeleteKey

刪除客戶端的自訂鍵值對。

```go
client.DeleteKey("role")
```

#### GetKey

獲取客戶端的鍵值對。

```go
role, ok := client.GetKey("role")
if ok {
	log.Printf("客戶端角色: %v", role)
}
```

## 配置選項

在初始化 `Fibril` 時，您可以自訂以下選項：

- **ShardCount**: 用於管理客戶端的分片數量（預設：16）。
- **MaxMessageSize**: 最大接收訊息大小（預設：512 字節）。
- **MessageBufferSize**: 訊息緩衝區大小（預設：256）。
- **WriteWait**: 關閉寫入連接前的等待時間（預設：10 秒）。
- **PongWait**: 等待客戶端 Pong 回應的時間（預設：60 秒）。
- **PingPeriod**: 發送 Ping 訊息的間隔（預設：54 秒）。

### 範例：

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

## 貢獻

歡迎 fork 項目、改進功能或提交 bug 修復來貢獻。

## 授權

本專案採用 MIT 授權，詳細內容請參見 [LICENSE](LICENSE) 文件。

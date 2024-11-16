package chats

import (
	"fmt"

	"github.com/gorilla/websocket"
)

// 客户端对象
type Client struct {
	Chat *Chat
	Conn *websocket.Conn
}

// 读取数据通道
func (c *Client) GetData() {
	defer func() {
		// 确保客户端被注销
		c.Chat.Unregister <- c
		// 确保 WebSocket 连接被关闭
		c.Conn.Close()
	}()

	for {
		// 读取消息
		_, message, err := c.Conn.ReadMessage()

		// 如果读取失败，则退出
		if err != nil {
			fmt.Println("Error reading message:", err)
			break
		}

		// 处理消息并广播
		c.Chat.Broadcast <- message
	}
}

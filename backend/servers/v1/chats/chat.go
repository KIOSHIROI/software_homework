package chats

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
)

// 存储聊天室所有在线用户
var (
	historyClientNum = 0
)

// 将消息广播各个用户
func broadcastMsg(chat *Chat, mt int, message []byte) {
	for client := range chat.Clients {
		client.Conn.WriteMessage(mt, message)
	}
}

type connectedDataS struct {
	Event            string `json:"_event"`
	HistoryClientNum int    `json:"historyClientNum"`
	OnlineClientNum  int    `json:"onlineClientNum"`
}

// 消息结构体
type messageDataS struct {
	Event   string `json:"_event"`
	Message string `json:"message"`
}

// 定义聊天室对象
type Chat struct {
	Clients    map[*Client]bool
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
}

// 定义聊天室的数据广播
func (c *Chat) Run() {
	for {
		select {
		case client := <-c.Register:
			fmt.Printf("进入聊天室\n")
			c.Clients[client] = true
			// 更新历史客户端数
			historyClientNum++
			// 进入聊天室
			var data = connectedDataS{
				Event:            "connected",
				HistoryClientNum: historyClientNum,
				OnlineClientNum:  len(c.Clients),
			}
			var dataJson, _ = json.Marshal(data)
			broadcastMsg(c, websocket.TextMessage, dataJson)
		case client := <-c.Unregister:
			fmt.Printf("退出聊天室\n")
			if _, ok := c.Clients[client]; ok {
				delete(c.Clients, client)
				// 广播离开
				var data = connectedDataS{
					Event:            "connected",
					HistoryClientNum: historyClientNum,
					OnlineClientNum:  len(c.Clients),
				}
				var dataJson, _ = json.Marshal(data)
				broadcastMsg(c, websocket.TextMessage, dataJson)
			}
		case message := <-c.Broadcast:
			var data = messageDataS{
				Event:   "message",
				Message: string(message),
			}
			dataJson, _ := json.Marshal(data)
			broadcastMsg(c, websocket.TextMessage, dataJson)
		}
	}
}

// 实例化聊天室对象
func NewChat() *Chat {
	return &Chat{
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
	}
}

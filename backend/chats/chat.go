package chats

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// 升级器
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// 用户连接映射
var connections = struct {
	sync.RWMutex
	mapping map[string]*websocket.Conn
}{
	mapping: make(map[string]*websocket.Conn),
}

// 定义消息结构
type Message struct {
	Sender  string `json:"sender"`
	Target  string `json:"target"`
	Content string `json:"content"`
}

func HandleConnections(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "缺少 user_id参数", http.StatusBadRequest)
		return
	}

	// 升级 HTTP -> Websocket
	ws, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Printf("Websocket 升级失败: %v", err)
		return
	}
	defer ws.Close()

	// 将用户加入映射
	connections.Lock()
	connections.mapping[userID] = ws
	connections.Unlock()

	log.Printf("用户 %s 已连接", userID)

	// 监听用户信息

	for {
		var msg Message
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("读取用户 %s 信息失败: %v", userID, err)
			break
		}
		connections.RLock()
		targetConn, ok := connections.mapping[msg.Target]
		connections.RUnlock()

		if !ok {
			log.Printf("目标用户 %s 不在线", msg.Target)
			err := ws.WriteJSON(Message{
				Sender:  "系统",
				Target:  userID,
				Content: fmt.Sprintf("用户 %s 不在线", msg.Target),
			})
			if err != nil {
				log.Printf("发送系统消息失败: %v", err)
			}
			continue
		}

		// 见消息转发给目标用户
		err = targetConn.WriteJSON(msg)
		if err != nil {
			log.Printf("转发消息失败: %v", err)
			targetConn.Close()
			connections.Lock()
			delete(connections.mapping, msg.Target)
			connections.Unlock()
		}
	}
	connections.Lock()
	delete(connections.mapping, userID)
	connections.Unlock()
	log.Printf("用户 %s 已断开链接", userID)
}

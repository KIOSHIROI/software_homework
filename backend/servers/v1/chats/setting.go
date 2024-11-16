package chats

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upGrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for now
		return true
	},
}

func CreateWs(c *gin.Context, chat *Chat) {
	conn, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	client := &Client{Chat: chat, Conn: conn}
	go client.GetData()
	client.Chat.Register <- client
}

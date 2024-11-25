package v1

import (
	"backend/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func SaveMessageHandler(c *gin.Context) {
	var payload struct {
		SenderID   uint   `json:"sender_id" binding:"required"`
		ReceiverID uint   `json:"receiver_id" binding:"required"`
		Content    string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := models.SaveMessage(payload.SenderID, payload.ReceiverID, payload.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Message saved successfully"})
}

func GetChatHistoryHandler(c *gin.Context) {
	senderID, err := strconv.Atoi(c.Query("sender_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sender ID"})
		return
	}

	receiverID, err := strconv.Atoi(c.Query("receiver_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid receiver ID"})
		return
	}

	messages, err := models.GetChatHistory(uint(senderID), uint(receiverID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch messages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"messages": messages})
}
func MarkMessagesAsReadHandler(c *gin.Context) {
	var payload struct {
		SenderID   uint `json:"sender_id" binding:"required"`
		ReceiverID uint `json:"receiver_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := models.MarkMessagesAsRead(payload.SenderID, payload.ReceiverID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update message status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Messages marked as read"})
}

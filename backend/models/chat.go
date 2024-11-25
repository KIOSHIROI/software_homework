package models

import (
	"time"

	"gorm.io/gorm"
)

type Messages struct {
	gorm.Model
	SenderID   uint      `gorm:"not null"`
	ReceiverID uint      `gorm:"not null"`
	Content    string    `gorm:"type:text;not null"`
	IsRead     bool      `gorm:"default:false"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
}

func SaveMessage(senderID, receiverID uint, content string) error {
	message := Messages{
		SenderID:   senderID,
		ReceiverID: receiverID,
		Content:    content,
	}
	return DB.Create(&message).Error
}

func GetChatHistory(senderID, receiverID uint) ([]Messages, error) {
	var messages []Messages
	err := DB.Where(
		"(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
		senderID, receiverID, receiverID, senderID,
	).Order("created_at ASC").Find(&messages).Error
	return messages, err
}

func MarkMessagesAsRead(senderID, receiverID uint) error {
	return DB.Model(&Messages{}).Where(
		"sender_id = ? AND receiver_id = ? AND is_read = ?", senderID, receiverID, false,
	).Update("is_read", true).Error
}

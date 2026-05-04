package models

import "time"

type Conversation struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ClientID  uint      `gorm:"not null;unique;index" json:"client_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Messages []Message `gorm:"foreignKey:ConversationID" json:"messages,omitempty"`
}

package models

import "time"

type Conversation struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	ClientID   uint      `gorm:"not null;index" json:"client_id"`
	BusinessID uint      `gorm:"not null;index" json:"business_id"` // Fixed typo
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// Add unique constraint for client+business pair
	_ struct{} `gorm:"uniqueIndex:idx_client_business,priority:1,columns:client_id,business_id"`

	Messages []Message `gorm:"foreignKey:ConversationID" json:"messages,omitempty"`
}

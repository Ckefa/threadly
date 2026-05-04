package models

import "time"

type ClientStatus string

const (
	StatusNew             ClientStatus = "new"
	StatusActive          ClientStatus = "active"
	StatusAwaitingPayment ClientStatus = "awaiting_payment"
	StatusCompleted       ClientStatus = "completed"
)

type Client struct {
	ID             uint         `gorm:"primaryKey" json:"id"`
	UserID         uint         `gorm:"not null;index" json:"user_id"`
	Name           string       `gorm:"not null" json:"name"`
	Email          string       `json:"email"`
	Phone          string       `json:"phone"`
	Status         ClientStatus `gorm:"default:'new'" json:"status"`
	IsOnline       bool         `gorm:"default:false" json:"is_online"`
	LastSeenAt     *time.Time   `json:"last_seen_at"`
	ConversationID uint         `json:"conversation_id"` // For template access
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`

	User         User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Conversation Conversation `gorm:"foreignKey:ClientID" json:"conversation,omitempty"`
}

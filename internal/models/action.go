package models

import "time"

type ActionType string

const (
	ActionTask     ActionType = "task"
	ActionReminder ActionType = "reminder"
	ActionBooking  ActionType = "booking"
)

type Action struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	MessageID     uint       `gorm:"not null;index" json:"message_id"`
	Type          ActionType `gorm:"not null" json:"type"`
	Title         string     `gorm:"not null" json:"title"`
	Description   string     `json:"description"`
	DueDate       *time.Time `json:"due_date"`
	Priority      string     `gorm:"default:'medium'" json:"priority"` // low, medium, high
	Status        string     `gorm:"default:'pending'" json:"status"`  // pending, in_progress, completed, cancelled
	AssignedTo    string     `json:"assigned_to"`                      // staff member name
	EstimatedCost float64    `json:"estimated_cost"`
	ActualCost    float64    `json:"actual_cost"`
	Notes         string     `json:"notes"`
	IsCompleted   bool       `gorm:"default:false" json:"is_completed"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`

	Message Message `gorm:"foreignKey:MessageID" json:"message,omitempty"`
}

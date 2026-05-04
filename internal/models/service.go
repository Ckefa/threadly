package models

import (
	"time"
)

type Service struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uint      `gorm:"not null;index" json:"user_id"`
	Name          string    `gorm:"not null" json:"name"`
	Description   string    `gorm:"type:text" json:"description"`
	MinPrice      float64   `gorm:"not null" json:"min_price"`
	MaxPrice      float64   `gorm:"not null" json:"max_price"`
	IsNegotiable  bool      `gorm:"default:false" json:"is_negotiable"`
	Duration      int       `gorm:"comment:duration in minutes" json:"duration"`
	ImageURL      string    `json:"image_url"`
	IsActive      bool      `gorm:"default:true" json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// Relationships
	User          User               `gorm:"foreignKey:UserID" json:"user,omitempty"`
	BookingItems  []BookingItem      `gorm:"foreignKey:ServiceID" json:"booking_items,omitempty"`
}

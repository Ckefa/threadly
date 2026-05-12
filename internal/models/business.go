package models

import "time"

type Business struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Email        string    `gorm:"unique;not null" json:"email"`
	Password     string    `gorm:"not null" json:"-"`
	Name         string    `json:"name"`
	Username     string    `json:"username"`
	BusinessType string    `json:"business_type"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	Clients []Client `gorm:"foreignKey:BusinessID" json:"clients,omitempty"`
}

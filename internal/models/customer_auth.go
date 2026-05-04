package models

import "time"

type CustomerAuth struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	ClientID     uint      `gorm:"not null;index" json:"client_id"`
	Email        string    `gorm:"not null" json:"email"`
	OTPCode      string    `gorm:"not null" json:"-"`
	OTPExpiresAt time.Time `gorm:"not null" json:"otp_expires_at"`
	IsVerified   bool      `gorm:"default:false" json:"is_verified"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	
	Client Client `gorm:"foreignKey:ClientID" json:"client,omitempty"`
}

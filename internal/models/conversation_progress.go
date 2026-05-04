package models

import "time"

type ConversationStage string

const (
	StageInitial       ConversationStage = "initial"        // First contact
	StageQualification ConversationStage = "qualification"  // Understanding needs
	StageNegotiation   ConversationStage = "negotiation"    // Discussing details/pricing
	StageConfirmation  ConversationStage = "confirmation"   // Booking confirmed
	StageInProgress    ConversationStage = "in_progress"    // Service being delivered
	StageCompleted     ConversationStage = "completed"      // Service completed
	StageFollowUp      ConversationStage = "follow_up"       // Post-service follow-up
)

type ConversationProgress struct {
	ID             uint              `gorm:"primaryKey" json:"id"`
	ConversationID uint              `gorm:"not null;uniqueIndex" json:"conversation_id"`
	CurrentStage   ConversationStage `gorm:"default:'initial'" json:"current_stage"`
	StageHistory   []StageTransition `gorm:"serializer:json" json:"stage_history"`
	ProgressScore  int               `gorm:"default:0" json:"progress_score"` // 0-100
	NextAction     string            `json:"next_action"`
	ExpectedClose  *time.Time        `json:"expected_close"`
	ActualClose    *time.Time        `json:"actual_close"`
	Value          float64           `json:"value"` // Potential deal value
	Notes          string            `json:"notes"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
	
	Conversation Conversation `gorm:"foreignKey:ConversationID" json:"conversation,omitempty"`
}

type StageTransition struct {
	Stage     ConversationStage `json:"stage"`
	ChangedAt time.Time         `json:"changed_at"`
	Reason    string            `json:"reason"`
	Duration  int               `json:"duration"` // Time spent in previous stage (hours)
}

package alert

import (
	"time"
)

type User struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"unique;not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

type Alert struct {
	ID          int       `json:"id" gorm:"primaryKey"`
	UserID      int       `json:"user_id"`
	Symbol      string    `json:"symbol" gorm:"not null"`
	TargetPrice float64   `json:"target_price" gorm:"not null"`
	Condition   string    `json:"condition" gorm:"not null"` // "ABOVE" or "BELOW"
	Triggered   bool      `json:"triggered" gorm:"default:false"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	User        User      `json:"-" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

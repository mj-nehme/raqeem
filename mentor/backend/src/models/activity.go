package models

import "time"

type Activity struct {
	ID        uint `gorm:"primaryKey"`
	UserID    string
	Location  string
	Filename  string
	Timestamp time.Time
}

package models

import "time"

type Movie struct {
	ID          uint
	Title       string `gorm:"size:100;unique"`
	Genre       string
	ReleasedAt  time.Time
	Description string
	Rating      float64 `gorm:"type:numeric(3,1)"`
}

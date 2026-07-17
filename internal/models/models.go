package models

import "time"

type Movie struct {
	ID           uint
	Title        string `gorm:"size:100;unique"`
	Genre        string
	ReleasedAt   time.Time
	Description  string
	Rating       float64 `gorm:"type:numeric(3,1)"`
	Reviews      []Review
	ReviewsCount int
	DirectorID   uint
	Director     Director
	Actors       []Actor `gorm:"many2many:movie_actors;"`
}

type Director struct {
	ID   uint
	Name string
}

type Actor struct {
	ID   uint
	Name string
}

type Review struct {
	ID      uint
	MovieID uint
	Score   int
	Text    string
}

// DTO
type MovieRatingDTO struct {
	Title        string
	Rating       float64
	ReviewsCount int
}

type MovieLeaderboardDTO struct {
	Title        string
	Rating       float64
	ReviewsCount int
	Rank         int
}

package models

import (
	"time"

	"gorm.io/gorm"
)

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

// Hooks
func (r *Review) AfterCreate(tx *gorm.DB) error {
	// Recalculate movie rating and review count from the reviews table
	return tx.Exec(`
		UPDATE movies
		SET
			reviews_count = (SELECT COUNT(*) FROM reviews WHERE movie_id = ?),
			rating = (SELECT AVG(score)::numeric(3,1) FROM reviews WHERE movie_id = ?)
		WHERE id = ?
	`, r.MovieID, r.MovieID, r.MovieID).Error
}

package models

import (
	"time"

	"gorm.io/gorm"
)

type Movie struct {
	ID           uint
	Title        string `gorm:"size:100;unique;not null"`
	Genre        string `gorm:"not null"`
	ReleasedAt   time.Time
	Description  string
	Rating       float64 `gorm:"type:numeric(3,1)"`
	Reviews      []Review
	ReviewsCount int
	DirectorID   *uint
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
	Score   int    `gorm:"check:score >= 1 AND score <= 10"`
	Text    string `gorm:"size:255;not null"`
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
	type Result struct {
		AvgScore float64
		Count    int64
	}

	var result Result
	if err := tx.Model(&Review{}).
		Select("COALESCE(AVG(score), 0) as avg_score, COUNT(*) as count").
		Where("movie_id = ?", r.MovieID).
		Scan(&result).Error; err != nil {
		return err
	}

	// Update the movie with recalculated values
	return tx.Model(&Movie{}).
		Where("id = ?", r.MovieID).
		Updates(map[string]any{
			"rating":        result.AvgScore,
			"reviews_count": result.Count,
		}).Error
}

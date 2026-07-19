package requests

import (
	"gorm.io/gorm"

	"github.com/EvgeniiIvanov/hexlet-gorm/internal/models"
)

// CreateReview creates a new review for a movie
func CreateReview(db *gorm.DB, movieID uint, score int, text string) error {
	review := models.Review{
		MovieID: movieID,
		Score:   score,
		Text:    text,
	}
	return db.Create(&review).Error
}

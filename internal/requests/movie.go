package requests

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/EvgeniiIvanov/hexlet-gorm/internal/models"
)

// ListMovies returns all movies with their directors
func ListMovies(db *gorm.DB) ([]models.Movie, error) {
	var movies []models.Movie
	if err := db.Preload("Director").Find(&movies).Error; err != nil {
		return nil, err
	}
	return movies, nil
}

// CreateMovie creates a new movie
func CreateMovie(db *gorm.DB, title, genre string, releasedAt time.Time) (*models.Movie, error) {
	movie := models.Movie{
		Title:      title,
		Genre:      genre,
		ReleasedAt: releasedAt,
	}
	if err := db.Create(&movie).Error; err != nil {
		return nil, err
	}
	return &movie, nil
}

// GetMovie retrieves a movie by ID with director and actors
func GetMovie(db *gorm.DB, id uint) (*models.Movie, error) {
	var movie models.Movie
	if err := db.Preload("Director").Preload("Actors").First(&movie, id).Error; err != nil {
		return nil, err
	}
	return &movie, nil
}

// UpdateMovie updates a specific field of a movie
func UpdateMovie(db *gorm.DB, id uint, field, value string) (int64, error) {
	result := db.Model(&models.Movie{}).
		Where("id = ?", id).
		Update(field, value)
	
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

// DeleteMovie deletes a movie by ID
func DeleteMovie(db *gorm.DB, id uint) (int64, error) {
	result := db.Delete(&models.Movie{}, id)
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

// GetUnratedMovies returns movies without ratings
func GetUnratedMovies(db *gorm.DB) ([]models.Movie, error) {
	var movies []models.Movie
	if err := db.Where("rating IS NULL").Find(&movies).Error; err != nil {
		return nil, err
	}
	return movies, nil
}

// GetMostRatedMovies returns movies with rating > 8.5
func GetMostRatedMovies(db *gorm.DB) ([]models.Movie, error) {
	var movies []models.Movie
	if err := db.Where("rating > ?", 8.5).Find(&movies).Error; err != nil {
		return nil, err
	}
	return movies, nil
}

// GetMoviesWithRatings returns all movies with their calculated ratings
func GetMoviesWithRatings(db *gorm.DB) ([]models.MovieRatingDTO, error) {
	var movies []models.MovieRatingDTO
	if err := db.Raw(`
		SELECT 
			m.title, 
			AVG(r.score) as rating, 
			COUNT(r.id) as reviews_count 
		FROM movies m
		LEFT JOIN reviews r ON r.movie_id = m.id
		GROUP BY m.id, m.title
		ORDER BY rating DESC NULLS LAST, m.title;
	`).Scan(&movies).Error; err != nil {
		return nil, err
	}
	return movies, nil
}

// GetMoviesLeaderboard returns ranked movies by average rating
func GetMoviesLeaderboard(db *gorm.DB) ([]models.MovieLeaderboardDTO, error) {
	var movies []models.MovieLeaderboardDTO
	if err := db.Raw(`
		SELECT
			m.title,
			COALESCE(AVG(r.score), 0) as rating,
			COUNT(r.id) as reviews_count,
			DENSE_RANK() OVER (ORDER BY COALESCE(AVG(r.score), 0) DESC) as rank
		FROM movies m
		LEFT JOIN reviews r ON r.movie_id = m.id
		GROUP BY m.id, m.title
		ORDER BY rating DESC, m.title;
	`).Scan(&movies).Error; err != nil {
		return nil, err
	}
	return movies, nil
}

// MovieExists checks if a movie exists by ID
func MovieExists(db *gorm.DB, id uint) (bool, error) {
	var movie models.Movie
	err := db.First(&movie, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

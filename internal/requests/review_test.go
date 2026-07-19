package requests_test

import (
	"testing"

	"gorm.io/gorm"

	"github.com/EvgeniiIvanov/hexlet-gorm/internal/models"
	"github.com/EvgeniiIvanov/hexlet-gorm/internal/requests"
	"github.com/EvgeniiIvanov/hexlet-gorm/internal/testutil"
)

func TestCreateReview(t *testing.T) {
	db := testutil.SetupTestDB(t)

	t.Run("creates review and updates movie rating", func(t *testing.T) {
		testutil.WithTx(t, db, func(tx *gorm.DB) {
			// Create a movie first
			movie := testutil.MovieFactory("Inception")
			if err := tx.Create(movie).Error; err != nil {
				t.Fatalf("failed to create movie: %v", err)
			}

			// Create a review
			err := requests.CreateReview(tx, movie.ID, 9, "Amazing movie!")
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			// Verify review was created
			var review models.Review
			if err := tx.Where("movie_id = ?", movie.ID).First(&review).Error; err != nil {
				t.Fatalf("failed to find review: %v", err)
			}

			if review.Score != 9 {
				t.Errorf("expected score 9, got %d", review.Score)
			}
			if review.Text != "Amazing movie!" {
				t.Errorf("expected text 'Amazing movie!', got %s", review.Text)
			}

			// Verify movie rating was updated (via AfterCreate hook)
			var updatedMovie models.Movie
			if err := tx.First(&updatedMovie, movie.ID).Error; err != nil {
				t.Fatalf("failed to find movie: %v", err)
			}

			if updatedMovie.Rating != 9.0 {
				t.Errorf("expected rating 9.0, got %.1f", updatedMovie.Rating)
			}
			if updatedMovie.ReviewsCount != 1 {
				t.Errorf("expected reviews_count 1, got %d", updatedMovie.ReviewsCount)
			}
		})
	})

	t.Run("creates multiple reviews and calculates average rating", func(t *testing.T) {
		testutil.WithTx(t, db, func(tx *gorm.DB) {
			// Create a movie
			movie := testutil.MovieFactory("The Matrix")
			if err := tx.Create(movie).Error; err != nil {
				t.Fatalf("failed to create movie: %v", err)
			}

			// Create multiple reviews
			if err := requests.CreateReview(tx, movie.ID, 10, "Perfect!"); err != nil {
				t.Fatalf("failed to create review 1: %v", err)
			}
			if err := requests.CreateReview(tx, movie.ID, 8, "Good"); err != nil {
				t.Fatalf("failed to create review 2: %v", err)
			}
			if err := requests.CreateReview(tx, movie.ID, 9, "Great"); err != nil {
				t.Fatalf("failed to create review 3: %v", err)
			}

			// Verify average rating: (10 + 8 + 9) / 3 = 9.0
			var updatedMovie models.Movie
			if err := tx.First(&updatedMovie, movie.ID).Error; err != nil {
				t.Fatalf("failed to find movie: %v", err)
			}

			expectedRating := 9.0
			if updatedMovie.Rating != expectedRating {
				t.Errorf("expected rating %.1f, got %.1f", expectedRating, updatedMovie.Rating)
			}
			if updatedMovie.ReviewsCount != 3 {
				t.Errorf("expected reviews_count 3, got %d", updatedMovie.ReviewsCount)
			}
		})
	})

	t.Run("prevents duplicate reviews with same text", func(t *testing.T) {
		testutil.WithTx(t, db, func(tx *gorm.DB) {
			// Create a movie
			movie := testutil.MovieFactory("Interstellar")
			if err := tx.Create(movie).Error; err != nil {
				t.Fatalf("failed to create movie: %v", err)
			}

			// Create first review
			if err := requests.CreateReview(tx, movie.ID, 10, "Mind-blowing!"); err != nil {
				t.Fatalf("failed to create first review: %v", err)
			}

			// Try to create duplicate review with same text
			err := requests.CreateReview(tx, movie.ID, 9, "Mind-blowing!")
			if err == nil {
				t.Error("expected error for duplicate review, got nil")
			}
		})
	})

	t.Run("fails for non-existent movie", func(t *testing.T) {
		testutil.WithTx(t, db, func(tx *gorm.DB) {
			// Try to create review for non-existent movie
			err := requests.CreateReview(tx, 99999, 8, "Review for nothing")
			if err == nil {
				t.Error("expected error for non-existent movie, got nil")
			}
		})
	})
}

func TestReviewRatingIntegration(t *testing.T) {
	db := testutil.SetupTestDB(t)

	t.Run("full flow: create movie, add reviews, verify leaderboard", func(t *testing.T) {
		testutil.WithTx(t, db, func(tx *gorm.DB) {
			// Create three movies
			movie1 := testutil.MovieFactory("The Godfather")
			movie2 := testutil.MovieFactory("The Dark Knight")
			movie3 := testutil.MovieFactory("Pulp Fiction")

			tx.Create(movie1)
			tx.Create(movie2)
			tx.Create(movie3)

			// Add reviews to movie1 (avg: 9.0)
			requests.CreateReview(tx, movie1.ID, 10, "Masterpiece!")
			requests.CreateReview(tx, movie1.ID, 8, "Very good")

			// Add reviews to movie2 (avg: 9.3)
			requests.CreateReview(tx, movie2.ID, 10, "Perfect!")
			requests.CreateReview(tx, movie2.ID, 9, "Great")
			requests.CreateReview(tx, movie2.ID, 9, "Awesome")

			// Add one review to movie3 (avg: 7.0)
			requests.CreateReview(tx, movie3.ID, 7, "Good but not great")

			// Get leaderboard
			leaderboard, err := requests.GetMoviesLeaderboard(tx)
			if err != nil {
				t.Fatalf("failed to get leaderboard: %v", err)
			}

			// Verify we have 3 movies
			if len(leaderboard) != 3 {
				t.Fatalf("expected 3 movies in leaderboard, got %d", len(leaderboard))
			}

			// Verify ranking (highest rating first)
			// movie2 should be rank 1 (rating: 9.3)
			if leaderboard[0].Title != "The Dark Knight" {
				t.Errorf("expected #1 to be 'The Dark Knight', got %s", leaderboard[0].Title)
			}
			if leaderboard[0].Rank != 1 {
				t.Errorf("expected rank 1, got %d", leaderboard[0].Rank)
			}
			if leaderboard[0].ReviewsCount != 3 {
				t.Errorf("expected 3 reviews, got %d", leaderboard[0].ReviewsCount)
			}

			// movie1 should be rank 2 (rating: 9.0)
			if leaderboard[1].Title != "The Godfather" {
				t.Errorf("expected #2 to be 'The Godfather', got %s", leaderboard[1].Title)
			}

			// movie3 should be rank 3 (rating: 7.0)
			if leaderboard[2].Title != "Pulp Fiction" {
				t.Errorf("expected #3 to be 'Pulp Fiction', got %s", leaderboard[2].Title)
			}
		})
	})
}

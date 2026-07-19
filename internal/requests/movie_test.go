package requests_test

import (
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/EvgeniiIvanov/hexlet-gorm/internal/models"
	"github.com/EvgeniiIvanov/hexlet-gorm/internal/requests"
	"github.com/EvgeniiIvanov/hexlet-gorm/internal/testutil"
)

func TestGetMovie(t *testing.T) {
	db := testutil.SetupTestDB(t)

	t.Run("retrieves movie with director and actors", func(t *testing.T) {
		testutil.WithTx(t, db, func(tx *gorm.DB) {
			// Create director
			director := testutil.DirectorFactory("Christopher Nolan")
			if err := tx.Create(director).Error; err != nil {
				t.Fatalf("failed to create director: %v", err)
			}

			// Create actors
			actor1 := testutil.ActorFactory("Leonardo DiCaprio")
			actor2 := testutil.ActorFactory("Tom Hardy")
			if err := tx.Create(&actor1).Error; err != nil {
				t.Fatalf("failed to create actor1: %v", err)
			}
			if err := tx.Create(&actor2).Error; err != nil {
				t.Fatalf("failed to create actor2: %v", err)
			}

			// Create movie with director and actors
			movie := testutil.MovieFactory("Inception")
			movie.DirectorID = &director.ID
			if err := tx.Create(movie).Error; err != nil {
				t.Fatalf("failed to create movie: %v", err)
			}

			// Associate actors
			if err := tx.Model(movie).Association("Actors").Append([]*models.Actor{actor1, actor2}); err != nil {
				t.Fatalf("failed to associate actors: %v", err)
			}

			// Get movie
			result, err := requests.GetMovie(tx, movie.ID)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			// Verify movie data
			if result.Title != "Inception" {
				t.Errorf("expected title 'Inception', got %s", result.Title)
			}

			// Verify director
			if result.Director.Name != "Christopher Nolan" {
				t.Errorf("expected director 'Christopher Nolan', got %s", result.Director.Name)
			}

			// Verify actors
			if len(result.Actors) != 2 {
				t.Fatalf("expected 2 actors, got %d", len(result.Actors))
			}
		})
	})

	t.Run("returns error for non-existent movie", func(t *testing.T) {
		testutil.WithTx(t, db, func(tx *gorm.DB) {
			_, err := requests.GetMovie(tx, 99999)
			if err == nil {
				t.Error("expected error for non-existent movie, got nil")
			}
		})
	})
}

func TestCreateMovie(t *testing.T) {
	db := testutil.SetupTestDB(t)

	t.Run("creates movie successfully", func(t *testing.T) {
		testutil.WithTx(t, db, func(tx *gorm.DB) {
			releasedAt := time.Date(2010, 7, 16, 0, 0, 0, 0, time.UTC)
			movie, err := requests.CreateMovie(tx, "Inception", "sci-fi", releasedAt)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if movie.ID == 0 {
				t.Error("expected movie to have an ID")
			}
			if movie.Title != "Inception" {
				t.Errorf("expected title 'Inception', got %s", movie.Title)
			}
			if movie.Genre != "sci-fi" {
				t.Errorf("expected genre 'sci-fi', got %s", movie.Genre)
			}
		})
	})

	t.Run("fails for duplicate title", func(t *testing.T) {
		testutil.WithTx(t, db, func(tx *gorm.DB) {
			releasedAt := time.Now()

			// Create first movie
			_, err := requests.CreateMovie(tx, "Unique Title", "drama", releasedAt)
			if err != nil {
				t.Fatalf("failed to create first movie: %v", err)
			}

			// Try to create duplicate
			_, err = requests.CreateMovie(tx, "Unique Title", "action", releasedAt)
			if err == nil {
				t.Error("expected error for duplicate title, got nil")
			}
		})
	})
}

func TestUpdateMovie(t *testing.T) {
	db := testutil.SetupTestDB(t)

	t.Run("updates movie field", func(t *testing.T) {
		testutil.WithTx(t, db, func(tx *gorm.DB) {
			// Create movie
			movie := testutil.MovieFactory("Interstellar")
			if err := tx.Create(movie).Error; err != nil {
				t.Fatalf("failed to create movie: %v", err)
			}

			// Update genre
			rowsAffected, err := requests.UpdateMovie(tx, movie.ID, "genre", "drama")
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if rowsAffected != 1 {
				t.Errorf("expected 1 row affected, got %d", rowsAffected)
			}

			// Verify update - get the updated movie
			updatedMovie, err := requests.GetMovie(tx, movie.ID)
			if err != nil {
				t.Fatalf("failed to find updated movie: %v", err)
			}

			if updatedMovie.Genre != "drama" {
				t.Errorf("expected genre 'drama', got %s", updatedMovie.Genre)
			}
		})
	})

	t.Run("returns 0 rows affected for non-existent movie", func(t *testing.T) {
		testutil.WithTx(t, db, func(tx *gorm.DB) {
			rowsAffected, err := requests.UpdateMovie(tx, 99999, "genre", "action")
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if rowsAffected != 0 {
				t.Errorf("expected 0 rows affected, got %d", rowsAffected)
			}
		})
	})
}

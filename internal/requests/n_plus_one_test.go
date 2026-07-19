package requests_test

import (
	"testing"

	"gorm.io/gorm"

	"github.com/EvgeniiIvanov/hexlet-gorm/internal/requests"
	"github.com/EvgeniiIvanov/hexlet-gorm/internal/testutil"
)

// TestNPlusOneQueries validates that we're using Preload correctly
// This is a documentation test - manual verification needed
func TestNPlusOneQueries(t *testing.T) {
	db := testutil.SetupTestDB(t)

	t.Run("ListMovies uses Preload to avoid N+1", func(t *testing.T) {
		testutil.WithTx(t, db, func(tx *gorm.DB) {
			// Create 10 movies with directors
			for i := 0; i < 10; i++ {
				director := testutil.DirectorFactory("")
				tx.Create(director)

				movie := testutil.MovieFactoryWithRandom()
				movie.DirectorID = &director.ID
				tx.Create(movie)
			}

			// List movies - should use Preload("Director")
			movies, err := requests.ListMovies(tx)
			if err != nil {
				t.Fatalf("failed to list movies: %v", err)
			}

			if len(movies) != 10 {
				t.Fatalf("expected 10 movies, got %d", len(movies))
			}

			// Verify directors are loaded
			for i, movie := range movies {
				if movie.Director.ID == 0 {
					t.Errorf("movie %d: director not preloaded", i)
				}
			}

			t.Log("✓ ListMovies correctly uses Preload(\"Director\")")
		})
	})

	t.Run("GetMovie preloads all relations", func(t *testing.T) {
		testutil.WithTx(t, db, func(tx *gorm.DB) {
			// Create movie with director and actors
			director := testutil.DirectorFactory("Test Director")
			tx.Create(director)

			movie := testutil.MovieFactory("Test Movie")
			movie.DirectorID = &director.ID
			tx.Create(movie)

			// Add 5 actors
			for i := 0; i < 5; i++ {
				actor := testutil.ActorFactory("")
				tx.Create(actor)
				tx.Model(movie).Association("Actors").Append(actor)
			}

			// Get movie with relations
			result, err := requests.GetMovie(tx, movie.ID)
			if err != nil {
				t.Fatalf("failed to get movie: %v", err)
			}

			// Verify director is loaded
			if result.Director.ID == 0 {
				t.Error("director not preloaded")
			}

			// Verify actors are loaded
			if len(result.Actors) != 5 {
				t.Errorf("expected 5 actors, got %d", len(result.Actors))
			}

			t.Log("✓ GetMovie correctly uses Preload(\"Director\") and Preload(\"Actors\")")
		})
	})
}

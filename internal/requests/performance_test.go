package requests_test

import (
	"testing"

	"gorm.io/gorm"

	"github.com/EvgeniiIvanov/hexlet-gorm/internal/requests"
	"github.com/EvgeniiIvanov/hexlet-gorm/internal/testutil"
)

// BenchmarkListMovies measures performance of listing movies with directors
func BenchmarkListMovies(b *testing.B) {
	db := testutil.SetupTestDB(&testing.T{})

	// Create test data: 100 movies with directors
	testutil.WithTx(&testing.T{}, db, func(tx *gorm.DB) {
		for i := 0; i < 100; i++ {
			director := testutil.DirectorFactory("")
			tx.Create(director)

			movie := testutil.MovieFactoryWithRandom()
			movie.DirectorID = &director.ID
			tx.Create(movie)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = requests.ListMovies(tx)
		}
	})
}

// BenchmarkGetMostRatedMovies measures filtering by rating
func BenchmarkGetMostRatedMovies(b *testing.B) {
	db := testutil.SetupTestDB(&testing.T{})

	testutil.WithTx(&testing.T{}, db, func(tx *gorm.DB) {
		// Create 100 movies with varying ratings
		for i := 0; i < 100; i++ {
			movie := testutil.MovieFactoryWithRandom()
			tx.Create(movie)

			// Add reviews to create ratings
			rating := 5.0 + float64(i%6) // Ratings from 5.0 to 10.0
			tx.Model(movie).Update("rating", rating)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = requests.GetMostRatedMovies(tx)
		}
	})
}

// BenchmarkGetUnratedMovies measures filtering by NULL rating
func BenchmarkGetUnratedMovies(b *testing.B) {
	db := testutil.SetupTestDB(&testing.T{})

	testutil.WithTx(&testing.T{}, db, func(tx *gorm.DB) {
		// Create 50 rated and 50 unrated movies
		for i := 0; i < 100; i++ {
			movie := testutil.MovieFactoryWithRandom()
			tx.Create(movie)

			if i < 50 {
				tx.Model(movie).Update("rating", 8.0)
			}
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = requests.GetUnratedMovies(tx)
		}
	})
}

// BenchmarkGetMoviesLeaderboard measures complex aggregation query
func BenchmarkGetMoviesLeaderboard(b *testing.B) {
	db := testutil.SetupTestDB(&testing.T{})

	testutil.WithTx(&testing.T{}, db, func(tx *gorm.DB) {
		// Create 50 movies with multiple reviews each
		for i := 0; i < 50; i++ {
			movie := testutil.MovieFactoryWithRandom()
			tx.Create(movie)

			// Add 3-5 reviews per movie
			for j := 0; j < 3+i%3; j++ {
				review := testutil.ReviewFactory(movie.ID, 7+j%4, "")
				tx.Create(review)
			}
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = requests.GetMoviesLeaderboard(tx)
		}
	})
}

// BenchmarkCreateReviewWithRatingUpdate measures review creation + hook
func BenchmarkCreateReviewWithRatingUpdate(b *testing.B) {
	db := testutil.SetupTestDB(&testing.T{})

	testutil.WithTx(&testing.T{}, db, func(tx *gorm.DB) {
		// Create a movie
		movie := testutil.MovieFactory("Test Movie")
		tx.Create(movie)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Each iteration creates a review with unique text
			_ = requests.CreateReview(tx, movie.ID, 8, testutil.ReviewFactory(0, 0, "").Text)
		}
	})
}

// BenchmarkGetMovieWithRelations measures preloading performance
func BenchmarkGetMovieWithRelations(b *testing.B) {
	db := testutil.SetupTestDB(&testing.T{})

	testutil.WithTx(&testing.T{}, db, func(tx *gorm.DB) {
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

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = requests.GetMovie(tx, movie.ID)
		}
	})
}

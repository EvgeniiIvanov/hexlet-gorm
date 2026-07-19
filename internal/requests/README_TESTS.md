# Testing Guide

## Setup

First, install the required dependencies:

```bash
go get gorm.io/driver/sqlite
go get github.com/go-faker/faker/v4
go mod tidy
```

## Running Tests

Run all tests:
```bash
go test ./internal/requests/...
```

Run tests with verbose output:
```bash
go test -v ./internal/requests/...
```

Run a specific test:
```bash
go test -v -run TestCreateReview ./internal/requests/
```

## Test Structure

### Test Database
- Uses in-memory SQLite (`:memory:`)
- No cleanup needed, database is discarded after tests
- Fast and isolated

### Transactions
- Each test runs in a transaction
- Transaction is rolled back after test completes
- Tests are fully isolated from each other
- No data persists between tests

### Factories
Located in `internal/testutil/factories.go`:

- `MovieFactory(title)` - Creates a movie with given title
- `MovieFactoryWithRandom()` - Creates a movie with random faker data
- `ReviewFactory(movieID, score, text)` - Creates a review
- `DirectorFactory(name)` - Creates a director
- `ActorFactory(name)` - Creates an actor

### Example Test

```go
func TestExample(t *testing.T) {
    db := testutil.SetupTestDB(t)
    
    t.Run("description", func(t *testing.T) {
        testutil.WithTx(t, db, func(tx *gorm.DB) {
            // Create test data
            movie := testutil.MovieFactory("Inception")
            tx.Create(movie)
            
            // Run your test
            result, err := requests.GetMovie(tx, movie.ID)
            
            // Assert
            if err != nil {
                t.Fatalf("unexpected error: %v", err)
            }
            if result.Title != "Inception" {
                t.Errorf("expected 'Inception', got %s", result.Title)
            }
        })
    })
}
```

## Coverage

Run tests with coverage:
```bash
go test -cover ./internal/requests/...
```

Generate HTML coverage report:
```bash
go test -coverprofile=coverage.out ./internal/requests/...
go tool cover -html=coverage.out
```

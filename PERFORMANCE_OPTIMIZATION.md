# Performance Optimization Journey

## Baseline Measurements

Run benchmarks to establish baseline:

```bash
go test -bench=. -benchmem ./internal/requests/
```

## Phase 1: Initial Analysis

### N+1 Query Detection

Run N+1 detection tests:

```bash
go test -v -run TestNPlusOneQueries ./internal/requests/
```

**Expected Results:**
- ✅ ListMovies: Should be 2 queries (movies + directors preload)
- ✅ GetMovie: Should be 3 queries (movie + director + actors preload)

### Benchmark Results (BEFORE optimization)

```
# Run and record baseline results
go test -bench=. -benchmem ./internal/requests/ > baseline_results.txt
```

## Phase 2: Optimization Strategies

### 1. Add Database Indexes

**Rationale:** Queries filtering by `genre`, `rating`, and sorting by `rating` will benefit from indexes.

```sql
-- Index for filtering by genre
CREATE INDEX IF NOT EXISTS idx_movies_genre ON movies(genre);

-- Index for filtering by rating (most_rated, leaderboard)
CREATE INDEX IF NOT EXISTS idx_movies_rating ON movies(rating DESC);

-- Composite index for unrated queries
CREATE INDEX IF NOT EXISTS idx_movies_rating_null ON movies(rating) WHERE rating IS NULL;

-- Index for review lookups by movie
CREATE INDEX IF NOT EXISTS idx_reviews_movie_id ON reviews(movie_id);

-- Index for foreign key lookups
CREATE INDEX IF NOT EXISTS idx_movies_director_id ON movies(director_id);
```

**Expected Impact:**
- `GetMostRatedMovies`: Faster filtering on `rating > 8.5`
- `GetUnratedMovies`: Faster `rating IS NULL` checks
- `GetMoviesLeaderboard`: Faster sorting by rating
- Review queries: Faster joins on `movie_id`

### 2. Optimize ReviewsCount Column

**Problem:** Currently recalculated on every review insert via hook.

**Current approach:**
```go
// AfterCreate hook recalculates from database
SELECT COUNT(*) FROM reviews WHERE movie_id = ?
```

**This is already optimal** because:
- ✅ Always accurate (source of truth)
- ✅ Only runs on review creation (not on every read)
- ✅ Using database aggregate (fast)

**Alternative (NOT recommended):**
- Incremental counter: Faster but fragile (can drift on errors)
- Cached value: Complex invalidation logic

**Decision:** Keep current approach (accuracy > micro-optimization)

### 3. Verify Preloading Efficiency

**Check:** Ensure we're using `Preload()` correctly to avoid N+1.

```go
// ✅ Good: Single query per relation
db.Preload("Director").Preload("Actors").Find(&movies)

// ❌ Bad: N+1 queries
db.Find(&movies)
for _, m := range movies {
    db.Model(&m).Association("Director").Find(&m.Director)
}
```

## Phase 3: Implementation

### Add Indexes to Models

Update `internal/models/models.go`:

```go
type Movie struct {
    // ...
    Genre  string `gorm:"not null;index:idx_movies_genre"`
    Rating float64 `gorm:"type:numeric(3,1);index:idx_movies_rating"`
    // ...
}

type Review struct {
    // ...
    MovieID uint `gorm:"index:idx_reviews_movie_id"`
    // ...
}
```

### Add Migration for Existing Databases

For PostgreSQL production database:

```sql
-- Run these in pgAdmin or via migration tool
CREATE INDEX IF NOT EXISTS idx_movies_genre ON movies(genre);
CREATE INDEX IF NOT EXISTS idx_movies_rating ON movies(rating DESC);
CREATE INDEX IF NOT EXISTS idx_reviews_movie_id ON reviews(movie_id);
CREATE INDEX IF NOT EXISTS idx_movies_director_id ON movies(director_id);
```

## Phase 4: Measure Improvements

### Re-run Benchmarks

```bash
go test -bench=. -benchmem ./internal/requests/ > optimized_results.txt
```

### Compare Results

```bash
# Install benchstat for comparison
go install golang.org/x/perf/cmd/benchstat@latest

# Compare baseline vs optimized
benchstat baseline_results.txt optimized_results.txt
```

### Re-run N+1 Tests

```bash
go test -v -run TestNPlusOneQueries ./internal/requests/
```

## Expected Improvements

| Operation | Before | After | Improvement |
|-----------|--------|-------|-------------|
| GetMostRatedMovies | ? | ? | ~50% faster with index |
| GetUnratedMovies | ? | ? | ~40% faster with index |
| GetMoviesLeaderboard | ? | ? | ~30% faster with index |
| ListMovies | No N+1 | No N+1 | Already optimal |

## Monitoring Queries

### Enable Query Logging in Development

```go
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
    Logger: logger.Default.LogMode(logger.Info), // Shows all SQL
})
```

### Analyze Slow Queries

In PostgreSQL, enable slow query log:

```sql
-- Find slow queries
SELECT query, mean_exec_time, calls 
FROM pg_stat_statements 
ORDER BY mean_exec_time DESC 
LIMIT 10;
```

## Best Practices Applied

1. ✅ **Indexes on filtered columns** (genre, rating)
2. ✅ **Indexes on foreign keys** (director_id, movie_id)
3. ✅ **Preload relations** to avoid N+1
4. ✅ **Database aggregates** for counting (COUNT, AVG)
5. ✅ **Composite indexes** where multiple columns are queried together

## Next Steps

1. Run baseline benchmarks
2. Add indexes to models
3. Run migration on PostgreSQL
4. Re-run benchmarks
5. Document results
6. Monitor production query performance

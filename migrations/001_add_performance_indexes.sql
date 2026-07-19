-- Migration: Add Performance Indexes
-- Created: 2026-07-19
-- Description: Adds indexes to improve query performance on filtered and sorted columns

-- Movies table indexes
CREATE INDEX IF NOT EXISTS idx_movies_genre ON movies(genre);
CREATE INDEX IF NOT EXISTS idx_movies_rating ON movies(rating);
CREATE INDEX IF NOT EXISTS idx_movies_reviews_count ON movies(reviews_count);
CREATE INDEX IF NOT EXISTS idx_movies_director_id ON movies(director_id);

-- Reviews table indexes
CREATE INDEX IF NOT EXISTS idx_reviews_movie_id ON reviews(movie_id);

-- Verify indexes were created
SELECT 
    schemaname,
    tablename,
    indexname,
    indexdef
FROM pg_indexes 
WHERE tablename IN ('movies', 'reviews')
  AND schemaname = 'public'
ORDER BY tablename, indexname;

-- Expected output:
-- idx_movies_director_id
-- idx_movies_genre
-- idx_movies_rating
-- idx_movies_reviews_count
-- idx_reviews_movie_id
-- reviews_movie_text_idx (unique constraint from earlier)

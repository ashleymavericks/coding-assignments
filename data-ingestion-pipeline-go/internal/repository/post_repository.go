package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/anurag/data-ingestion-pipeline-go/internal/models"
	"github.com/anurag/data-ingestion-pipeline-go/pkg/database"
	"github.com/anurag/data-ingestion-pipeline-go/pkg/logger"
)

// PostRepository interface defines post data operations
// Go Concept: Interface segregation - only define what we need
type PostRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, post *models.Post) error
	GetByID(ctx context.Context, id int) (*models.Post, error)
	Update(ctx context.Context, post *models.Post) error
	Delete(ctx context.Context, id int) error

	// Query operations
	GetAll(ctx context.Context, filter *models.PostFilter) ([]models.Post, error)
	GetByUserID(ctx context.Context, userID int) ([]models.Post, error)
	GetStats(ctx context.Context) (*models.PostStats, error)

	// Batch operations
	CreateBatch(ctx context.Context, posts []models.Post) error

	// Existence checks
	Exists(ctx context.Context, id int) (bool, error)

	// Cleanup operations
	DeleteOldPosts(ctx context.Context, olderThan time.Time) (int64, error)
}

// postRepository implements PostRepository interface
// Go Concept: Struct implementing interface with dependencies
type postRepository struct {
	db     database.DB   // Database interface for testability
	logger logger.Logger // Logger for observability
}

// NewPostRepository creates a new post repository
// Go Concept: Constructor function with dependency injection
func NewPostRepository(db database.DB, logger logger.Logger) PostRepository {
	return &postRepository{
		db:     db,
		logger: logger,
	}
}

// Create inserts a new post into the database
// Go Concept: Method with context, pointer receiver, and error handling
func (pr *postRepository) Create(ctx context.Context, post *models.Post) error {
	// Validate the post before insertion
	if err := post.Validate(); err != nil {
		pr.logger.Error("Post validation failed", logger.Error(err))
		return fmt.Errorf("invalid post data: %w", err)
	}

	// SQL query for inserting a post
	// Go Concept: Multi-line string literals with backticks
	query := `
		INSERT INTO posts (id, user_id, title, body, ingested_at, source, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	// Prepare timestamp values
	now := time.Now().UTC()

	// Execute the query
	// Go Concept: Variadic arguments (...interface{}) for SQL parameters
	result, err := pr.db.ExecContext(ctx, query,
		post.ID,
		post.UserID,
		post.Title,
		post.Body,
		post.IngestedAt,
		post.Source,
		now, // created_at
		now, // updated_at
	)

	if err != nil {
		pr.logger.Error("Failed to create post",
			logger.Int("post_id", post.ID),
			logger.Error(err),
		)
		return models.NewDatabaseError("INSERT", "posts", err)
	}

	// Check if the insert was successful
	// Go Concept: Checking return values from database operations
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		pr.logger.Warn("Could not get rows affected", logger.Error(err))
	} else if rowsAffected == 0 {
		return fmt.Errorf("post creation failed: no rows affected")
	}

	pr.logger.Info("Post created successfully",
		logger.Int("post_id", post.ID),
		logger.Int("user_id", post.UserID),
		logger.String("source", post.Source),
	)

	return nil
}

// GetByID retrieves a post by its ID
// Go Concept: Method returning pointer and error
func (pr *postRepository) GetByID(ctx context.Context, id int) (*models.Post, error) {
	query := `
		SELECT id, user_id, title, body, ingested_at, source, created_at, updated_at
		FROM posts
		WHERE id = ?
	`

	// QueryRowContext returns a single row
	// Go Concept: Single row query pattern
	row := pr.db.QueryRowContext(ctx, query, id)

	// Scan the row into a Post struct
	// Go Concept: Scanning database results into struct fields
	post := &models.Post{}
	var createdAt, updatedAt time.Time

	err := row.Scan(
		&post.ID,
		&post.UserID,
		&post.Title,
		&post.Body,
		&post.IngestedAt,
		&post.Source,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// Go Concept: Handling specific error types
			pr.logger.Debug("Post not found", logger.Int("post_id", id))
			return nil, models.ErrPostNotFound
		}
		pr.logger.Error("Failed to get post by ID",
			logger.Int("post_id", id),
			logger.Error(err),
		)
		return nil, models.NewDatabaseError("SELECT", "posts", err)
	}

	pr.logger.Debug("Post retrieved successfully", logger.Int("post_id", id))
	return post, nil
}

// Update modifies an existing post
// Go Concept: UPDATE operation with optimistic locking consideration
func (pr *postRepository) Update(ctx context.Context, post *models.Post) error {
	if err := post.Validate(); err != nil {
		return fmt.Errorf("invalid post data: %w", err)
	}

	query := `
		UPDATE posts 
		SET user_id = ?, title = ?, body = ?, ingested_at = ?, source = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := pr.db.ExecContext(ctx, query,
		post.UserID,
		post.Title,
		post.Body,
		post.IngestedAt,
		post.Source,
		time.Now().UTC(), // updated_at
		post.ID,
	)

	if err != nil {
		pr.logger.Error("Failed to update post",
			logger.Int("post_id", post.ID),
			logger.Error(err),
		)
		return models.NewDatabaseError("UPDATE", "posts", err)
	}

	// Check if the update actually modified a row
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		pr.logger.Warn("Could not get rows affected", logger.Error(err))
	} else if rowsAffected == 0 {
		return models.ErrPostNotFound
	}

	pr.logger.Info("Post updated successfully", logger.Int("post_id", post.ID))
	return nil
}

// Delete removes a post by ID
func (pr *postRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM posts WHERE id = ?`

	result, err := pr.db.ExecContext(ctx, query, id)
	if err != nil {
		pr.logger.Error("Failed to delete post",
			logger.Int("post_id", id),
			logger.Error(err),
		)
		return models.NewDatabaseError("DELETE", "posts", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		pr.logger.Warn("Could not get rows affected", logger.Error(err))
	} else if rowsAffected == 0 {
		return models.ErrPostNotFound
	}

	pr.logger.Info("Post deleted successfully", logger.Int("post_id", id))
	return nil
}

// GetAll retrieves posts with optional filtering
// Go Concept: Dynamic query building and filtering
func (pr *postRepository) GetAll(ctx context.Context, filter *models.PostFilter) ([]models.Post, error) {
	// Build dynamic query based on filter
	// Go Concept: String building for dynamic SQL
	query := `
		SELECT id, user_id, title, body, ingested_at, source, created_at, updated_at
		FROM posts
	`

	var conditions []string
	var args []interface{}

	// Add WHERE conditions based on filter
	if filter != nil {
		if filter.UserID != nil {
			conditions = append(conditions, "user_id = ?")
			args = append(args, *filter.UserID)
		}

		if filter.Since != nil {
			conditions = append(conditions, "ingested_at >= ?")
			args = append(args, *filter.Since)
		}
	}

	// Add WHERE clause if we have conditions
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Add ORDER BY clause
	query += " ORDER BY ingested_at DESC"

	// Add LIMIT and OFFSET for pagination
	if filter != nil {
		if filter.Limit > 0 {
			query += " LIMIT ?"
			args = append(args, filter.Limit)
		}

		if filter.Offset > 0 {
			query += " OFFSET ?"
			args = append(args, filter.Offset)
		}
	}

	// Execute query
	// Go Concept: Multiple row query pattern
	rows, err := pr.db.QueryContext(ctx, query, args...)
	if err != nil {
		pr.logger.Error("Failed to get posts",
			logger.Any("filter", filter),
			logger.Error(err),
		)
		return nil, models.NewDatabaseError("SELECT", "posts", err)
	}
	defer rows.Close() // Always close rows

	// Scan all rows into slice
	// Go Concept: Slice building from database results
	var posts []models.Post
	for rows.Next() {
		var post models.Post
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.Title,
			&post.Body,
			&post.IngestedAt,
			&post.Source,
			&createdAt,
			&updatedAt,
		)

		if err != nil {
			pr.logger.Error("Failed to scan post row", logger.Error(err))
			return nil, fmt.Errorf("failed to scan post: %w", err)
		}

		posts = append(posts, post)
	}

	// Check for iteration errors
	// Go Concept: Checking rows.Err() after iteration
	if err = rows.Err(); err != nil {
		pr.logger.Error("Error iterating over posts", logger.Error(err))
		return nil, fmt.Errorf("error iterating posts: %w", err)
	}

	pr.logger.Debug("Posts retrieved successfully",
		logger.Int("count", len(posts)),
		logger.Any("filter", filter),
	)

	return posts, nil
}

// GetByUserID retrieves all posts for a specific user
// Go Concept: Convenience method wrapping GetAll
func (pr *postRepository) GetByUserID(ctx context.Context, userID int) ([]models.Post, error) {
	filter := &models.PostFilter{
		UserID: &userID,
	}
	return pr.GetAll(ctx, filter)
}

// GetStats retrieves statistics about posts
// Go Concept: Aggregate queries and custom result types
func (pr *postRepository) GetStats(ctx context.Context) (*models.PostStats, error) {
	query := `
		SELECT 
			COUNT(*) as total_posts,
			MAX(ingested_at) as last_ingestion,
			COUNT(DISTINCT user_id) as unique_users
		FROM posts
	`

	row := pr.db.QueryRowContext(ctx, query)

	stats := &models.PostStats{}
	var lastIngestion sql.NullTime // Handle potential NULL values

	err := row.Scan(
		&stats.TotalPosts,
		&lastIngestion,
		&stats.UniqueUsers,
	)

	if err != nil {
		pr.logger.Error("Failed to get post stats", logger.Error(err))
		return nil, models.NewDatabaseError("SELECT", "posts", err)
	}

	// Handle NULL timestamp
	// Go Concept: Handling NULL values from database
	if lastIngestion.Valid {
		stats.LastIngestion = lastIngestion.Time
	}

	pr.logger.Debug("Post stats retrieved successfully",
		logger.Int("total_posts", stats.TotalPosts),
		logger.Int("unique_users", stats.UniqueUsers),
	)

	return stats, nil
}

// CreateBatch inserts multiple posts in a single transaction
// Go Concept: Batch operations and transaction management
func (pr *postRepository) CreateBatch(ctx context.Context, posts []models.Post) error {
	if len(posts) == 0 {
		return nil // Nothing to do
	}

	// Start a transaction for batch operation
	// Go Concept: Transaction for atomicity
	tx, err := pr.db.BeginTx(ctx)
	if err != nil {
		pr.logger.Error("Failed to begin transaction", logger.Error(err))
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Ensure rollback if not committed

	// Prepare the insert statement
	// Go Concept: Prepared statements for efficiency
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO posts (id, user_id, title, body, ingested_at, source, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		pr.logger.Error("Failed to prepare batch insert statement", logger.Error(err))
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	now := time.Now().UTC()
	successCount := 0

	// Execute insert for each post
	for i, post := range posts {
		// Validate each post
		if err := post.Validate(); err != nil {
			pr.logger.Warn("Skipping invalid post in batch",
				logger.Int("index", i),
				logger.Int("post_id", post.ID),
				logger.Error(err),
			)
			continue
		}

		_, err = stmt.ExecContext(ctx,
			post.ID,
			post.UserID,
			post.Title,
			post.Body,
			post.IngestedAt,
			post.Source,
			now,
			now,
		)

		if err != nil {
			pr.logger.Error("Failed to insert post in batch",
				logger.Int("index", i),
				logger.Int("post_id", post.ID),
				logger.Error(err),
			)
			// Continue with other posts rather than failing entire batch
			continue
		}

		successCount++
	}

	// Commit the transaction
	// Go Concept: Explicit transaction commit
	if err = tx.Commit(); err != nil {
		pr.logger.Error("Failed to commit batch transaction", logger.Error(err))
		return fmt.Errorf("failed to commit batch: %w", err)
	}

	pr.logger.Info("Batch insert completed",
		logger.Int("total_posts", len(posts)),
		logger.Int("successful_inserts", successCount),
		logger.Int("failed_inserts", len(posts)-successCount),
	)

	return nil
}

// Exists checks if a post with the given ID exists
// Go Concept: Existence check with minimal data transfer
func (pr *postRepository) Exists(ctx context.Context, id int) (bool, error) {
	query := `SELECT 1 FROM posts WHERE id = ? LIMIT 1`

	var exists int
	err := pr.db.QueryRowContext(ctx, query, id).Scan(&exists)

	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil // Post doesn't exist
		}
		pr.logger.Error("Failed to check post existence",
			logger.Int("post_id", id),
			logger.Error(err),
		)
		return false, fmt.Errorf("failed to check existence: %w", err)
	}

	return true, nil
}

// DeleteOldPosts removes posts older than the specified time
// Go Concept: Cleanup operations and bulk deletes
func (pr *postRepository) DeleteOldPosts(ctx context.Context, olderThan time.Time) (int64, error) {
	query := `DELETE FROM posts WHERE ingested_at < ?`

	result, err := pr.db.ExecContext(ctx, query, olderThan)
	if err != nil {
		pr.logger.Error("Failed to delete old posts",
			logger.Time("older_than", olderThan),
			logger.Error(err),
		)
		return 0, models.NewDatabaseError("DELETE", "posts", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		pr.logger.Warn("Could not get rows affected for cleanup", logger.Error(err))
		return 0, nil
	}

	pr.logger.Info("Old posts cleaned up",
		logger.Time("older_than", olderThan),
		logger.Int64("deleted_count", rowsAffected),
	)

	return rowsAffected, nil
}

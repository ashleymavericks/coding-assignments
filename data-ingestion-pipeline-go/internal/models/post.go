package models

import (
	"time"
)

// Post represents a blog post from the JSONPlaceholder API
// Go Concept: Struct definition with JSON and database tags
type Post struct {
	// ID is the unique identifier for the post
	// JSON tag: tells Go how to marshal/unmarshal JSON
	// db tag: used by database libraries for column mapping
	ID int `json:"id" db:"id"`

	// UserID identifies which user created the post
	UserID int `json:"userId" db:"user_id"`

	// Title is the post's title
	Title string `json:"title" db:"title"`

	// Body contains the post content
	Body string `json:"body" db:"body"`

	// IngestedAt is added during our transformation process
	// This demonstrates how we extend external API data
	// time.Time is Go's standard time type
	IngestedAt time.Time `json:"ingested_at" db:"ingested_at"`

	// Source identifies where this data came from
	// This is a static field we add to track data lineage
	Source string `json:"source" db:"source"`
}

// PostResponse represents the API response structure
// Go Concept: Separating external API structure from internal models
type PostResponse struct {
	ID     int    `json:"id"`
	UserID int    `json:"userId"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}

// ToPost converts a PostResponse to our internal Post model
// Go Concept: Method with receiver - this is how Go does "methods" on types
// The (pr PostResponse) part makes this a method on PostResponse
func (pr PostResponse) ToPost() Post {
	return Post{
		ID:         pr.ID,
		UserID:     pr.UserID,
		Title:      pr.Title,
		Body:       pr.Body,
		IngestedAt: time.Now().UTC(),  // Always use UTC for consistency
		Source:     "placeholder_api", // Static source identifier
	}
}

// PostFilter represents query parameters for filtering posts
// Go Concept: Struct for encapsulating query parameters
type PostFilter struct {
	UserID *int       `json:"user_id,omitempty"` // Pointer allows nil (optional)
	Since  *time.Time `json:"since,omitempty"`   // Filter posts after this time
	Limit  int        `json:"limit,omitempty"`   // Maximum number of results
	Offset int        `json:"offset,omitempty"`  // Pagination offset
}

// PostStats represents statistics about ingested posts
// Go Concept: Struct for aggregated data
type PostStats struct {
	TotalPosts    int       `json:"total_posts"`
	LastIngestion time.Time `json:"last_ingestion"`
	UniqueUsers   int       `json:"unique_users"`
}

// Validate checks if a Post has all required fields
// Go Concept: Method for data validation, returns error
func (p Post) Validate() error {
	if p.ID <= 0 {
		return ErrInvalidPostID
	}
	if p.UserID <= 0 {
		return ErrInvalidUserID
	}
	if p.Title == "" {
		return ErrEmptyTitle
	}
	if p.Source == "" {
		return ErrEmptySource
	}
	return nil
}

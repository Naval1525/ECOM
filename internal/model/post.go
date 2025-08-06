package model

import (
	"time"

	"github.com/docker/distribution/uuid"
)

// Post represents a social media post
type Post struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Content   string    `json:"content" db:"content"`
	ImageURL  string    `json:"image_url" db:"image_url"`
	LikeCount int       `json:"like_count" db:"like_count"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	// Joined fields (not stored in DB, populated via JOINs)
	Author  *UserResponse `json:"author,omitempty"`
	IsLiked bool          `json:"is_liked"`
}

// PostRequest represents the JSON structure for creating posts
type PostRequest struct {
	Content  string `json:"content" validate:"required,max=500"`
	ImageURL string `json:"image_url"`
}

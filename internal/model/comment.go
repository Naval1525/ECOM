package model

import (
	"time"

	"github.com/docker/distribution/uuid"
)

type Comment struct {
	ID        uuid.UUID `json:"id" db:"id"`
	PostID    uuid.UUID `json:"post_id" db:"post_id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Content   string    `json:"content" db:"content"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	// Joined fields
	Author *UserResponse `json:"author,omitempty"`
}

// CommentRequest represents the JSON structure for creating comments
type CommentRequest struct {
	Content string `json:"content" validate:"required,max=500"`
}

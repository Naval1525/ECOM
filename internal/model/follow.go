package model

import (
	"time"

	"github.com/docker/distribution/uuid"
)

type Follow struct {
	ID          uuid.UUID `json:"id" db:"id"`
	FollowerID  uuid.UUID `json:"follower_id" db:"follower_id"`   // User who follows
	FollowingID uuid.UUID `json:"following_id" db:"following_id"` // User being followed
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// Like represents a like on a post
type Like struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	PostID    uuid.UUID `json:"post_id" db:"post_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

package repository

import (
	"context"

	"github.com/docker/distribution/uuid"
	"github.com/naval1525/ECOM/internal/model"
)

// UserRepository defines methods for user data operations
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetById(ctx context.Context, id uuid.UUID) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type PostRepository interface {
	Create(ctx context.Context, post *model.Post) error
	GetById(ctx context.Context, id uuid.UUID) ([]*model.Post, error)
	GetByUserId(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*model.Post, error)
	GetFeed(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*model.Post, error)
	Update(ctx context.Context, post *model.Post) error
	Delete(ctx context.Context, id uuid.UUID) error
	IncrementLikeCount(ctx context.Context, postid uuid.UUID) error
	DecrementLikeCount(ctx context.Context, postid uuid.UUID) error
}

type CommentRepository interface {
	Create(ctx context.Context, comment *model.Comment) error
	GetByPostID(ctx context.Context, postID uuid.UUID, limit, offset int) ([]*model.Comment, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type FollowRepository interface {
	Follow(ctx context.Context, followerID, followingID uuid.UUID) error
	Unfollow(ctx context.Context, followerID, followingID uuid.UUID) error
	IsFollowing(ctx context.Context, followerID, followingID uuid.UUID) (bool, error)
	GetFollowers(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*model.User, error)
	GetFollowing(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*model.User, error)
}

type LikeRepository interface {
	Like(ctx context.Context, userID, postID uuid.UUID) error
	Unlike(ctx context.Context, userID, postID uuid.UUID) error
	IsLiked(ctx context.Context, userID, postID uuid.UUID) (bool, error)
}

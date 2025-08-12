package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/naval1525/Social_Media_Backend/internal/model"
)

type UserService interface {
	Register(ctx context.Context, req *model.UserRequest) (*model.UserResponse, error)
	Login(ctx context.Context, email, password string) (*model.UserResponse, string, error)
	GetProfile(ctx context.Context, userID uuid.UUID) (*model.UserResponse, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, updates map[string]interface{}) (*model.UserResponse, error)
    ValidateJWT(tokenString string) (uuid.UUID, error)
}

package service

import (
	"context"
	"fmt"

	"github.com/naval1525/ECOM/internal/model"
	"github.com/naval1525/ECOM/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type userService struct {
	userRepo  repository.UserRepository
	jwtSecret string
}

// NewUserService create a new user service
func NewUserService(userRepo repository.UserRepository, jwtSecret string) UserService {
	return &userService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

// Regsiter create a new user account
func (s *userService) Register(ctx context.Context, req *model.UserRequest) (*model.UserResponse, error) {
	//check if exisitng user
	existingUser, _ := s.userRepo.GetByEmail(ctx, req.Email)
	if existingUser != nil {
		return nil, fmt.Errorf("user with email %s already exists", req.Email)
	}
	existingUser, _ = s.userRepo.GetByUsername(ctx, req.Username)
	if existingUser != nil {
		return nil, fmt.Errorf("user with username %s already exists", req.Username)
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %v", err)
	}
	user := &model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		FullName: req.FullName,
		Bio:      "",
		Avatar:   "",
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}
	return &model.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		FullName:  user.FullName,
		Bio:       user.Bio,
		Avatar:    user.Avatar,
		CreatedAt: user.CreatedAt,
	}, nil

}

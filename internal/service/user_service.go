package service

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"

	"github.com/naval1525/Social_Media_Backend/internal/model"
	"github.com/naval1525/Social_Media_Backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type userService struct {
	userRepo  repository.UserRepository
	jwtSecret string
}

// NewUserService creates a new user service
func NewUserService(userRepo repository.UserRepository, jwtSecret string) UserService {
	return &userService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

// Register creates a new user account
func (s *userService) Register(ctx context.Context, req *model.UserRequest) (*model.UserResponse, error) {
	// Check if user already exists
	existingUser, _ := s.userRepo.GetByEmail(ctx, req.Email)
	if existingUser != nil {
		return nil, fmt.Errorf("user with email %s already exists", req.Email)
	}

	existingUser, _ = s.userRepo.GetByUsername(ctx, req.Username)
	if existingUser != nil {
		return nil, fmt.Errorf("username %s is already taken", req.Username)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user model
	user := &model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		FullName: req.FullName,
		Bio:      "",
		Avatar:   "",
	}

	// Save to database
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Return user response (without password)
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

// Login authenticates a user and returns a JWT token
func (s *userService) Login(ctx context.Context, email, password string) (*model.UserResponse, string, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, "", fmt.Errorf("invalid email or password")
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, "", fmt.Errorf("invalid email or password")
	}

	// Generate JWT token
	token, err := s.generateJWT(user.ID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	// Return user response and token
	userResponse := &model.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		FullName:  user.FullName,
		Bio:       user.Bio,
		Avatar:    user.Avatar,
		CreatedAt: user.CreatedAt,
	}

	return userResponse, token, nil
}

// GetProfile retrieves a user's profile
func (s *userService) GetProfile(ctx context.Context, userID uuid.UUID) (*model.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
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

// UpdateProfile updates a user's profile information
func (s *userService) UpdateProfile(ctx context.Context, userID uuid.UUID, updates map[string]interface{}) (*model.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Update fields if provided
	if fullName, ok := updates["full_name"]; ok {
		user.FullName = fullName.(string)
	}
	if bio, ok := updates["bio"]; ok {
		user.Bio = bio.(string)
	}
	if avatar, ok := updates["avatar"]; ok {
		user.Avatar = avatar.(string)
	}

	// Save changes
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
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

// generateJWT creates a JWT token for the user
func (s *userService) generateJWT(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(), // 7 days
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// ValidateJWT validates and parses a JWT token
func (s *userService) ValidateJWT(tokenString string) (uuid.UUID, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return uuid.UUID{}, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userIDStr, ok := claims["user_id"].(string)
		if !ok {
			return uuid.UUID{}, fmt.Errorf("invalid user_id in token")
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return uuid.UUID{}, fmt.Errorf("invalid user_id format: %w", err)
		}

		return userID, nil
	}

	return uuid.UUID{}, fmt.Errorf("invalid token")
}

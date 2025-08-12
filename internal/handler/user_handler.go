package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/naval1525/Social_Media_Backend/internal/model"
	"github.com/naval1525/Social_Media_Backend/internal/service"
)

type UserHandler struct {
	userService service.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// Register handles user registration
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.UserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Basic validation
	if req.Username == "" || req.Email == "" || req.Password == "" || req.FullName == "" {
		writeErrorResponse(w, http.StatusBadRequest, "All fields are required")
		return
	}

	user, err := h.userService.Register(r.Context(), &req)
	if err != nil {
		writeErrorResponse(w, http.StatusConflict, err.Error())
		return
	}

	writeSuccessResponse(w, http.StatusCreated, "User registered successfully", user)
}

// Login handles user authentication
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		writeErrorResponse(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	user, token, err := h.userService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		writeErrorResponse(w, http.StatusUnauthorized, err.Error())
		return
	}

	response := struct {
		User  *model.UserResponse `json:"user"`
		Token string              `json:"token"`
	}{
		User:  user,
		Token: token,
	}

	writeSuccessResponse(w, http.StatusOK, "Login successful", response)
}

// GetProfile handles getting user profile
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["id"]

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	user, err := h.userService.GetProfile(r.Context(), userID)
	if err != nil {
		writeErrorResponse(w, http.StatusNotFound, "User not found")
		return
	}

	writeSuccessResponse(w, http.StatusOK, "Profile retrieved successfully", user)
}

// GetMyProfile handles getting current user's profile
func (h *UserHandler) GetMyProfile(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromContext(r.Context())
	if err != nil {
		writeErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, err := h.userService.GetProfile(r.Context(), userID)
	if err != nil {
		writeErrorResponse(w, http.StatusNotFound, "User not found")
		return
	}

	writeSuccessResponse(w, http.StatusOK, "Profile retrieved successfully", user)
}

// UpdateProfile handles updating user profile
func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromContext(r.Context())
	if err != nil {
		writeErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	user, err := h.userService.UpdateProfile(r.Context(), userID, updates)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeSuccessResponse(w, http.StatusOK, "Profile updated successfully", user)
}

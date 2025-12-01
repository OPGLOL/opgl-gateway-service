package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/OPGLOL/opgl-gateway-service/internal/auth"
	apierrors "github.com/OPGLOL/opgl-gateway-service/internal/errors"
	"github.com/OPGLOL/opgl-gateway-service/internal/repository"
	"github.com/google/uuid"
)

// AuthHandler manages authentication HTTP request handlers
type AuthHandler struct {
	authService    *auth.AuthService
	userRepository repository.UserRepository
}

// NewAuthHandler creates a new AuthHandler instance
func NewAuthHandler(authService *auth.AuthService, userRepository repository.UserRepository) *AuthHandler {
	return &AuthHandler{
		authService:    authService,
		userRepository: userRepository,
	}
}

// RegisterRequest represents the request body for user registration
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterResponse represents the response for successful registration
type RegisterResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

// Register handles POST /api/v1/auth/register
func (authHandler *AuthHandler) Register(writer http.ResponseWriter, request *http.Request) {
	var registerRequest RegisterRequest

	if err := json.NewDecoder(request.Body).Decode(&registerRequest); err != nil {
		apierrors.WriteError(writer, apierrors.InvalidRequestBody("Invalid JSON format"))
		return
	}

	// Validate email
	if registerRequest.Email == "" {
		apierrors.WriteError(writer, apierrors.ValidationFailed("email is required"))
		return
	}
	if !strings.Contains(registerRequest.Email, "@") {
		apierrors.WriteError(writer, apierrors.ValidationFailed("invalid email format"))
		return
	}

	// Validate password
	if registerRequest.Password == "" {
		apierrors.WriteError(writer, apierrors.ValidationFailed("password is required"))
		return
	}
	if len(registerRequest.Password) < 8 {
		apierrors.WriteError(writer, apierrors.ValidationFailed("password must be at least 8 characters"))
		return
	}

	// Check if user already exists
	existingUser, err := authHandler.userRepository.GetByEmail(registerRequest.Email)
	if err != nil {
		apierrors.WriteError(writer, apierrors.InternalError("Failed to check existing user"))
		return
	}
	if existingUser != nil {
		apierrors.WriteError(writer, apierrors.NewAPIError(
			apierrors.ErrCodeEmailAlreadyExists,
			"Email already registered",
			http.StatusConflict,
		))
		return
	}

	// Hash password
	passwordHash, err := auth.HashPassword(registerRequest.Password)
	if err != nil {
		apierrors.WriteError(writer, apierrors.InternalError("Failed to hash password"))
		return
	}

	// Create user
	user, err := authHandler.userRepository.Create(registerRequest.Email, passwordHash)
	if err != nil {
		apierrors.WriteError(writer, apierrors.InternalError("Failed to create user"))
		return
	}

	response := RegisterResponse{
		ID:    user.ID.String(),
		Email: user.Email,
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusCreated)
	json.NewEncoder(writer).Encode(response)
}

// LoginRequest represents the request body for user login
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse represents the response for successful login
type LoginResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"`
	User         struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	} `json:"user"`
}

// Login handles POST /api/v1/auth/login
func (authHandler *AuthHandler) Login(writer http.ResponseWriter, request *http.Request) {
	var loginRequest LoginRequest

	if err := json.NewDecoder(request.Body).Decode(&loginRequest); err != nil {
		apierrors.WriteError(writer, apierrors.InvalidRequestBody("Invalid JSON format"))
		return
	}

	// Validate input
	if loginRequest.Email == "" || loginRequest.Password == "" {
		apierrors.WriteError(writer, apierrors.ValidationFailed("email and password are required"))
		return
	}

	// Find user by email
	user, err := authHandler.userRepository.GetByEmail(loginRequest.Email)
	if err != nil {
		apierrors.WriteError(writer, apierrors.InternalError("Failed to find user"))
		return
	}
	if user == nil {
		apierrors.WriteError(writer, apierrors.NewAPIError(
			apierrors.ErrCodeInvalidCredentials,
			"Invalid email or password",
			http.StatusUnauthorized,
		))
		return
	}

	// Verify password
	if !auth.VerifyPassword(loginRequest.Password, user.PasswordHash) {
		apierrors.WriteError(writer, apierrors.NewAPIError(
			apierrors.ErrCodeInvalidCredentials,
			"Invalid email or password",
			http.StatusUnauthorized,
		))
		return
	}

	// Generate tokens
	tokenPair, err := authHandler.authService.GenerateTokenPair(user.ID)
	if err != nil {
		apierrors.WriteError(writer, apierrors.InternalError("Failed to generate tokens"))
		return
	}

	response := LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	}
	response.User.ID = user.ID.String()
	response.User.Email = user.Email

	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(response)
}

// RefreshRequest represents the request body for token refresh
type RefreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

// RefreshResponse represents the response for successful token refresh
type RefreshResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"`
}

// Refresh handles POST /api/v1/auth/refresh
func (authHandler *AuthHandler) Refresh(writer http.ResponseWriter, request *http.Request) {
	var refreshRequest RefreshRequest

	if err := json.NewDecoder(request.Body).Decode(&refreshRequest); err != nil {
		apierrors.WriteError(writer, apierrors.InvalidRequestBody("Invalid JSON format"))
		return
	}

	if refreshRequest.RefreshToken == "" {
		apierrors.WriteError(writer, apierrors.ValidationFailed("refreshToken is required"))
		return
	}

	// Validate refresh token
	userID, err := authHandler.authService.ValidateRefreshToken(refreshRequest.RefreshToken)
	if err != nil {
		apierrors.WriteError(writer, apierrors.NewAPIError(
			apierrors.ErrCodeInvalidToken,
			"Invalid or expired refresh token",
			http.StatusUnauthorized,
		))
		return
	}

	// Verify user still exists
	user, err := authHandler.userRepository.GetByID(userID)
	if err != nil || user == nil {
		apierrors.WriteError(writer, apierrors.NewAPIError(
			apierrors.ErrCodeInvalidToken,
			"User not found",
			http.StatusUnauthorized,
		))
		return
	}

	// Generate new token pair
	tokenPair, err := authHandler.authService.GenerateTokenPair(userID)
	if err != nil {
		apierrors.WriteError(writer, apierrors.InternalError("Failed to generate tokens"))
		return
	}

	response := RefreshResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	}

	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(response)
}

// MeResponse represents the response for getting current user
type MeResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	CreatedAt string `json:"createdAt"`
}

// Me handles POST /api/v1/auth/me (protected endpoint)
func (authHandler *AuthHandler) Me(writer http.ResponseWriter, request *http.Request) {
	// Get user ID from context (set by auth middleware)
	userIDValue := request.Context().Value("userID")
	if userIDValue == nil {
		apierrors.WriteError(writer, apierrors.NewAPIError(
			apierrors.ErrCodeUnauthorized,
			"Unauthorized",
			http.StatusUnauthorized,
		))
		return
	}

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		apierrors.WriteError(writer, apierrors.InternalError("Invalid user context"))
		return
	}

	// Get user from database
	user, err := authHandler.userRepository.GetByID(userID)
	if err != nil {
		apierrors.WriteError(writer, apierrors.InternalError("Failed to get user"))
		return
	}
	if user == nil {
		apierrors.WriteError(writer, apierrors.NewAPIError(
			apierrors.ErrCodeUserNotFound,
			"User not found",
			http.StatusNotFound,
		))
		return
	}

	response := MeResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(response)
}

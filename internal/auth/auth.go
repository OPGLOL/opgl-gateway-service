package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// TokenType distinguishes between access and refresh tokens
type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

// Claims represents the JWT claims structure
type Claims struct {
	UserID    string    `json:"user_id"`
	TokenType TokenType `json:"token_type"`
	jwt.RegisteredClaims
}

// AuthService handles authentication operations
type AuthService struct {
	jwtSecret        []byte
	accessTokenTTL   time.Duration
	refreshTokenTTL  time.Duration
}

// NewAuthService creates a new authentication service
func NewAuthService(jwtSecret string, accessTokenTTL time.Duration, refreshTokenTTL time.Duration) *AuthService {
	return &AuthService{
		jwtSecret:       []byte(jwtSecret),
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}

// TokenPair contains both access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"`
}

// GenerateTokenPair creates both access and refresh tokens for a user
func (authService *AuthService) GenerateTokenPair(userID uuid.UUID) (*TokenPair, error) {
	accessToken, err := authService.generateToken(userID, TokenTypeAccess, authService.accessTokenTTL)
	if err != nil {
		return nil, err
	}

	refreshToken, err := authService.generateToken(userID, TokenTypeRefresh, authService.refreshTokenTTL)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(authService.accessTokenTTL.Seconds()),
	}, nil
}

// generateToken creates a JWT token with the specified type and TTL
func (authService *AuthService) generateToken(userID uuid.UUID, tokenType TokenType, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:    userID.String(),
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "opgl-gateway",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(authService.jwtSecret)
}

// ValidateAccessToken validates an access token and returns the user ID
func (authService *AuthService) ValidateAccessToken(tokenString string) (uuid.UUID, error) {
	return authService.validateToken(tokenString, TokenTypeAccess)
}

// ValidateRefreshToken validates a refresh token and returns the user ID
func (authService *AuthService) ValidateRefreshToken(tokenString string) (uuid.UUID, error) {
	return authService.validateToken(tokenString, TokenTypeRefresh)
}

// validateToken validates a JWT token and returns the user ID
func (authService *AuthService) validateToken(tokenString string, expectedType TokenType) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return authService.jwtSecret, nil
	})

	if err != nil {
		return uuid.Nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return uuid.Nil, errors.New("invalid token")
	}

	// Verify token type
	if claims.TokenType != expectedType {
		return uuid.Nil, errors.New("invalid token type")
	}

	// Parse user ID
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return uuid.Nil, errors.New("invalid user ID in token")
	}

	return userID, nil
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// VerifyPassword checks if the provided password matches the hash
func VerifyPassword(password string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

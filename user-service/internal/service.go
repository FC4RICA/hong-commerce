package internal

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo      *Repository
	jwtSecret []byte
}

func NewService(repo *Repository, jwtSecret string) *Service {
	return &Service{
		repo:      repo,
		jwtSecret: []byte(jwtSecret),
	}
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (*UserResponse, error) {
	return s.createUser(ctx, req, "user")
}

func (s *Service) RegisterAdmin(ctx context.Context, req RegisterRequest) (*UserResponse, error) {
	return s.createUser(ctx, req, "admin")
}

func (s *Service) createUser(ctx context.Context, req RegisterRequest, role string) (*UserResponse, error) {
	if err := validateRegisterRequest(req); err != nil {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.repo.CreateUser(ctx, req.Email, string(hash), req.Name, role)
	if err != nil {
		return nil, err // ErrEmailAlreadyExists passes through as-is
	}

	resp := user.ToResponse()
	return &resp, nil
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			// Don't reveal whether email exists
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := s.issueToken(user)
	if err != nil {
		return nil, fmt.Errorf("issue token: %w", err)
	}

	userResp := user.ToResponse()
	return &LoginResponse{Token: token, User: userResp}, nil
}

func (s *Service) GetMe(ctx context.Context, userID string) (*UserResponse, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	resp := user.ToResponse()
	return &resp, nil
}

// issueToken creates a signed JWT for the given user.
func (s *Service) issueToken(user *User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"email":   user.Email,
		"name":    user.Name,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// Sentinel errors for the service layer.
var ErrInvalidCredentials = errors.New("invalid credentials")

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string { return e.Message }

func validateRegisterRequest(req RegisterRequest) error {
	if req.Email == "" {
		return &ValidationError{Message: "email is required"}
	}
	if len(req.Password) < 8 {
		return &ValidationError{Message: "password must be at least 8 characters"}
	}
	if req.Name == "" {
		return &ValidationError{Message: "name is required"}
	}
	return nil
}

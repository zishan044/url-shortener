package auth

import (
	"context"
	"errors"
	"time"

	"github.com/zishan044/url-shortener/internal/models"
	"github.com/zishan044/url-shortener/internal/validators"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Register(ctx context.Context, username, email, password string) (*models.User, error)
	Login(ctx context.Context, email, password string) (string, error)
}

type service struct {
	repo      Repository
	jwtSecret string
	jwtExpiry time.Duration
}

func NewService(repo Repository, jwtSecret string, jwtExpiry time.Duration) Service {
	return &service{
		repo:      repo,
		jwtSecret: jwtSecret,
		jwtExpiry: jwtExpiry,
	}
}

func (s *service) Register(ctx context.Context, username, email, password string) (*models.User, error) {

	if err := validators.ValidateUsername(username); err != nil {
		return nil, err
	}

	if err := validators.ValidateEmail(email); err != nil {
		return nil, err
	}

	if err := validators.ValidatePassword(password); err != nil {
		return nil, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(hashedPassword),
	}

	err = s.repo.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *service) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", errors.New("invalid email or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", errors.New("invalid email or password")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(s.jwtExpiry).Unix(),
		"iat": time.Now().Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
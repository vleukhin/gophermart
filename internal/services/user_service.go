package services

import (
	"context"
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"

	"github.com/vleukhin/gophermart/internal/storage"
)

type UserService struct {
	storage storage.Storage
	jwtKey  []byte
}

type claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

var ErrUsernameTaken = errors.New("this username is already taken")

func NewUserService(storage storage.Storage, jwtKey string) UserService {
	return UserService{
		storage: storage,
		jwtKey:  []byte(jwtKey),
	}
}

func (s UserService) Register(ctx context.Context, name, password string) (string, time.Time, error) {
	passwordHash, err := s.hashPassword(password)
	if err != nil {
		return "", time.Now(), err
	}

	created, err := s.storage.CreateUser(ctx, name, passwordHash)
	if err != nil {
		return "", time.Now(), err
	}

	if !created {
		return "", time.Now(), ErrUsernameTaken
	}

	log.Debug().Msgf("User %s created", name)

	return s.authorize(name)
}

func (s UserService) Login(ctx context.Context, name, password string) (string, time.Time, error) {
	user, err := s.storage.GetUser(ctx, name)
	if err != nil {
		return "", time.Now(), err
	}

	if user == nil || !s.checkPasswordHash(password, user.Password) {
		return "", time.Now(), nil
	}

	return s.authorize(user.Name)
}

func (s UserService) authorize(name string) (string, time.Time, error) {
	ttl := time.Now().Add(24 * time.Hour)
	claims := &claims{
		Username: name,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: ttl.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtKey)
	if err != nil {
		return "", time.Now(), err
	}

	return tokenString, ttl, nil
}

func (s UserService) hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func (s UserService) checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

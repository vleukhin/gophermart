package services

import (
	"context"
	"errors"

	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"

	"github.com/vleukhin/gophermart/internal/storage"
)

type UserService struct {
	storage storage.Storage
}

var ErrUsernameTaken = errors.New("this username is already taken")

func NewUserService(storage storage.Storage) UserService {
	return UserService{storage: storage}
}

func (s UserService) UserRegister(ctx context.Context, login, password string) error {
	passwordHash, err := s.hashPassword(password)
	if err != nil {
		return err
	}

	created, err := s.storage.CreateUser(ctx, login, passwordHash)
	if err != nil {
		return err
	}

	if !created {
		return ErrUsernameTaken
	}

	log.Debug().Msgf("User %s created", login)

	return nil
}

func (s UserService) hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func (s UserService) checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

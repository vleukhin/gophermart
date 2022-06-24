package services

import (
	"context"
	"errors"

	"github.com/rs/zerolog/log"

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
	created, err := s.storage.CreateUser(ctx, login, password)
	if err != nil {
		return err
	}

	if !created {
		return ErrUsernameTaken
	}

	log.Debug().Msgf("User %s created", login)

	return nil
}

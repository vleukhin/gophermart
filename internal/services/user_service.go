package services

import (
	"errors"

	"github.com/rs/zerolog/log"

	"github.com/vleukhin/gophermart/internal/storage"
)

type UserService struct {
	storage storage.Storage
}

var ErrLUsernameTaken = errors.New("this username is already taken")

func NewUserService(storage storage.Storage) UserService {
	return UserService{storage: storage}
}

func (s UserService) UserRegister(login, password string) error {
	user, err := s.storage.GetUser(login)
	if err != nil {
		return err
	}

	if user != nil {
		return ErrLUsernameTaken
	}

	err = s.storage.CreateUser(login, password)
	if err != nil {
		return err
	}

	log.Debug().Msgf("User %s created", login)

	return nil
}

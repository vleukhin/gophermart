package services

import (
	"github.com/rs/zerolog/log"
	"github.com/vleukhin/gophermart/internal/storage"
)

type UserService struct {
	storage storage.Storage
}

func NewUserService(storage storage.Storage) UserService {
	return UserService{storage: storage}
}

func (s UserService) UserRegister(name, password string) error {
	err := s.storage.CreateUser(name, password)
	if err != nil {
		return err
	}

	log.Debug().Msgf("User %s created", name)

	return nil
}

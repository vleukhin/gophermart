package balance

import (
	"context"
	"sync"

	"github.com/vleukhin/gophermart/internal/storage"
)

type Service struct {
	mutex   sync.Mutex
	storage storage.Storage
}

func NewService(storage storage.Storage) *Service {
	return &Service{
		storage: storage,
	}
}

func (s *Service) Withdraw(ctx context.Context, userID int, amount float32) (bool, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	balance, err := s.storage.GetBalance(ctx, userID)
	if err != nil || balance < amount {
		return false, err
	}

	if err := s.storage.CreateWithdraw(ctx, userID, amount); err != nil {
		return false, err
	}

	return true, nil
}

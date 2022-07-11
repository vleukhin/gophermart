package balance

import (
	"context"
	"sync"

	"github.com/vleukhin/gophermart/internal/storage"
	"github.com/vleukhin/gophermart/internal/types"
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

func (s *Service) Balance(ctx context.Context, userID int) (types.Balance, error) {
	accrual, err := s.storage.GetAccrualSum(ctx, userID)
	if err != nil {
		return types.Balance{}, err
	}
	withdrawn, err := s.storage.GetWithdrawalsSum(ctx, userID)
	if err != nil {
		return types.Balance{}, err
	}

	return types.Balance{
		Current:   accrual - withdrawn,
		Withdrawn: withdrawn,
	}, nil
}

func (s *Service) CreateWithdraw(ctx context.Context, userID int, orderID string, sum float32) (bool, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	accrual, err := s.storage.GetAccrualSum(ctx, userID)
	if err != nil || accrual < sum {
		return false, err
	}

	if err := s.storage.CreateWithdraw(ctx, userID, orderID, sum); err != nil {
		return false, err
	}

	return true, nil
}

func (s *Service) WithdrawalsList(ctx context.Context, userID int) ([]types.Withdraw, error) {
	return s.storage.GetWithdrawals(ctx, userID)
}

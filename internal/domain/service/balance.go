package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/mbiwapa/gophermart.git/internal/domain/entity"
	"github.com/mbiwapa/gophermart.git/internal/lib/contexter"
	"github.com/mbiwapa/gophermart.git/internal/lib/logger"
)

type BalanceRepository interface {
	GetBalance(ctx context.Context, userUUID uuid.UUID) (*entity.Balance, error)
	GetWithdrawOperations(ctx context.Context, userUUID uuid.UUID) ([]entity.BalanceOperation, error)
	Execute(ctx context.Context, operation entity.BalanceOperation) (*entity.Balance, error)
}

type BalanceService struct {
	repository BalanceRepository
	logger     *logger.Logger
}

// FIXME add repository
func NewBalanceService(logger *logger.Logger) *BalanceService {
	return &BalanceService{
		repository: nil,
		logger:     logger,
	}
}

func (s *BalanceService) GetBalance(ctx context.Context, userUUID uuid.UUID) (*entity.Balance, error) {
	const op = "domain.services.BalanceService.GetBalance"
	log := s.logger.With(
		s.logger.StringField("op", op),
		s.logger.StringField("request_id", contexter.GetRequestID(ctx)),
	)

	balance, err := s.repository.GetBalance(ctx, userUUID)
	if err != nil {
		log.Error("Failed to get balance", log.ErrorField(err))
		return nil, err
	}

	log.Info("Balance retrieved")
	return balance, nil
}

func (s *BalanceService) Execute(ctx context.Context, operation entity.BalanceOperation) (*entity.Balance, error) {
	const op = "domain.services.BalanceService.Execute"
	log := s.logger.With(
		s.logger.StringField("op", op),
		s.logger.StringField("request_id", contexter.GetRequestID(ctx)),
	)

	balance, err := s.repository.Execute(ctx, operation)
	if err != nil {
		log.Error("Failed to get balance", log.ErrorField(err))
		return nil, err
	}
	// FIXME проверка что есть деньги для списания если это списание, просто начислить если это начисление

	log.Info("Balance updated")
	return balance, nil
}

func (s *BalanceService) GetWithdrawOperations(ctx context.Context, userUUID uuid.UUID) ([]entity.BalanceOperation, error) {
	const op = "domain.services.BalanceService.GetOperations"
	log := s.logger.With(
		s.logger.StringField("op", op),
		s.logger.StringField("request_id", contexter.GetRequestID(ctx)),
	)

	operations, err := s.repository.GetWithdrawOperations(ctx, userUUID)
	if err != nil {
		log.Error("Failed to get operations", log.ErrorField(err))
		return nil, err
	}

	log.Info("Operations retrieved")
	return operations, nil
}

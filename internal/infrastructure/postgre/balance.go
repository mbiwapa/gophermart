package postgre

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mbiwapa/gophermart.git/internal/domain/entity"
	"github.com/mbiwapa/gophermart.git/internal/lib/contexter"
	"github.com/mbiwapa/gophermart.git/internal/lib/logger"
)

// BalanceRepository is structure for repository for balance service
type BalanceRepository struct {
	db  *pgxpool.Pool
	log *logger.Logger
}

// NewBalanceRepository returns a new BalanceRepository
func NewBalanceRepository(ctx context.Context, db *pgxpool.Pool, log *logger.Logger) (*BalanceRepository, error) {
	const op = "infrastructure.postgre.NewBalanceRepository"
	logWith := log.With(log.StringField("op", op))

	storage := &BalanceRepository{db: db, log: log}

	_, err := db.Exec(ctx, `CREATE TABLE IF NOT EXISTS balances (
        user_uuid uuid PRIMARY KEY NOT NULL,
        current INTEGER NOT NULL,
        withdraw INTEGER NOT NULL)`)
	if err != nil {
		logWith.Error("Failed to create table balances", log.ErrorField(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	_, err = db.Exec(ctx, `CREATE TABLE IF NOT EXISTS balance_operations (
        uuid uuid PRIMARY KEY NOT NULL,
        user_uuid uuid NOT NULL,
        accrual INTEGER NOT NULL,
        withdrawal INTEGER NOT NULL,
        order_number INTEGER NOT NULL,
        processed_at TIMESTAMP NOT NULL)`)
	if err != nil {
		logWith.Error("Failed to create table balance_operations", log.ErrorField(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return storage, nil
}

// GetBalance returns the balance of the user
func (r *BalanceRepository) GetBalance(ctx context.Context, userUUID uuid.UUID) (*entity.Balance, error) {
	const op = "infrastructure.postgre.BalanceRepository.GetBalance"
	log := r.log.With(
		r.log.StringField("op", op),
		r.log.StringField("request_id", contexter.GetRequestID(ctx)),
		r.log.StringField("user_uuid", userUUID.String()),
	)
	var balance entity.Balance
	err := r.db.QueryRow(ctx, `SELECT current, withdraw FROM balances WHERE user_uuid = $1`, userUUID).Scan(&balance.Current, &balance.Withdraw)
	if err != nil {
		log.Error("Failed to get balance", log.ErrorField(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &balance, nil
}

// GetWithdrawOperations returns a list of withdrawal operations
func (r *BalanceRepository) GetWithdrawOperations(ctx context.Context, userUUID uuid.UUID) ([]entity.BalanceOperation, error) {
	const op = "infrastructure.postgre.BalanceRepository.GetWithdrawOperations"
	log := r.log.With(
		r.log.StringField("op", op),
		r.log.StringField("request_id", contexter.GetRequestID(ctx)),
		r.log.StringField("user_uuid", userUUID.String()),
	)
	var operations []entity.BalanceOperation
	result, err := r.db.Query(ctx, `SELECT 
											uuid, 
											user_uuid, 
											accrual, 
											withdrawal, 
											order_number, 
											processed_at 
										FROM 
										    balance_operations 
										WHERE 
										    user_uuid = $1 
										  AND 
										    withdrawal > 0`)
	if err != nil {
		log.Error("Failed to get withdraw operations", log.ErrorField(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer result.Close()
	for result.Next() {
		var operation entity.BalanceOperation
		err = result.Scan(&operation.UUID, &operation.UserUUID, &operation.Accrual, &operation.Withdrawal, &operation.OrderNumber, &operation.ProcessedAt)
		if err != nil {
			log.Error("Failed to scan row", log.ErrorField(err))
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		operations = append(operations, operation)
	}
	if len(operations) == 0 {
		log.Info("No withdraw operations found")
		return nil, entity.ErrBalanceOperationsNotFound
	}
	return operations, nil
}

// CreateBalance creates a new balance for the user
func (r *BalanceRepository) CreateBalance(ctx context.Context, userUUID uuid.UUID) error {
	const op = "infrastructure.postgre.BalanceRepository.CreateBalance"
	log := r.log.With(
		r.log.StringField("op", op),
		r.log.StringField("request_id", contexter.GetRequestID(ctx)),
		r.log.StringField("user_uuid", userUUID.String()),
	)
	_, err := r.db.Exec(ctx, `INSERT INTO balances (user_uuid, current, withdraw) VALUES ($1, 0, 0)`, userUUID)
	if err != nil {
		log.Error("Failed to create balance", log.ErrorField(err))
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// Withdraw executes a balance operation
func (r *BalanceRepository) Withdraw(ctx context.Context, operation entity.BalanceOperation) error {
	const op = "infrastructure.postgre.BalanceRepository.Withdraw"
	log := r.log.With(
		r.log.StringField("op", op),
		r.log.StringField("request_id", contexter.GetRequestID(ctx)),
		r.log.StringField("user_uuid", operation.UserUUID.String()),
		r.log.AnyField("order_number", operation.OrderNumber),
		r.log.AnyField("accrual", operation.Accrual),
		r.log.AnyField("withdrawal", operation.Withdrawal),
	)

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		log.Error("Failed to begin transaction", log.ErrorField(err))
		return fmt.Errorf("%s: %w", op, err)
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		err := tx.Rollback(ctx)
		if err != nil {
			log.Error("Failed to rollback transaction", log.ErrorField(err))
		}
	}(tx, ctx)

	var current float64
	balance := tx.QueryRow(ctx, `SELECT current FROM balances WHERE user_uuid = $1 FOR UPDATE`, operation.UserUUID)
	err = balance.Scan(&current)
	if err != nil {
		log.Error("Failed to get current balance", log.ErrorField(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	if current < operation.Withdrawal {
		log.Error("Insufficient funds in the account")
		return entity.ErrBalanceInsufficientFunds
	}

	_, err = tx.Exec(ctx, `UPDATE balances SET current = current - $1 WHERE user_uuid = $2`, operation.Withdrawal, operation.UserUUID)
	if err != nil {
		log.Error("Failed to update balance", log.ErrorField(err))
		return fmt.Errorf("%s: %w", op, err)
	}
	_, err = tx.Exec(ctx, `INSERT INTO balance_operations (
                                user_uuid, 
                                accrual, 
                                withdrawal, 
                                order_number, 
                                processed_at, 
                                uuid
                                ) VALUES ($1, $2, $3, $4, $5, $6)`,
		operation.UserUUID,
		operation.Accrual,
		operation.Withdrawal,
		operation.OrderNumber,
		operation.ProcessedAt,
		operation.UUID)
	if err != nil {
		log.Error("Failed to create balance operation", log.ErrorField(err))
		return fmt.Errorf("%s: %w", op, err)
	}
	err = tx.Commit(ctx)
	if err != nil {
		log.Error("Failed to commit transaction", log.ErrorField(err))
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *BalanceRepository) Accrue(ctx context.Context, operation entity.BalanceOperation) error {
	const op = "infrastructure.postgre.BalanceRepository.Accrue"
	log := r.log.With(
		r.log.StringField("op", op),
		r.log.StringField("request_id", contexter.GetRequestID(ctx)),
		r.log.StringField("user_uuid", operation.UserUUID.String()),
		r.log.AnyField("order_number", operation.OrderNumber),
		r.log.AnyField("accrual", operation.Accrual),
		r.log.AnyField("withdrawal", operation.Withdrawal),
	)

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		log.Error("Failed to begin transaction", log.ErrorField(err))
		return fmt.Errorf("%s: %w", op, err)
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		err := tx.Rollback(ctx)
		if err != nil {
			log.Error("Failed to rollback transaction", log.ErrorField(err))
		}
	}(tx, ctx)

	var current float64
	balance := tx.QueryRow(ctx, `SELECT current FROM balances WHERE user_uuid = $1 FOR UPDATE`, operation.UserUUID)
	err = balance.Scan(&current)
	if err != nil {
		log.Error("Failed to get current balance", log.ErrorField(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = tx.Exec(ctx, `UPDATE balances SET current = current + $1 WHERE user_uuid = $2`, operation.Accrual, operation.UserUUID)
	if err != nil {
		log.Error("Failed to update balance", log.ErrorField(err))
		return fmt.Errorf("%s: %w", op, err)
	}
	_, err = tx.Exec(ctx, `INSERT INTO balance_operations (
                                user_uuid, 
                                accrual, 
                                withdrawal, 
                                order_number, 
                                processed_at, 
                                uuid
                                ) VALUES ($1, $2, $3, $4, $5, $6)`,
		operation.UserUUID,
		operation.Accrual,
		operation.Withdrawal,
		operation.OrderNumber,
		operation.ProcessedAt,
		operation.UUID)
	if err != nil {
		log.Error("Failed to create balance operation", log.ErrorField(err))
		return fmt.Errorf("%s: %w", op, err)
	}
	err = tx.Commit(ctx)
	if err != nil {
		log.Error("Failed to commit transaction", log.ErrorField(err))
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

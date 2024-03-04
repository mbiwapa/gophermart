package workers

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mbiwapa/gophermart.git/internal/domain/entity"
	"github.com/mbiwapa/gophermart.git/internal/domain/service"
	"github.com/mbiwapa/gophermart.git/internal/infrastructure/postgre"
	"github.com/mbiwapa/gophermart.git/internal/lib/contexter"
	"github.com/mbiwapa/gophermart.git/internal/lib/logger"
)

type BalanceWorker struct {
	logger         *logger.Logger
	errorChan      chan error
	ctx            context.Context
	balanceQueue   chan entity.BalanceOperation
	balanceService *service.BalanceService
	db             *pgxpool.Pool
}

func NewBalanceWorker(ctx context.Context, logger *logger.Logger, balanceQueue chan entity.BalanceOperation, errorChanel chan error, db *pgxpool.Pool) *BalanceWorker {
	return &BalanceWorker{
		balanceQueue: balanceQueue,
		logger:       logger,
		errorChan:    errorChanel,
		ctx:          ctx,
		db:           db,
	}
}

func (w *BalanceWorker) Run() {
	const op = "app.workers.BalanceWorker.Run"
	log := w.logger.With(w.logger.StringField("op", op))

	balanceRepository, err := postgre.NewBalanceRepository(w.ctx, w.db, w.logger)
	if err != nil {
		log.Error("Failed to create balance repository", log.ErrorField(err))
		os.Exit(1)
	}
	w.balanceService = service.NewBalanceService(w.logger, balanceRepository)

	for i := 1; i <= 3; i++ {
		go w.worker()
	}
	log.Info("Star 3 balance workers")
}

// worker is a goroutine that is responsible for processing balance.
func (w *BalanceWorker) worker() {
	const op = "app.workers.balance"
	log := w.logger.With(w.logger.StringField("op", op))
	for {
		select {
		case <-w.ctx.Done():
			return
		case operation, ok := <-w.balanceQueue:
			if !ok {
				w.errorChan <- fmt.Errorf("balance queue is closed")
				return
			}
			//TODO что-то получше сделать
			reqID := "req_order" + fmt.Sprintf("%d", operation.OrderNumber)
			ctx := context.WithValue(w.ctx, contexter.RequestID, reqID)

			err := w.balanceService.Execute(ctx, operation)
			log.Info("Update balance", log.ErrorField(err))
			if err != nil {
				w.errorChan <- fmt.Errorf("%s: %w", op, err)
			}
		default:
			time.Sleep(1 * time.Second)
		}
	}
}

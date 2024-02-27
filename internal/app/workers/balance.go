package workers

import (
	"context"
	"fmt"

	"github.com/mbiwapa/gophermart.git/internal/domain/entity"
	"github.com/mbiwapa/gophermart.git/internal/domain/service"
	"github.com/mbiwapa/gophermart.git/internal/lib/logger"
)

type BalanceWorker struct {
	logger         *logger.Logger
	errorChan      chan error
	ctx            context.Context
	balanceQueue   chan entity.BalanceOperation //FIXME Operation entity
	balanceService *service.BalanceService
}

func NewBalanceWorker(ctx context.Context, logger *logger.Logger, balanceQueue chan entity.BalanceOperation, errorChanel chan error) *BalanceWorker {
	return &BalanceWorker{
		balanceQueue: balanceQueue,
		logger:       logger,
		errorChan:    errorChanel,
		ctx:          ctx,
	}
}

func (w *BalanceWorker) Run() {
	const op = "workers.BalanceWorker.Run"
	log := w.logger.With(w.logger.StringField("op", op))

	//FIXME добавить репозиторий заказов
	w.balanceService = service.NewBalanceService(w.logger)

	for i := 1; i <= 20; i++ {
		go w.worker()
	}
	log.Info("Star 20 balance workers")
}

// Worker is a worker that sends orders to the order service
func (w *BalanceWorker) worker() {
	const op = "workers.worker"
	log := w.logger.With(w.logger.StringField("op", op))
	for {
		select {
		case <-w.ctx.Done():
			//FIXME add log
			log.Info("END")
			return
		case operation, ok := <-w.balanceQueue:
			if !ok {
				//Error example
				w.errorChan <- fmt.Errorf("balance queue is closed")
				return
			}
			//FIXME ДОбавить всю логику
			balance, err := w.balanceService.Execute(w.ctx, operation)
			if err != nil {
				_ = balance
				w.errorChan <- fmt.Errorf("%s: %w", op, err)
				//return
			}
		default:
		}
	}
}

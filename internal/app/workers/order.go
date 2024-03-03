package workers

import (
	"context"
	"fmt"

	"github.com/mbiwapa/gophermart.git/internal/domain/entity"
	"github.com/mbiwapa/gophermart.git/internal/domain/service"
	"github.com/mbiwapa/gophermart.git/internal/lib/logger"
)

type OrderWorker struct {
	orderService *service.OrderService
	orderQueue   chan entity.Order
	logger       *logger.Logger
	errorChan    chan error
	ctx          context.Context
	balanceQueue chan entity.BalanceOperation //FIXME Operation entity
}

func NewOrderWorker(ctx context.Context, logger *logger.Logger, orderQueue chan entity.Order, errorChanel chan error, balanceQueue chan entity.BalanceOperation) *OrderWorker {
	return &OrderWorker{
		orderQueue:   orderQueue,
		balanceQueue: balanceQueue,
		logger:       logger,
		errorChan:    errorChanel,
		ctx:          ctx,
	}
}

func (w *OrderWorker) Run() {
	const op = "workers.OrderWorker.Run"
	log := w.logger.With(w.logger.StringField("op", op))

	//FIXME добавить репозиторий заказов
	w.orderService = service.NewOrderService(w.logger, w.orderQueue)

	for i := 1; i <= 20; i++ {
		go w.worker()
	}
	log.Info("Star 20 order workers")
}

// Worker is a worker that sends orders to the order service
func (w *OrderWorker) worker() {
	const op = "workers.worker"
	log := w.logger.With(w.logger.StringField("op", op))
	for {
		select {
		case <-w.ctx.Done():
			//FIXME add log
			log.Info("END")
			return
		case order, ok := <-w.orderQueue:
			if !ok {
				//Error example
				w.errorChan <- fmt.Errorf("order queue is closed")
				return
			}
			//FIXME ДОбавить всю логику, сон, ретрай и тд. РЕШИТЬ ОТКУДА ДОБЫТЬ request_id(что-то другое для воркера)
			order, err := w.orderService.Check(w.ctx, order)
			if err != nil {
				w.errorChan <- fmt.Errorf("%s: %w", op, err)
				//return
			}
			//FIXME проверить все
			err = w.orderService.Update(w.ctx, order)
			if err != nil {
				w.errorChan <- fmt.Errorf("%s: %w", op, err)
				//return
			}
			//FIXME добавить логику и запись бонусов в очередь на начисление бонусов(в случае если есть начисление)
			w.balanceQueue <- entity.BalanceOperation{
				Accrual:     order.Accrual,
				UserUUID:    order.UserUUID,
				OrderNumber: order.Number,
			}
		default:
		}
	}
}

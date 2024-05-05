package workers

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mbiwapa/gophermart.git/internal/domain/entity"
	"github.com/mbiwapa/gophermart.git/internal/domain/service"
	httpc "github.com/mbiwapa/gophermart.git/internal/infrastructure/http-client"
	"github.com/mbiwapa/gophermart.git/internal/infrastructure/postgre"
	"github.com/mbiwapa/gophermart.git/internal/lib/contexter"
	"github.com/mbiwapa/gophermart.git/internal/lib/logger"
)

type OrderWorker struct {
	orderService *service.OrderService
	orderQueue   chan entity.Order
	logger       *logger.Logger
	errorChan    chan error
	ctx          context.Context
	balanceQueue chan entity.BalanceOperation
	db           *pgxpool.Pool
	accrualURL   string
}

func NewOrderWorker(ctx context.Context, logger *logger.Logger, orderQueue chan entity.Order, errorChanel chan error, balanceQueue chan entity.BalanceOperation, db *pgxpool.Pool, accrualURL string) *OrderWorker {
	return &OrderWorker{
		orderQueue:   orderQueue,
		balanceQueue: balanceQueue,
		logger:       logger,
		errorChan:    errorChanel,
		ctx:          ctx,
		db:           db,
		accrualURL:   accrualURL,
	}
}

func (w *OrderWorker) Run() {
	const op = "app.workers.OrderWorker.Run"
	log := w.logger.With(w.logger.StringField("op", op))

	client, err := httpc.NewOrderClient(w.accrualURL, w.logger)
	if err != nil {
		log.Error("Failed to create order client", log.ErrorField(err))
		os.Exit(1)
	}

	orderRepository := postgre.NewOrderRepository(w.db, w.logger)
	w.orderService = service.NewOrderService(w.logger, w.orderQueue, orderRepository)
	w.orderService.SetClient(client)

	for i := 1; i <= 3; i++ {
		go w.worker()
	}
	log.Info("Start 3 order workers")
}

// worker is a goroutine that is responsible for processing orders.
func (w *OrderWorker) worker() {
	const op = "app.workers.order"
	log := w.logger.With(w.logger.StringField("op", op))
	for {
		select {
		case <-w.ctx.Done():
			return
		case order, ok := <-w.orderQueue:
			if !ok {
				w.errorChan <- fmt.Errorf("order queue is closed")
				return
			}
			//TODO что-то получше сделать
			reqID := "req_order" + fmt.Sprintf("%d", order.Number)
			ctx := context.WithValue(w.ctx, contexter.RequestID, reqID)

			log.Info("Processing order", log.AnyField("order_number", order.Number), log)
			bonuses, err := w.orderService.Check(ctx, order)
			if err != nil {
				w.errorChan <- fmt.Errorf("%s: %w", op, err)
				return
			}

			log.Info("Order processed",
				log.AnyField("order_number", order.Number),
				log.AnyField("bonuses", bonuses),
			)
			if bonuses > 0 {
				w.balanceQueue <- entity.NewBalanceOperation(order.UserUUID, bonuses, 0, order.Number)
			}
		}
	}
}

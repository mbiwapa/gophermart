package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/mbiwapa/gophermart.git/internal/domain/entity"
	"github.com/mbiwapa/gophermart.git/internal/lib/contexter"
	"github.com/mbiwapa/gophermart.git/internal/lib/logger"
)

// OrderRepository is an interface for orders repository.
type OrderRepository interface {
	AddOrderForUser(ctx context.Context, order entity.Order) error
	GetAllUserOrders(ctx context.Context, userUUID uuid.UUID) ([]entity.Order, error)
	UpdateOrderForUser(ctx context.Context, order entity.Order) error
}

type OrderClient interface {
	Check(ctx context.Context, number string) (entity.Order, error)
}

// OrderService is a service for managing orders.
type OrderService struct {
	repository OrderRepository
	logger     *logger.Logger
	client     OrderClient
	orderQueue chan entity.Order
}

// NewOrderService returns a new order service.
func NewOrderService(logger *logger.Logger, orderQueue chan entity.Order) *OrderService {
	// FIXME add Repository and OrderClient
	return &OrderService{
		repository: nil,
		logger:     logger,
		orderQueue: orderQueue,
	}
}

// Add adds a new order for a user.
func (s *OrderService) Add(ctx context.Context, orderNumber string, userUUID uuid.UUID) error {
	const op = "domain.services.OrderService.Add"
	log := s.logger.With(
		s.logger.StringField("op", op),
		s.logger.StringField("request_id", contexter.GetRequestID(ctx)),
	)
	//FIXME add validation and error handling and etc
	order := entity.NewOrder(userUUID, orderNumber)

	err := s.repository.AddOrderForUser(ctx, order)
	if err != nil {
		log.Error("Failed to add order", log.ErrorField(err))
		return err
	}
	//Add order to queue channel
	s.orderQueue <- order

	log.Info("Order added")
	return nil
}

// GetAll returns all orders for a user.
func (s *OrderService) GetAll(ctx context.Context, userUUID uuid.UUID) ([]entity.Order, error) {
	const op = "domain.services.OrderService.GetAll"
	log := s.logger.With(
		s.logger.StringField("op", op),
		s.logger.StringField("request_id", contexter.GetRequestID(ctx)),
	)
	//FIXME add validation and error handling and etc

	orders, err := s.repository.GetAllUserOrders(ctx, userUUID)
	if err != nil {
		log.Error("Failed to get orders", log.ErrorField(err))
		return nil, err
	}

	log.Info("Orders retrieved")
	return orders, nil
}

// Check return count bonuses for the order
func (s *OrderService) Check(ctx context.Context, order entity.Order) (entity.Order, error) {
	const op = "domain.services.OrderService.Check"
	log := s.logger.With(
		s.logger.StringField("op", op),
		s.logger.StringField("request_id", contexter.GetRequestID(ctx)),
	)
	order, err := s.client.Check(ctx, order.Number)
	if err != nil {
		_ = order
		//FIXME добавить обработку ошибок, и тд
		log.Error("Failed to get order", log.ErrorField(err))
		return order, err
	}
	//FIXME сделать запись правильную в канал по добавлению баланса пользователю.
	return order, nil
}

func (s *OrderService) Update(ctx context.Context, order entity.Order) error {
	const op = "domain.services.OrderService.Update"
	log := s.logger.With(
		s.logger.StringField("op", op),
		s.logger.StringField("request_id", contexter.GetRequestID(ctx)),
	)
	//FIXME добавить проверки и тд и тп

	err := s.repository.UpdateOrderForUser(ctx, order)
	if err != nil {
		log.Error("Failed to update order", log.ErrorField(err))
		return err
	}

	log.Info("Order updated")
	return nil
}

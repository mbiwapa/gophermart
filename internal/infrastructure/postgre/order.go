package postgre

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mbiwapa/gophermart.git/internal/domain/entity"
	"github.com/mbiwapa/gophermart.git/internal/lib/contexter"
	"github.com/mbiwapa/gophermart.git/internal/lib/logger"
)

// OrderRepository is an implementation of order repository.
type OrderRepository struct {
	db  *pgxpool.Pool
	log *logger.Logger
}

// NewOrderRepository returns a new postgre user repository
func NewOrderRepository(ctx context.Context, db *pgxpool.Pool, log *logger.Logger) (*OrderRepository, error) {
	const op = "infrastructure.postgre.NewOrderRepository"
	logWith := log.With(log.StringField("op", op))

	storage := &OrderRepository{db: db, log: log}

	_, err := db.Exec(ctx, `CREATE TABLE IF NOT EXISTS orders (
    	user_uuid uuid NOT NULL,
        number INTEGER PRIMARY KEY,
        status TEXT NOT NULL,
        accrual FLOAT NOT NULL DEFAULT 0,
        uploaded_at TIMESTAMP NOT NULL);`)
	if err != nil {
		logWith.Error("Failed to create table", log.ErrorField(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return storage, nil
}

// AddOrderForUser adds a new order to the user.
func (r *OrderRepository) AddOrderForUser(ctx context.Context, order entity.Order) error {
	const op = "infrastructure.postgre.OrderRepository.AddOrderForUser"

	log := r.log.With(r.log.StringField("op", op),
		r.log.StringField("request_id", contexter.GetRequestID(ctx)),
		r.log.AnyField("order_number", order.Number),
		r.log.AnyField("user_uuid", order.UserUUID),
	)

	_, err := r.db.Exec(ctx, `INSERT INTO orders (
                    user_uuid,
                    number,
                    status,
                    uploaded_at,
                    accrual
                    ) VALUES ($1, $2, $3, $4, 0)`, order.UserUUID, order.Number, order.Status, order.UploadedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				var dbUserUUid uuid.UUID
				err = r.db.QueryRow(ctx, `SELECT user_uuid FROM orders WHERE number = $1`, order.Number).Scan(&dbUserUUid)
				if err != nil {
					log.Info("Unknown error", log.ErrorField(err))
					return fmt.Errorf("%s: %w", op, err)
				}
				if dbUserUUid == order.UserUUID {
					log.Info("Order already uploaded from current user")
					return entity.ErrOrderAlreadyUploaded
				} else {
					log.Info("Order already uploaded from another user")
					return entity.ErrOrderAlreadyUploadedByAnotherUser
				}
			}
		}
		log.Error("Failed to create order", log.ErrorField(err))
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// GetAllUserOrders returns all orders for a user.
func (r *OrderRepository) GetAllUserOrders(ctx context.Context, userUUID uuid.UUID) ([]entity.Order, error) {
	const op = "infrastructure.postgre.OrderRepository.GetAllUserOrders"
	log := r.log.With(r.log.StringField("op", op),
		r.log.StringField("request_id", contexter.GetRequestID(ctx)),
		r.log.StringField("user_uuid", userUUID.String()),
	)
	rows, err := r.db.Query(ctx, `SELECT * FROM orders WHERE user_uuid = $1`, userUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Info("No orders found")
			return nil, entity.ErrOrderNotFound
		}
		log.Error("Failed to get orders", log.ErrorField(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()
	var orders []entity.Order
	for rows.Next() {
		var order entity.Order
		err = rows.Scan(&order.UserUUID, &order.Number, &order.Status, &order.UploadedAt, &order.Accrual)
		if err != nil {
			log.Error("Failed to scan row", log.ErrorField(err))
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		orders = append(orders, order)
	}
	return orders, nil
}

func (r *OrderRepository) UpdateOrderForUser(ctx context.Context, order entity.Order) error {
	const op = "infrastructure.postgre.OrderRepository.UpdateOrderForUser"
	log := r.log.With(r.log.StringField("op", op),
		r.log.StringField("request_id", contexter.GetRequestID(ctx)),
		r.log.AnyField("order_number", order.Number),
		r.log.AnyField("user_uuid", order.UserUUID),
	)
	_, err := r.db.Exec(ctx, `UPDATE orders SET 
                  status = $1,
                  accrual = $2
              WHERE
                  user_uuid = $3 
                AND
                  number = $4`, order.Status, order.Accrual, order.UserUUID, order.Number)
	if err != nil {
		log.Error("Failed to update order", log.ErrorField(err))
	}
	return nil
}
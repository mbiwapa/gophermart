package httpc

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/mbiwapa/gophermart.git/internal/domain/entity"
	"github.com/mbiwapa/gophermart.git/internal/lib/logger"
)

// OrderClient структура возвращаемая для работы, клиент
type OrderClient struct {
	url    string
	client *http.Client
	logger *logger.Logger
}

// NewOrderClient возвращает экземпляр клиента
func NewOrderClient(url string, logger *logger.Logger) (*OrderClient, error) {
	var client OrderClient
	client.url = url
	client.client = &http.Client{
		Transport: &http.Transport{},
	}
	client.logger = logger
	return &client, nil
}

// get отправляет запрос к указанному адресу и возвращает ответ
func (c *OrderClient) get(ctx context.Context, path string) ([]byte, error) {
	const op = "http-client.send.get"
	log := c.logger.With(c.logger.StringField("op", op))

	req, err := http.NewRequestWithContext(ctx, "GET", c.url+path, nil)
	if err != nil {
		log.Error("Cant create request", log.ErrorField(err))
		return nil, err
	}
	req.Close = true // Close the connection after sending the request

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		log.Error("Failed to send request", log.ErrorField(err))
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Error("No response", log.AnyField("code", resp.StatusCode))
		return nil, err
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("Cant  read response", log.ErrorField(err))
		return nil, err
	}
	log.Info("Request completed successfully!")
	return bodyBytes, nil

}

// Check возвращает информацию о заказе по номеру
func (c *OrderClient) Check(ctx context.Context, number string) (entity.Order, error) {
	const op = "http-client.send.GetOrderInfo"
	log := c.logger.With(c.logger.StringField("op", op))

	path := "/api/orders/" + number
	bodyBytes, err := c.get(ctx, path)
	if err != nil {
		return entity.Order{}, err
	}

	var order entity.Order
	err = json.Unmarshal(bodyBytes, &order)
	if err != nil {
		log.Error("Cant unmarshal response", log.ErrorField(err))
		return order, err
	}
	return order, nil
}

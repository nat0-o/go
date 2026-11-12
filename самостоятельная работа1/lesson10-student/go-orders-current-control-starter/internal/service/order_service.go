package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"example.com/order-service/internal/order"
)

type OrderService struct {
	repository order.Repository
}

func NewOrderService(repository order.Repository) *OrderService {
	return &OrderService{repository: repository}
}

func (s *OrderService) Create(ctx context.Context, input order.CreateOrderInput) (order.Order, error) {
	if len(input.Items) == 0 {
		return order.Order{}, fmt.Errorf("%w: empty items", order.ErrInvalidOrder)
	}

	var totalPrice int64
	for _, item := range input.Items {
		if strings.TrimSpace(item.Name) == "" {
			return order.Order{}, fmt.Errorf("%w: empty name", order.ErrInvalidOrder)
		}
		if item.Quantity <= 0 {
			return order.Order{}, fmt.Errorf("%w: invalid quantity", order.ErrInvalidOrder)
		}
		if item.Price <= 0 {
			return order.Order{}, fmt.Errorf("%w: invalid price", order.ErrInvalidOrder)
		}
		totalPrice += item.Price * int64(item.Quantity)
	}

	now := time.Now()
	newOrder := order.Order{
		Items:      input.Items,
		TotalPrice: totalPrice,
		Status:     order.StatusNew,
		CreatedAt:  now,
		History: []order.StatusChange{
			{
				Status:    order.StatusNew,
				ChangedAt: now,
			},
		},
	}

	return s.repository.Create(ctx, newOrder)
}

func (s *OrderService) GetByID(ctx context.Context, id int64) (order.Order, error) {
	return s.repository.GetByID(ctx, id)
}

func (s *OrderService) List(ctx context.Context) ([]order.Order, error) {
	return s.repository.List(ctx)
}

func (s *OrderService) UpdateStatus(ctx context.Context, id int64, status order.Status) (order.Order, error) {
	if status != order.StatusNew && status != order.StatusProcessing && status != order.StatusCompleted && status != order.StatusCancelled {
		return order.Order{}, fmt.Errorf("%w: unknown status", order.ErrInvalidStatus)
	}

	o, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return order.Order{}, err
	}

	if !checkTransition(o.Status, status) {
		return order.Order{}, fmt.Errorf("%w: %s to %s", order.ErrInvalidStatusTransition, o.Status, status)
	}

	o.Status = status
	o.History = append(o.History, order.StatusChange{
		Status:    status,
		ChangedAt: time.Now(),
	})

	return s.repository.Update(ctx, o)
}

func (s *OrderService) History(ctx context.Context, id int64) ([]order.StatusChange, error) {
	o, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return o.History, nil
}

func checkTransition(oldStatus, newStatus order.Status) bool {
	switch oldStatus {
	case order.StatusNew:
		return newStatus == order.StatusProcessing || newStatus == order.StatusCancelled
	case order.StatusProcessing:
		return newStatus == order.StatusCompleted || newStatus == order.StatusCancelled
	default:
		return false
	}
}

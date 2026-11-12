package memory

import (
	"context"

	"example.com/order-service/internal/order"
)

type OrderRepository struct {
	items  []order.Order
	nextID int64
}

func NewOrderRepository() *OrderRepository {
	return &OrderRepository{
		items:  make([]order.Order, 0),
		nextID: 1,
	}
}

func (r *OrderRepository) Create(_ context.Context, value order.Order) (order.Order, error) {
	value.ID = r.nextID
	r.nextID++

	r.items = append(r.items, copyOrder(value))
	return copyOrder(value), nil
}

func (r *OrderRepository) GetByID(_ context.Context, id int64) (order.Order, error) {
	for _, o := range r.items {
		if o.ID == id {
			return copyOrder(o), nil
		}
	}
	return order.Order{}, order.ErrOrderNotFound
}

func (r *OrderRepository) List(_ context.Context) ([]order.Order, error) {
	res := make([]order.Order, 0, len(r.items))
	for _, o := range r.items {
		res = append(res, copyOrder(o))
	}
	return res, nil
}

func (r *OrderRepository) Update(_ context.Context, value order.Order) (order.Order, error) {
	for i, o := range r.items {
		if o.ID == value.ID {
			r.items[i] = copyOrder(value)
			return copyOrder(value), nil
		}
	}
	return order.Order{}, order.ErrOrderNotFound
}

func copyOrder(o order.Order) order.Order {
	items := make([]order.Item, len(o.Items))
	copy(items, o.Items)
	o.Items = items

	history := make([]order.StatusChange, len(o.History))
	copy(history, o.History)
	o.History = history

	return o
}

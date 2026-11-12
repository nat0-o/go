package service_test

import (
	"context"
	"testing"
	"time"

	"example.com/order-service/internal/order"
	"example.com/order-service/internal/service"
	"github.com/stretchr/testify/assert"
)

type fakeRepo struct {
	orders []order.Order
	nextID int64
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		orders: make([]order.Order, 0),
		nextID: 1,
	}
}

func (f *fakeRepo) Create(_ context.Context, value order.Order) (order.Order, error) {
	value.ID = f.nextID
	f.nextID++
	f.orders = append(f.orders, value)
	return value, nil
}

func (f *fakeRepo) GetByID(_ context.Context, id int64) (order.Order, error) {
	for _, o := range f.orders {
		if o.ID == id {
			return o, nil
		}
	}
	return order.Order{}, order.ErrOrderNotFound
}

func (f *fakeRepo) List(_ context.Context) ([]order.Order, error) {
	return f.orders, nil
}

func (f *fakeRepo) Update(_ context.Context, value order.Order) (order.Order, error) {
	for i, o := range f.orders {
		if o.ID == value.ID {
			f.orders[i] = value
			return value, nil
		}
	}
	return order.Order{}, order.ErrOrderNotFound
}

func TestOrderService_Create(t *testing.T) {
	tests := []struct {
		name      string
		input     order.CreateOrderInput
		wantTotal int64
		wantErr   error
	}{
		{
			name: "ok",
			input: order.CreateOrderInput{
				Items: []order.Item{
					{Name: "Keyboard", Quantity: 2, Price: 1000},
					{Name: "Mouse", Quantity: 1, Price: 500},
				},
			},
			wantTotal: 2500,
			wantErr:   nil,
		},
		{
			name:    "empty items",
			input:   order.CreateOrderInput{Items: []order.Item{}},
			wantErr: order.ErrInvalidOrder,
		},
		{
			name: "empty name",
			input: order.CreateOrderInput{
				Items: []order.Item{{Name: "", Quantity: 1, Price: 1000}},
			},
			wantErr: order.ErrInvalidOrder,
		},
		{
			name: "zero quantity",
			input: order.CreateOrderInput{
				Items: []order.Item{{Name: "Keyboard", Quantity: 0, Price: 1000}},
			},
			wantErr: order.ErrInvalidOrder,
		},
		{
			name: "negative quantity",
			input: order.CreateOrderInput{
				Items: []order.Item{{Name: "Keyboard", Quantity: -1, Price: 1000}},
			},
			wantErr: order.ErrInvalidOrder,
		},
		{
			name: "zero price",
			input: order.CreateOrderInput{
				Items: []order.Item{{Name: "Keyboard", Quantity: 1, Price: 0}},
			},
			wantErr: order.ErrInvalidOrder,
		},
		{
			name: "negative price",
			input: order.CreateOrderInput{
				Items: []order.Item{{Name: "Keyboard", Quantity: 1, Price: -10}},
			},
			wantErr: order.ErrInvalidOrder,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := service.NewOrderService(newFakeRepo())

			got, err := svc.Create(context.Background(), tt.input)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantTotal, got.TotalPrice)
			assert.Equal(t, order.StatusNew, got.Status)
			assert.Len(t, got.History, 1)
		})
	}
}

func TestOrderService_GetAndList(t *testing.T) {
	ctx := context.Background()
	svc := service.NewOrderService(newFakeRepo())

	created, err := svc.Create(ctx, order.CreateOrderInput{
		Items: []order.Item{{Name: "Keyboard", Quantity: 1, Price: 5000}},
	})
	assert.NoError(t, err)

	got, err := svc.GetByID(ctx, created.ID)
	assert.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)

	_, err = svc.GetByID(ctx, 999)
	assert.ErrorIs(t, err, order.ErrOrderNotFound)

	list, err := svc.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, list, 1)
}

func TestOrderService_StatusAndHistory(t *testing.T) {
	tests := []struct {
		name    string
		from    order.Status
		to      order.Status
		wantErr error
		wantLen int
	}{
		{
			name:    "new to processing",
			from:    order.StatusNew,
			to:      order.StatusProcessing,
			wantErr: nil,
			wantLen: 2,
		},
		{
			name:    "new to cancelled",
			from:    order.StatusNew,
			to:      order.StatusCancelled,
			wantErr: nil,
			wantLen: 2,
		},
		{
			name:    "processing to completed",
			from:    order.StatusProcessing,
			to:      order.StatusCompleted,
			wantErr: nil,
			wantLen: 3,
		},
		{
			name:    "processing to cancelled",
			from:    order.StatusProcessing,
			to:      order.StatusCancelled,
			wantErr: nil,
			wantLen: 3,
		},
		{
			name:    "new to completed error",
			from:    order.StatusNew,
			to:      order.StatusCompleted,
			wantErr: order.ErrInvalidStatusTransition,
		},
		{
			name:    "completed is final",
			from:    order.StatusCompleted,
			to:      order.StatusCancelled,
			wantErr: order.ErrInvalidStatusTransition,
		},
		{
			name:    "cancelled is final",
			from:    order.StatusCancelled,
			to:      order.StatusProcessing,
			wantErr: order.ErrInvalidStatusTransition,
		},
		{
			name:    "unknown status",
			from:    order.StatusNew,
			to:      order.Status("bad"),
			wantErr: order.ErrInvalidStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			repo := newFakeRepo()
			svc := service.NewOrderService(repo)

			now := time.Now()
			history := []order.StatusChange{{Status: order.StatusNew, ChangedAt: now}}
			if tt.from != order.StatusNew {
				history = append(history, order.StatusChange{Status: tt.from, ChangedAt: now.Add(time.Second)})
			}

			repo.orders = append(repo.orders, order.Order{
				ID:      1,
				Status:  tt.from,
				History: history,
			})
			repo.nextID = 2

			updated, err := svc.UpdateStatus(ctx, 1, tt.to)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.to, updated.Status)

			hist, err := svc.History(ctx, 1)
			assert.NoError(t, err)
			assert.Len(t, hist, tt.wantLen)
			assert.Equal(t, tt.to, hist[len(hist)-1].Status)
			assert.False(t, hist[len(hist)-1].ChangedAt.Before(hist[0].ChangedAt))

		})
	}
}

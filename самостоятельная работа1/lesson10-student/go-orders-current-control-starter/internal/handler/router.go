package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	_ "example.com/order-service/docs"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"

	"example.com/order-service/internal/order"
	"github.com/go-chi/chi/v5"
)

type OrderService interface {
	Create(ctx context.Context, input order.CreateOrderInput) (order.Order, error)
	GetByID(ctx context.Context, id int64) (order.Order, error)
	List(ctx context.Context) ([]order.Order, error)
	UpdateStatus(ctx context.Context, id int64, status order.Status) (order.Order, error)
	History(ctx context.Context, id int64) ([]order.StatusChange, error)
}

type Handler struct {
	service OrderService
	logger  *slog.Logger
}

func NewRouter(service OrderService, logger *slog.Logger) http.Handler {
	h := &Handler{service: service, logger: logger}
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(recoveryMiddleware(logger))
	router.Use(requestLoggingMiddleware(logger))

	router.Get("/health", h.health)
	router.Get("/swagger/*", httpSwagger.WrapHandler)

	router.Route("/api/v1/orders", func(r chi.Router) {
		r.Post("/", h.createOrder)
		r.Get("/", h.listOrders)
		r.Get("/{id}", h.getOrder)
		r.Patch("/{id}/status", h.updateOrderStatus)
		r.Get("/{id}/history", h.getOrderHistory)
	})

	return router
}

// health godoc
// @Summary Проверка сервиса
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// createOrder godoc
// @Summary Создать заказ
// @Tags orders
// @Accept json
// @Produce json
// @Param input body order.CreateOrderInput true "Данные заказа"
// @Success 201 {object} order.Order
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/orders [post]
func (h *Handler) createOrder(w http.ResponseWriter, r *http.Request) {
	var input order.CreateOrderInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error: "invalid json",
			Code:  "INVALID_JSON",
		})
		return
	}

	res, err := h.service.Create(r.Context(), input)
	if err != nil {
		h.writeErr(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, res)
}

// listOrders godoc
// @Summary Список заказов
// @Tags orders
// @Produce json
// @Success 200 {array} order.Order
// @Failure 500 {object} errorResponse
// @Router /api/v1/orders [get]
func (h *Handler) listOrders(w http.ResponseWriter, r *http.Request) {
	res, err := h.service.List(r.Context())
	if err != nil {
		h.writeErr(w, err)
		return
	}

	writeJSON(w, http.StatusOK, res)
}

// getOrder godoc
// @Summary Получить заказ
// @Tags orders
// @Produce json
// @Param id path int true "ID заказа"
// @Success 200 {object} order.Order
// @Failure 400 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/orders/{id} [get]
func (h *Handler) getOrder(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error: "invalid id",
			Code:  "INVALID_ID",
		})
		return
	}

	res, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		h.writeErr(w, err)
		return
	}

	writeJSON(w, http.StatusOK, res)
}

// updateOrderStatus godoc
// @Summary Изменить статус заказа
// @Tags orders
// @Accept json
// @Produce json
// @Param id path int true "ID заказа"
// @Param input body order.UpdateStatusInput true "Новый статус"
// @Success 200 {object} order.Order
// @Failure 400 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 409 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/orders/{id}/status [patch]
func (h *Handler) updateOrderStatus(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error: "invalid id",
			Code:  "INVALID_ID",
		})
		return
	}

	var input order.UpdateStatusInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error: "invalid json",
			Code:  "INVALID_JSON",
		})
		return
	}

	res, err := h.service.UpdateStatus(r.Context(), id, input.Status)
	if err != nil {
		h.writeErr(w, err)
		return
	}

	writeJSON(w, http.StatusOK, res)
}

// getOrderHistory godoc
// @Summary История статусов заказа
// @Tags orders
// @Produce json
// @Param id path int true "ID заказа"
// @Success 200 {array} order.StatusChange
// @Failure 400 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/orders/{id}/history [get]
func (h *Handler) getOrderHistory(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error: "invalid id",
			Code:  "INVALID_ID",
		})
		return
	}

	res, err := h.service.History(r.Context(), id)
	if err != nil {
		h.writeErr(w, err)
		return
	}

	writeJSON(w, http.StatusOK, res)
}

func (h *Handler) writeErr(w http.ResponseWriter, err error) {
	if errors.Is(err, order.ErrInvalidOrder) || errors.Is(err, order.ErrInvalidStatus) {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error: err.Error(),
			Code:  "BAD_REQUEST",
		})
		return
	}
	if errors.Is(err, order.ErrOrderNotFound) {
		writeJSON(w, http.StatusNotFound, errorResponse{
			Error: "order not found",
			Code:  "NOT_FOUND",
		})
		return
	}
	if errors.Is(err, order.ErrInvalidStatusTransition) {
		writeJSON(w, http.StatusConflict, errorResponse{
			Error: "status transition is not allowed",
			Code:  "INVALID_STATUS_TRANSITION",
		})
		return
	}

	h.logger.Error("internal error", slog.String("err", err.Error()))
	writeJSON(w, http.StatusInternalServerError, errorResponse{
		Error: "internal server error",
		Code:  "INTERNAL_ERROR",
	})
}

type errorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

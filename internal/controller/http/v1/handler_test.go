package v1

import (
    "encoding/json"
    "errors"
    "net/http"
    "net/http/httptest"
    "order/internal/entity"
    "order/internal/service/mock"
    "testing"

    "github.com/gorilla/mux"
    "github.com/stretchr/testify/assert"
    "go.uber.org/mock/gomock"
)

func TestHandler_GetOrder(t *testing.T) {
    t.Run("Valid order", func(t *testing.T) {
        ctrl := gomock.NewController(t)
        defer ctrl.Finish()

        mockService := mock.NewMockService(ctrl)
        handler := NewHandler(mockService)

        orderUID := "test-uid"
        order := entity.Order{
            OrderUID:    orderUID,
            Delivery:    entity.Delivery{Name: "John", Phone: "1234567890"},
            Payment:     entity.Payment{Amount: 1000},
            Items:       []entity.Item{{ChrtID: 1, Price: 500}},
            DateCreated: "2025-08-09T10:30:00Z",
        }
        req := httptest.NewRequest(http.MethodGet, "/order/"+orderUID, nil)
        req = mux.SetURLVars(req, map[string]string{"order_uid": orderUID})
        ctx := req.Context()
        mockService.EXPECT().GetOrder(ctx, orderUID).Return(order, nil)

        w := httptest.NewRecorder()
        handler.GetOrder(w, req)

        assert.Equal(t, http.StatusOK, w.Code)
        var result entity.Order
        err := json.NewDecoder(w.Body).Decode(&result)
        assert.NoError(t, err)
        assert.Equal(t, order, result)
        assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
    })

    t.Run("Missing order_uid", func(t *testing.T) {
        ctrl := gomock.NewController(t)
        defer ctrl.Finish()

        mockService := mock.NewMockService(ctrl)
        handler := NewHandler(mockService)

        req := httptest.NewRequest(http.MethodGet, "/order/", nil)
        w := httptest.NewRecorder()

        handler.GetOrder(w, req)

        assert.Equal(t, http.StatusBadRequest, w.Code)
        assert.Contains(t, w.Body.String(), "order_uid is required")
    })

    t.Run("Order not found", func(t *testing.T) {
        ctrl := gomock.NewController(t)
        defer ctrl.Finish()

        mockService := mock.NewMockService(ctrl)
        handler := NewHandler(mockService)

        orderUID := "test-uid"
        req := httptest.NewRequest(http.MethodGet, "/order/"+orderUID, nil)
        req = mux.SetURLVars(req, map[string]string{"order_uid": orderUID})
        ctx := req.Context()
        mockService.EXPECT().GetOrder(ctx, orderUID).Return(entity.Order{}, errors.New("not found"))

        w := httptest.NewRecorder()
        handler.GetOrder(w, req)

        assert.Equal(t, http.StatusNotFound, w.Code)
        assert.Contains(t, w.Body.String(), "Order not found")
    })

    t.Run("JSON encode error", func(t *testing.T) {
        ctrl := gomock.NewController(t)
        defer ctrl.Finish()

        mockService := mock.NewMockService(ctrl)
        handler := NewHandler(mockService)

        orderUID := "test-uid"
        order := entity.Order{
            OrderUID:    orderUID,
            Delivery:    entity.Delivery{Name: "John", Phone: "1234567890"},
            Payment:     entity.Payment{Amount: 1000},
            Items:       []entity.Item{{ChrtID: 1, Price: 500}},
            DateCreated: "2025-08-09T10:30:00Z",
        }
        req := httptest.NewRequest(http.MethodGet, "/order/"+orderUID, nil)
        req = mux.SetURLVars(req, map[string]string{"order_uid": orderUID})
        ctx := req.Context()
        mockService.EXPECT().GetOrder(ctx, orderUID).Return(order, nil)

        w := &errorResponseWriter{Recorder: httptest.NewRecorder()}
        handler.GetOrder(w, req)

        assert.Equal(t, http.StatusInternalServerError, w.Recorder.Code)
        assert.Contains(t, w.Recorder.Body.String(), "Internal server error")
    })
}

// errorResponseWriter заставляет json.NewEncoder завершаться с ошибкой,
// но позволяет http.Error записать сообщение об ошибке
type errorResponseWriter struct {
    Recorder   *httptest.ResponseRecorder
    jsonFailed bool
}

func (w *errorResponseWriter) Header() http.Header {
    return w.Recorder.Header()
}

func (w *errorResponseWriter) Write(b []byte) (int, error) {
    // Если это попытка записи JSON, возвращаем ошибку
    if !w.jsonFailed {
        w.jsonFailed = true
        return 0, errors.New("write error")
    }
    // Позволяем http.Error записать сообщение об ошибке
    return w.Recorder.Write(b)
}

func (w *errorResponseWriter) WriteHeader(statusCode int) {
    w.Recorder.WriteHeader(statusCode)
}

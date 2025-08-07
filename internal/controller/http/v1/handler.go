package v1

import (
	"encoding/json"
	"log"
	"net/http"
	"order/internal/service"

	"github.com/gorilla/mux"
)

type Handler struct {
	service service.Service
}

func NewHandler(service service.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderUID := vars["order_uid"]
	if orderUID == "" {
		log.Printf("Invalid request: empty order_uid")
		http.Error(w, "order_uid is required", http.StatusBadRequest)
		return
	}

	order, err := h.service.GetOrder(r.Context(), orderUID)
	if err != nil {
		log.Printf("Failed to get order %s: %v", orderUID, err)
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(order); err != nil {
		log.Printf("Failed to encode response for order %s: %v", orderUID, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully served order %s", orderUID)
}

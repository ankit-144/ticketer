package booking

import (
	"encoding/json"
	"net/http"
)

// RouteRegistrar is implemented by any handler that wants to register
// its routes on the shared HTTP mux.  This is the contract used by the
// fx "routes" group.
type RouteRegistrar interface {
	RegisterRoutes(mux *http.ServeMux)
}

// Handler exposes the BookingService over HTTP.
type Handler struct {
	service Service
}

// NewHandler constructs a Handler with its service dependency.
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes mounts all booking-related endpoints on the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/bookings", h.initiateBooking)
	mux.HandleFunc("POST /api/bookings/{id}/confirm", h.confirmBooking)
	mux.HandleFunc("POST /api/bookings/{id}/cancel", h.cancelBooking)
	mux.HandleFunc("POST /api/bookings/{id}/revert", h.revertBooking)
}

// --- request / response types ------------------------------------------------

type initiateBookingRequest struct {
	UserID  string   `json:"user_id"`
	ShowID  string   `json:"show_id"`
	SeatIDs []string `json:"seat_ids"`
}

type bookingResponse struct {
	ID      string   `json:"id"`
	ShowID  string   `json:"show_id"`
	UserID  string   `json:"user_id"`
	SeatIDs []string `json:"seat_ids"`
	Price   float64  `json:"price"`
	Status  string   `json:"status"`
}

// --- handlers ----------------------------------------------------------------

func (h *Handler) initiateBooking(w http.ResponseWriter, r *http.Request) {
	var req initiateBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	b, err := h.service.InitiateBooking(req.UserID, req.ShowID, req.SeatIDs)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(bookingResponse{
		ID:      b.ID,
		ShowID:  b.ShowID,
		UserID:  b.UserID,
		SeatIDs: b.SeatIDs,
		Price:   b.Price,
		Status:  string(b.Status),
	})
}

func (h *Handler) confirmBooking(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.service.ConfirmBooking(id); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "confirmed"})
}

func (h *Handler) cancelBooking(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.service.CancelBooking(id); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "cancelled"})
}

func (h *Handler) revertBooking(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.service.RevertBooking(id); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "reverted"})
}

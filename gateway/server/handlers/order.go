package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/yash-gadgil/glyph/gateway/server/utils"
)

func GetOrders(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")

	if status == "" || !utils.Contains(utils.ValidOrderStatuses, status) {

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "invalid or missing 'status' parameter",
		})
		return
	}

	w.Write([]byte("your " + status + " orders"))
}

func CreateOrder(w http.ResponseWriter, r *http.Request) {
	var no utils.Order
	if err := json.NewDecoder(r.Body).Decode(&no); err != nil {

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "invalid order parameter",
		})
		return
	}

	w.Write([]byte("New order made"))
}

func DeleteOrder(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "id")

	w.Write([]byte("Order " + orderID + " to be Deleted"))
}

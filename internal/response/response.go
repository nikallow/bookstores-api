package response

import (
	"encoding/json"
	"net/http"

	"github.com/nikallow/bookstores-api/internal/middleware"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func WriteJSON(w http.ResponseWriter, r *http.Request, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		middleware.LoggerFromContext(r.Context()).Error("Failed to write HTTP response", "error", err)
	}
}

func WriteError(w http.ResponseWriter, r *http.Request, status int, msg string) {
	WriteJSON(w, r, status, ErrorResponse{Error: msg})
}

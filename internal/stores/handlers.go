package stores

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/nikallow/bookstores-api/internal/middleware"
)

type Handler struct {
	service  Service
	validate *validator.Validate
}

func NewHandler(service Service) *Handler {
	return &Handler{
		service:  service,
		validate: validator.New(),
	}
}

func (h *Handler) writeJSONResponse(w http.ResponseWriter, r *http.Request, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		middleware.LoggerFromContext(r.Context()).Error("Failed to write HTTP response", "error", err)
	}
}

func (h *Handler) writeErrorResponse(w http.ResponseWriter, r *http.Request, status int, msg string) {
	h.writeJSONResponse(w, r, status, map[string]string{"error": msg})
}

// CreateStore - POST /stores
func (h *Handler) CreateStore(w http.ResponseWriter, r *http.Request) {
	log := middleware.LoggerFromContext(r.Context())

	var req CreateStoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn("Failed to read create store request", "error", err)
		h.writeErrorResponse(w, r, http.StatusBadRequest, "Invalid request body")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		log.Warn("Validation failed for create store request", "error", err)
		h.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
	}

	store, err := h.service.Create(r.Context(), req.Name, req.Address)
	if err != nil {
		log.Error("Failed to create store", "error", err)
		h.writeErrorResponse(w, r, http.StatusInternalServerError, "Internal server error")
		return
	}

	resp := StoreResponse{
		UUID:    store.Uuid.Bytes,
		Name:    store.Name,
		Address: store.Address,
	}

	h.writeJSONResponse(w, r, http.StatusCreated, resp)
}

// ListStores - GET /stores
func (h *Handler) ListStores(w http.ResponseWriter, r *http.Request) {
	log := middleware.LoggerFromContext(r.Context())

	stores, err := h.service.List(r.Context())
	if err != nil {
		log.Error("Failed to list stores", "error", err)
		h.writeErrorResponse(w, r, http.StatusInternalServerError, "Internal server error")
		return
	}

	resp := make([]StoreResponse, len(stores))
	for i, s := range stores {
		resp[i] = StoreResponse{
			UUID:    s.Uuid.Bytes,
			Name:    s.Name,
			Address: s.Address,
		}
	}

	h.writeJSONResponse(w, r, http.StatusOK, resp)
}

// GetStore - GET /stores/{storeUUID}
func (h *Handler) GetStore(w http.ResponseWriter, r *http.Request) {
	log := middleware.LoggerFromContext(r.Context())

	uuidStr := chi.URLParam(r, "storeUUID")
	id, err := uuid.Parse(uuidStr)
	if err != nil {
		log.Warn("Invalid store UUID format", "error", err, "uuid_str", uuidStr)
		h.writeErrorResponse(w, r, http.StatusBadRequest, "Invalid store UUID format")
		return
	}

	store, err := h.service.GetByUUID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrStoreNotFound) {
			h.writeErrorResponse(w, r, http.StatusNotFound, "Store not found")
		} else {
			log.Error("Failed to get store by UUID", "error", err, "store_uuid", id)
			h.writeErrorResponse(w, r, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	resp := StoreResponse{
		UUID:    store.Uuid.Bytes,
		Name:    store.Name,
		Address: store.Address,
	}
	h.writeJSONResponse(w, r, http.StatusOK, resp)
}

// UpdateStore - PUT /stores/{storeUUID}
func (h *Handler) UpdateStore(w http.ResponseWriter, r *http.Request) {
	log := middleware.LoggerFromContext(r.Context())

	uuidStr := chi.URLParam(r, "storeUUID")
	id, err := uuid.Parse(uuidStr)
	if err != nil {
		log.Warn("Invalid store UUID format for update", "error", err, "uuid_str", uuidStr)
		h.writeErrorResponse(w, r, http.StatusBadRequest, "Invalid store UUID format")
		return
	}

	var req UpdateStoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn("Failed to read update store request", "error", err)
		h.writeErrorResponse(w, r, http.StatusBadRequest, "Invalid request body")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		log.Warn("Validation failed for update store request", "error", err)
		h.writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	store, err := h.service.Update(r.Context(), id, req.Name, req.Address)
	if err != nil {
		if errors.Is(err, ErrStoreNotFound) {
			h.writeErrorResponse(w, r, http.StatusNotFound, "Store not found")
		} else {
			log.Error("Failed to update store", "error", err, "store_uuid", id)
			h.writeErrorResponse(w, r, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	resp := StoreResponse{
		UUID:    store.Uuid.Bytes,
		Name:    store.Name,
		Address: store.Address,
	}

	h.writeJSONResponse(w, r, http.StatusOK, resp)
}

// DeleteStore - DELETE /stores/{storeUUID}
func (h *Handler) DeleteStore(w http.ResponseWriter, r *http.Request) {
	log := middleware.LoggerFromContext(r.Context())

	uuidStr := chi.URLParam(r, "storeUUID")
	id, err := uuid.Parse(uuidStr)
	if err != nil {
		log.Warn("Invalid store UUID format for delete", "error", err, "uuid_str", uuidStr)
		h.writeErrorResponse(w, r, http.StatusBadRequest, "Invalid store UUID format")
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		log.Error("Failed to delete store", "error", err, "store_uuid", id)
		h.writeErrorResponse(w, r, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

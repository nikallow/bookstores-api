package inventory

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/nikallow/bookstores-api/internal/middleware"
	"github.com/nikallow/bookstores-api/internal/response"
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

// CreateSKU - POST /skus
func (h *Handler) CreateSKU(w http.ResponseWriter, r *http.Request) {
	log := middleware.LoggerFromContext(r.Context())

	var req CreateSKURequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn("Failed to read create SKU request", "error", err)
		response.WriteError(w, r, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		log.Warn("Validation failed", "error", err)
		response.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	sku, err := h.service.CreateSKU(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrBookNotFound), errors.Is(err, ErrStoreNotFound):
			response.WriteError(w, r, http.StatusNotFound, err.Error())
		case errors.Is(err, ErrSKUAlreadyExists):
			response.WriteError(w, r, http.StatusConflict, err.Error())
		default:
			log.Error("Failed to create sku", "error", err)
			response.WriteError(w, r, http.StatusInternalServerError, err.Error())
		}
		return
	}

	response.WriteJSON(w, r, http.StatusCreated, sku)
}

// GetSKU - GET /skus/{skuUUID}
func (h *Handler) GetSKU(w http.ResponseWriter, r *http.Request) {
	log := middleware.LoggerFromContext(r.Context())

	skuUUID, err := uuid.Parse(chi.URLParam(r, "skuUUID"))
	if err != nil {
		log.Error("Error parsing UUID", "error", err, "skuUUID", skuUUID)
		response.WriteError(w, r, http.StatusBadRequest, "Invalid sku uuid format")
		return
	}

	sku, err := h.service.GetSKU(r.Context(), skuUUID)
	if err != nil {
		if errors.Is(err, ErrSKUNotFound) {
			log.Error("SKU not found", "error", err, "skuUUID", skuUUID)
			response.WriteError(w, r, http.StatusNotFound, "SKU not found")
		} else {
			log.Error("Failed to get sku", "error", err)
			response.WriteError(w, r, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	response.WriteJSON(w, r, http.StatusOK, sku)
}

// UpdateSKUPrice - PUT /skus/{skuUUID}/price
func (h *Handler) UpdateSKUPrice(w http.ResponseWriter, r *http.Request) {
	log := middleware.LoggerFromContext(r.Context())

	skuUUID, err := uuid.Parse(chi.URLParam(r, "skuUUID"))
	if err != nil {
		log.Error("Error parsing UUID", "error", err, "skuUUID", skuUUID)
		response.WriteError(w, r, http.StatusBadRequest, "Invalid sku uuid format")
		return
	}

	var req UpdateSKUPriceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn("Failed to read update SKU price request", "error", err)
		response.WriteError(w, r, http.StatusBadRequest, "Invalid request body")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		log.Warn("Validation failed", "error", err)
		response.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	sku, err := h.service.UpdateSKUPrice(r.Context(), skuUUID, req.NewPriceInKopeks)
	if err != nil {
		if errors.Is(err, ErrSKUNotFound) {
			response.WriteError(w, r, http.StatusNotFound, "SKU not found")
		} else {
			log.Error("Failed to update sku price", "error", err)
			response.WriteError(w, r, http.StatusInternalServerError, "Internal server error")
		}
		return
	}
	response.WriteJSON(w, r, http.StatusOK, sku)
}

// AdjustSKUStock - POST /skus/{skuUUID}/stock-adjustments
func (h *Handler) AdjustSKUStock(w http.ResponseWriter, r *http.Request) {
	log := middleware.LoggerFromContext(r.Context())

	skuUUID, err := uuid.Parse(chi.URLParam(r, "skuUUID"))
	if err != nil {
		log.Error("Error parsing UUID", "error", err, "skuUUID", skuUUID)
		response.WriteError(w, r, http.StatusBadRequest, "Invalid sku uuid format")
		return
	}

	var req AdjustSKUStockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn("Failed to read adjust SKU stock request", "error", err)
		response.WriteError(w, r, http.StatusBadRequest, "Invalid request body")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		log.Warn("Validation failed", "error", err)
		response.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	sku, err := h.service.AdjustSKUStock(r.Context(), skuUUID, req.ChangeBy)
	if err != nil {
		switch {
		case errors.Is(err, ErrSKUNotFound):
			response.WriteError(w, r, http.StatusNotFound, "SKU not found")
		case errors.Is(err, ErrInsufficientStock):
			response.WriteError(w, r, http.StatusConflict, err.Error())
		default:
			log.Error("Failed to adjust sku stock", "error", err)
			response.WriteError(w, r, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	response.WriteJSON(w, r, http.StatusOK, sku)
}

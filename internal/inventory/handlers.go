package inventory

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	repo "github.com/nikallow/bookstores-api/internal/adapters/postgres/sqlc"
	"github.com/nikallow/bookstores-api/internal/books"
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

// CreateSKU
//
//	@Summary		Создать SKU
//	@Description	Создает новую товарную позицию (SKU), связывая книгу с магазином, ценой и остатком.
//	@Tags			skus
//	@Accept			json
//	@Produce		json
//	@Param			input	body		CreateSKURequest		true	"Данные для создания SKU"
//	@Success		201		{object}	SKUResponse				"SKU успешно создан"
//	@Failure		400		{object}	response.ErrorResponse	"Bad request error"
//	@Failure		404		{object}	response.ErrorResponse	"Книга или магазин не найдены"
//	@Failure		409		{object}	response.ErrorResponse	"SKU для этой книги в этом магазине уже существует"
//	@Failure		500		{object}	response.ErrorResponse	"Internal server error"
//	@Router			/skus [post]
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

	response.WriteJSON(w, r, http.StatusCreated, toSKUResponse(sku))
}

// GetSKU
//
//	@Summary		Получить SKU
//	@Description	Возвращает детальную информацию о SKU (включая данные о книге) по его UUID.
//	@Tags			skus
//	@Produce		json
//	@Param			skuUUID	path		string					true	"UUID товарной позиции (SKU)"
//	@Success		200		{object}	SKUWithBookResponse		"Информация о SKU и связанной книге"
//	@Failure		400		{object}	response.ErrorResponse	"Bad request error"
//	@Failure		404		{object}	response.ErrorResponse	"SKU не найден"
//	@Failure		500		{object}	response.ErrorResponse	"Internal server error"
//	@Router			/skus/{skuUUID} [get]
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

	response.WriteJSON(w, r, http.StatusOK, toSKUWithBookResponse(sku))
}

// UpdateSKUPrice
//
//	@Summary		Обновить цену SKU
//	@Description	Устанавливает новую цену для существующей товарной позиции (SKU).
//	@Tags			skus
//	@Accept			json
//	@Produce		json
//	@Param			skuUUID	path		string					true	"UUID товарной позиции (SKU)"
//	@Param			input	body		UpdateSKUPriceRequest	true	"Новая цена"
//	@Success		200		{object}	SKUResponse				"Обновленный SKU"
//	@Failure		400		{object}	response.ErrorResponse	"Bad request error"
//	@Failure		404		{object}	response.ErrorResponse	"SKU не найден"
//	@Failure		500		{object}	response.ErrorResponse	"Internal server error"
//	@Router			/skus/{skuUUID}/price [put]
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
	response.WriteJSON(w, r, http.StatusOK, toSKUResponse(sku))
}

// AdjustSKUStock
//
//	@Summary		Скорректировать остатки
//	@Description	Увеличивает или уменьшает количество товара на складе. Для уменьшения используйте отрицательное значение.
//	@Tags			skus
//	@Accept			json
//	@Produce		json
//	@Param			skuUUID	path		string					true	"UUID товарной позиции (SKU)"
//	@Param			input	body		AdjustSKUStockRequest	true	"Количество для изменения"
//	@Success		200		{object}	SKUResponse				"Обновленный SKU"
//	@Failure		400		{object}	response.ErrorResponse	"Bad request error"
//	@Failure		404		{object}	response.ErrorResponse	"SKU не найден"
//	@Failure		409		{object}	response.ErrorResponse	"Недостаточно товара для списания"
//	@Failure		500		{object}	response.ErrorResponse	"Internal error"
//	@Router			/skus/{skuUUID}/stock-adjustments [post]
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

	response.WriteJSON(w, r, http.StatusOK, toSKUResponse(sku))
}

func toSKUResponse(sku repo.Sku) SKUResponse {
	return SKUResponse{
		ID:            sku.ID,
		UUID:          mustConvertUUID(sku.Uuid),
		BookID:        sku.BookID,
		StoreID:       sku.StoreID,
		PriceInKopeks: sku.PriceInKopeks,
		StockCount:    sku.StockCount,
		CreatedAt:     sku.CreatedAt.Time,
		UpdatedAt:     sku.UpdatedAt.Time,
	}
}

func toSKUWithBookResponse(row repo.GetSKUByUUIDRow) SKUWithBookResponse {
	return SKUWithBookResponse{
		SKU:  toSKUResponse(row.Sku),
		Book: books.ToBookResponse(row.Book),
	}
}

func mustConvertUUID(pgUUID pgtype.UUID) uuid.UUID {
	if !pgUUID.Valid {
		return uuid.Nil
	}
	return pgUUID.Bytes
}

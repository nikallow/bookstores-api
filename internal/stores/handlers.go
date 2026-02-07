package stores

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

// CreateStore
//
//	@Summary		Создать новый магазин
//	@Description	Создаёт новую запись о магазине в базе данных.
//	@Tags			stores
//	@Accept			json
//	@Produce		json
//	@Param			input	body		CreateStoreRequest		true	"Данные для создания магазина"
//	@Success		201		{object}	StoreResponse			"Магазин успешно создан"
//	@Failure		400		{object}	response.ErrorResponse	"Bad request error"
//	@Failure		500		{object}	response.ErrorResponse	"Internal server error"
//	@Router			/stores [post]
func (h *Handler) CreateStore(w http.ResponseWriter, r *http.Request) {
	log := middleware.LoggerFromContext(r.Context())

	var req CreateStoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn("Failed to read create store request", "error", err)
		response.WriteError(w, r, http.StatusBadRequest, "Invalid request body")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		log.Warn("Validation failed for create store request", "error", err)
		response.WriteError(w, r, http.StatusBadRequest, err.Error())
	}

	store, err := h.service.Create(r.Context(), req.Name, req.Address)
	if err != nil {
		log.Error("Failed to create store", "error", err)
		response.WriteError(w, r, http.StatusInternalServerError, "Internal server error")
		return
	}

	resp := StoreResponse{
		UUID:    store.Uuid.Bytes,
		Name:    store.Name,
		Address: store.Address,
	}

	response.WriteJSON(w, r, http.StatusCreated, resp)
}

// ListStores
//
//	@Summary		Получить список магазинов
//	@Description	Возвращает список всех действующих магазинов.
//	@Tags			stores
//	@Produce		json
//	@Success		200	{array}		StoreResponse			"Список действующих магазинов"
//	@Failure		500	{object}	response.ErrorResponse	"Internal server error"
//	@Router			/stores [get]
func (h *Handler) ListStores(w http.ResponseWriter, r *http.Request) {
	log := middleware.LoggerFromContext(r.Context())

	stores, err := h.service.List(r.Context())
	if err != nil {
		log.Error("Failed to list stores", "error", err)
		response.WriteError(w, r, http.StatusInternalServerError, "Internal server error")
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

	response.WriteJSON(w, r, http.StatusOK, resp)
}

// GetStore
//
//	@Summary		Получить информацию об одном магазине
//	@Description	Возвращает детальную информацию о магазине по его UUID.
//	@Tags			stores
//	@Produce		json
//	@Param			storeUUID	path		string					true	"UUID магазина"
//	@Success		200			{object}	StoreResponse			"Инфо о найденном магазине"
//	@Failure		400			{object}	response.ErrorResponse	"Bad request error"
//	@Failure		404			{object}	response.ErrorResponse	"Искомый магазин отсутствует"
//	@Failure		500			{object}	response.ErrorResponse	"Internal server error"
//	@Router			/stores/{storeUUID} [get]
func (h *Handler) GetStore(w http.ResponseWriter, r *http.Request) {
	log := middleware.LoggerFromContext(r.Context())

	uuidStr := chi.URLParam(r, "storeUUID")
	id, err := uuid.Parse(uuidStr)
	if err != nil {
		log.Warn("Invalid store UUID format", "error", err, "uuid_str", uuidStr)
		response.WriteError(w, r, http.StatusBadRequest, "Invalid store UUID format")
		return
	}

	store, err := h.service.GetByUUID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrStoreNotFound) {
			response.WriteError(w, r, http.StatusNotFound, "Store not found")
		} else {
			log.Error("Failed to get store by UUID", "error", err, "store_uuid", id)
			response.WriteError(w, r, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	resp := StoreResponse{
		UUID:    store.Uuid.Bytes,
		Name:    store.Name,
		Address: store.Address,
	}
	response.WriteJSON(w, r, http.StatusOK, resp)
}

// UpdateStore
//
//	@Summary		Обновить информацию о магазине
//	@Description	Обнавляет запись о магазине в базе данных.
//	@Tags			stores
//	@Accept			json
//	@Produce		json
//	@Param			storeUUID	path		string					true	"UUID магазина"
//	@Param			input		body		UpdateStoreRequest		true	"Данные для обновления информации о магазине"
//	@Success		200			{object}	StoreResponse			"Обновлённое инфо об обновлённом магазине"
//	@Failure		400			{object}	response.ErrorResponse	"Bad request error"
//	@Failure		500			{object}	response.ErrorResponse	"Internal server error"
//	@Router			/stores/{storeUUID} [put]
func (h *Handler) UpdateStore(w http.ResponseWriter, r *http.Request) {
	log := middleware.LoggerFromContext(r.Context())

	uuidStr := chi.URLParam(r, "storeUUID")
	id, err := uuid.Parse(uuidStr)
	if err != nil {
		log.Warn("Invalid store UUID format for update", "error", err, "uuid_str", uuidStr)
		response.WriteError(w, r, http.StatusBadRequest, "Invalid store UUID format")
		return
	}

	var req UpdateStoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn("Failed to read update store request", "error", err)
		response.WriteError(w, r, http.StatusBadRequest, "Invalid request body")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		log.Warn("Validation failed for update store request", "error", err)
		response.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	store, err := h.service.Update(r.Context(), id, req.Name, req.Address)
	if err != nil {
		if errors.Is(err, ErrStoreNotFound) {
			response.WriteError(w, r, http.StatusNotFound, "Store not found")
		} else {
			log.Error("Failed to update store", "error", err, "store_uuid", id)
			response.WriteError(w, r, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	resp := StoreResponse{
		UUID:    store.Uuid.Bytes,
		Name:    store.Name,
		Address: store.Address,
	}

	response.WriteJSON(w, r, http.StatusOK, resp)
}

// DeleteStore
//
//	@Summary		Удалить магазин из доступных
//	@Description	Выполняет мягкое удаление магазина
//	@Tags			stores
//	@Param			storeUUID	path	string	true	"UUID магазина"
//	@Success		204			"Магазин удалён (деактивирован)"
//	@Failure		400			{object}	response.ErrorResponse	"Bad request error"
//	@Failure		500			{object}	response.ErrorResponse	"Internal server error"
//	@Router			/stores/{storeUUID} [delete]
func (h *Handler) DeleteStore(w http.ResponseWriter, r *http.Request) {
	log := middleware.LoggerFromContext(r.Context())

	uuidStr := chi.URLParam(r, "storeUUID")
	id, err := uuid.Parse(uuidStr)
	if err != nil {
		log.Warn("Invalid store UUID format for delete", "error", err, "uuid_str", uuidStr)
		response.WriteError(w, r, http.StatusBadRequest, "Invalid store UUID format")
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		log.Error("Failed to delete store", "error", err, "store_uuid", id)
		response.WriteError(w, r, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

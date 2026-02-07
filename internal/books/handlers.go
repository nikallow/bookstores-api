package books

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	repo "github.com/nikallow/bookstores-api/internal/adapters/postgres/sqlc"
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

// CreateBook
//
//	@Summary		Создать новую книгу
//	@Description	Создаёт новую книгу в глобальном каталоге. Если книга с таким ISBN уже существует, обновляет ее данные.
//	@Tags			books
//	@Accept			json
//	@Produce		json
//	@Param			input	body		CreateBookRequest	true	"Данные для создания книги"
//	@Success		201		{object}	BookResponse		"Инфо об книге"
//	@Failure		400		{object}	response.ErrorResponse
//	@Failure		500		{object}	response.ErrorResponse
//	@Router			/books [post]
func (h *Handler) CreateBook(w http.ResponseWriter, r *http.Request) {
	log := middleware.LoggerFromContext(r.Context())

	var req CreateBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn("Failed to read create book request", "error", err)
		response.WriteError(w, r, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		log.Warn("Validation failed for create book request", "error", err)
		response.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	book, err := h.service.Create(r.Context(), req)
	if err != nil {
		log.Error("Failed to create book", "error", err)
		response.WriteError(w, r, http.StatusInternalServerError, "failed to create book")
		return
	}

	response.WriteJSON(w, r, http.StatusCreated, ToBookResponse(book))
}

// ListBooks
//
//	@Summary		Получить глобальный список книг
//	@Description	Возвращает список всех книг в глобальном каталоге.
//	@Tags			books
//	@Produce		json
//	@Success		200	{array}		BookResponse			"Список книг"
//	@Failure		500	{object}	response.ErrorResponse	"Internal server error"
//	@Router			/books [get]
func (h *Handler) ListBooks(w http.ResponseWriter, r *http.Request) {
	log := middleware.LoggerFromContext(r.Context())

	books, err := h.service.List(r.Context())
	if err != nil {
		log.Error("Failed to list books", "error", err)
		response.WriteError(w, r, http.StatusInternalServerError, "failed to list books")
		return
	}

	resp := make([]BookResponse, len(books))
	for i, b := range books {
		resp[i] = ToBookResponse(b)
	}

	response.WriteJSON(w, r, http.StatusOK, resp)
}

// GetBook
//
//	@Summary		Получить инфо об одной книге
//	@Description	Возвращает информацию о книге по её ID.
//	@Tags			books
//	@Produce		json
//	@Param			bookID	path		int						true	"ID книги"
//	@Success		200		{object}	BookResponse			"Инфо о книге"
//	@Failure		400		{object}	response.ErrorResponse	"Bad request error"
//	@Failure		404		{object}	response.ErrorResponse	"Книга отсутствует"
//	@Failure		500		{object}	response.ErrorResponse	"Internal server error"
//	@Router			/books/{bookID} [get]
func (h *Handler) GetBook(w http.ResponseWriter, r *http.Request) {
	log := middleware.LoggerFromContext(r.Context())

	bookIDStr := chi.URLParam(r, "bookID")
	bookID, err := strconv.ParseInt(bookIDStr, 10, 64)
	if err != nil {
		log.Warn("Invalid book ID format", "book_id", bookIDStr)
		response.WriteError(w, r, http.StatusBadRequest, "Invalid book ID format")
		return
	}

	book, err := h.service.GetByID(r.Context(), bookID)
	if err != nil {
		if errors.Is(err, ErrBookNotFound) {
			response.WriteError(w, r, http.StatusNotFound, "Book not found")
		} else {
			log.Error("Failed to get book by ID", "error", err, "book_id", bookID)
			response.WriteError(w, r, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	response.WriteJSON(w, r, http.StatusOK, ToBookResponse(book))
}

// SearchBooks
//
//	@Summary		Поиск книг
//	@Description	Ищет книги по части названия или имени автора.
//	@Tags			books
//	@Produce		json
//	@Param			q	query		string					true	"Поисковый запрос"
//	@Success		200	{array}		BookResponse			"Список найденных книг"
//	@Failure		400	{object}	response.ErrorResponse	"Bad request error"
//	@Failure		500	{object}	response.ErrorResponse	"Internal server error"
//	@Router			/books/search [get]
func (h *Handler) SearchBooks(w http.ResponseWriter, r *http.Request) {
	log := middleware.LoggerFromContext(r.Context())

	query := r.URL.Query().Get("q")
	if query == "" {
		log.Error("No search query")
		response.WriteError(w, r, http.StatusBadRequest, "query parameter 'q' is required")
		return
	}

	books, err := h.service.Search(r.Context(), query)
	if err != nil {
		log.Error("Failed to search books", "error", err)
		response.WriteError(w, r, http.StatusInternalServerError, "failed to search books")
		return
	}

	resp := make([]BookResponse, len(books))
	for i, b := range books {
		resp[i] = ToBookResponse(b)
	}

	response.WriteJSON(w, r, http.StatusOK, resp)
}

// GetBookAvailability
//
//	@Summary		Доступность книги
//	@Description	Показывает, в каких магазинах, по какой цене и в каком количестве доступна книга.
//	@Tags			books
//	@Produce		json
//	@Param			bookID	path		int	true	"ID книги"
//	@Success		200		{array}		AvailabilityResponse
//	@Failure		400		{object}	response.ErrorResponse	"Bad request error"
//	@Failure		404		{object}	response.ErrorResponse	"Книга отсутствует"
//	@Failure		500		{object}	response.ErrorResponse	"Internal server error"
//	@Router			/books/{bookID}/availability [get]
func (h *Handler) GetBookAvailability(w http.ResponseWriter, r *http.Request) {
	log := middleware.LoggerFromContext(r.Context())

	bookIDStr := chi.URLParam(r, "bookID")
	bookID, err := strconv.ParseInt(bookIDStr, 10, 64)
	if err != nil {
		log.Warn("Invalid book ID format", "book_id", bookIDStr)
		response.WriteError(w, r, http.StatusBadRequest, "Invalid book ID format")
		return
	}

	availability, err := h.service.GetAvailability(r.Context(), bookID)
	if err != nil {
		if errors.Is(err, ErrBookNotFound) {
			log.Error("Book not found", "book_id", bookID)
			response.WriteError(w, r, http.StatusNotFound, "Book not found")
			return
		}
		log.Error("Failed to get book availability", "error", err, "book_id", bookID)
		response.WriteError(w, r, http.StatusInternalServerError, "Internal server error")
		return
	}

	resp := make([]AvailabilityResponse, len(availability))
	for i, a := range availability {
		resp[i] = AvailabilityResponse{
			StoreUUID:     a.Store.Uuid.Bytes,
			StoreName:     a.Store.Name,
			SkuUUID:       a.Sku.Uuid.Bytes,
			PriceInKopeks: a.Sku.PriceInKopeks,
			StockCount:    a.Sku.StockCount,
		}
	}

	response.WriteJSON(w, r, http.StatusOK, resp)
}

func ToBookResponse(book repo.Book) BookResponse {
	resp := BookResponse{
		ID:     book.ID,
		Title:  book.Title,
		Author: book.Author,
	}
	if book.Isbn.Valid {
		resp.ISBN = &book.Isbn.String
	}
	if book.Description.Valid {
		resp.Description = &book.Description.String
	}
	if book.PageCount.Valid {
		resp.PageCount = &book.PageCount.Int32
	}
	if book.PublicationYear.Valid {
		resp.PublicationYear = &book.PublicationYear.Int32
	}
	return resp
}

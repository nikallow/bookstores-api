package inventory

import (
	"time"

	"github.com/google/uuid"
	"github.com/nikallow/bookstores-api/internal/books"
)

type CreateSKURequest struct {
	BookID        int64     `json:"book_id"         validate:"required"`
	StoreUUID     uuid.UUID `json:"store_uuid"      validate:"required"`
	PriceInKopeks int32     `json:"price_in_kopeks" validate:"gte=0"`
	StockCount    int32     `json:"stock_count"     validate:"gte=0"`
}

type UpdateSKUPriceRequest struct {
	NewPriceInKopeks int32 `json:"new_price_in_kopeks" validate:"gte=0"`
}

type AdjustSKUStockRequest struct {
	ChangeBy int32 `json:"change_by"`
}

type SKUResponse struct {
	ID            int64     `json:"id"`
	UUID          uuid.UUID `json:"uuid"`
	BookID        int64     `json:"book_id"`
	StoreID       int64     `json:"store_id"`
	PriceInKopeks int32     `json:"price_in_kopeks"`
	StockCount    int32     `json:"stock_count"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type SKUWithBookResponse struct {
	SKU  SKUResponse        `json:"sku"`
	Book books.BookResponse `json:"book"`
}

package inventory

import "github.com/google/uuid"

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
	ChangeBy int32  `json:"change_by"`
	Comment  string `json:"comment,omitempty"`
}

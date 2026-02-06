package books

import "github.com/google/uuid"

type CreateBookRequest struct {
	ISBN            *string `json:"isbn,omitempty" validate:"required"`
	Title           string  `json:"title" validate:"required"`
	Author          string  `json:"author" validate:"required"`
	Description     *string `json:"description,omitempty"`
	PageCount       *int32  `json:"page_count,omitempty" validate:"omitempty,gt=0"`
	PublicationYear *int32  `json:"publication_year,omitempty"`
}

type BookResponse struct {
	ID              int64   `json:"id"`
	ISBN            *string `json:"isbn,omitempty"`
	Title           string  `json:"title"`
	Author          string  `json:"author"`
	Description     *string `json:"description,omitempty"`
	PageCount       *int32  `json:"page_count,omitempty"`
	PublicationYear *int32  `json:"publication_year,omitempty"`
}

type AvailabilityResponse struct {
	StoreUUID     uuid.UUID `json:"store_uuid"`
	StoreName     string    `json:"store_name"`
	SkuUUID       uuid.UUID `json:"sku_uuid"`
	PriceInKopeks int32     `json:"price_in_kopeks"`
	StockCount    int32     `json:"stock_count"`
}

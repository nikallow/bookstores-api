package stores

import "github.com/google/uuid"

type CreateStoreRequest struct {
	Name    string `json:"name"    validate:"required"`
	Address string `json:"address" validate:"required"`
}

type UpdateStoreRequest struct {
	Name    string `json:"name"    validate:"required"`
	Address string `json:"address" validate:"required"`
}

type StoreResponse struct {
	UUID    uuid.UUID `json:"uuid"`
	Name    string    `json:"name"`
	Address string    `json:"address"`
}

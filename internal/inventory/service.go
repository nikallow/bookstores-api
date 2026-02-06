package inventory

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	repo "github.com/nikallow/bookstores-api/internal/adapters/postgres/sqlc"
	"github.com/nikallow/bookstores-api/internal/middleware"
)

var (
	ErrStoreNotFound     = errors.New("store not found")
	ErrBookNotFound      = errors.New("book not found")
	ErrSKUNotFound       = errors.New("sku not found")
	ErrSKUAlreadyExists  = errors.New("this book already exists in this store")
	ErrInsufficientStock = errors.New("insufficient stock")
)

type Service interface {
	CreateSKU(ctx context.Context, params CreateSKURequest) (repo.Sku, error)
	GetSKU(ctx context.Context, skuUUID uuid.UUID) (repo.GetSKUByUUIDRow, error)
	UpdateSKUPrice(ctx context.Context, skuUUID uuid.UUID, newPrice int32) (repo.Sku, error)
	AdjustSKUStock(ctx context.Context, skuUUID uuid.UUID, changeBy int32) (repo.Sku, error)
}

type service struct {
	repo repo.Querier
	db   *pgx.Conn
}

func NewService(repo repo.Querier, db *pgx.Conn) Service {
	return &service{repo: repo, db: db}
}

// CreateSKU - POST /skus
func (s *service) CreateSKU(ctx context.Context, params CreateSKURequest) (repo.Sku, error) {
	log := middleware.LoggerFromContext(ctx)

	if _, err := s.repo.GetBookByID(ctx, params.BookID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repo.Sku{}, ErrBookNotFound
		}
		log.Error("Failed to check book existence", "error", err)
		return repo.Sku{}, err
	}

	store, err := s.repo.GetStoreByUUID(ctx, uuidToPgUUID(params.StoreUUID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repo.Sku{}, ErrStoreNotFound
		}
		log.Error("Failed to get store by uuid", "error", err)
		return repo.Sku{}, err
	}

	_, err = s.repo.GetSKUByBookAndStore(ctx, repo.GetSKUByBookAndStoreParams{
		BookID:  params.BookID,
		StoreID: store.ID,
	})
	if err == nil {
		return repo.Sku{}, ErrSKUAlreadyExists
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		log.Error("Failed to check sku existence", "error", err)
		return repo.Sku{}, err
	}

	sku, err := s.repo.CreateSKU(ctx, repo.CreateSKUParams{
		BookID:        params.BookID,
		StoreID:       store.ID,
		PriceInKopeks: params.PriceInKopeks,
		StockCount:    params.StockCount,
	})
	if err != nil {
		log.Error("failed to create sku", "error", err)
		return repo.Sku{}, err
	}

	log.Info("SKU created successfully", "sku_id", sku.ID)
	return sku, nil
}

func (s *service) GetSKU(ctx context.Context, skuUUID uuid.UUID) (repo.GetSKUByUUIDRow, error) {
	row, err := s.repo.GetSKUByUUID(ctx, uuidToPgUUID(skuUUID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repo.GetSKUByUUIDRow{}, ErrSKUNotFound
		}
		return repo.GetSKUByUUIDRow{}, err
	}
	return row, nil
}

func (s *service) UpdateSKUPrice(ctx context.Context, skuUUID uuid.UUID, newPrice int32) (repo.Sku, error) {
	log := middleware.LoggerFromContext(ctx)

	_, err := s.GetSKU(ctx, skuUUID)
	if err != nil {
		return repo.Sku{}, err
	}

	sku, err := s.repo.UpdateSKUPrice(ctx, repo.UpdateSKUPriceParams{
		Uuid:          uuidToPgUUID(skuUUID),
		PriceInKopeks: newPrice,
	})
	if err != nil {
		log.Error("Failed to update sku price", "error", err, "sku_uuid", skuUUID)
		return repo.Sku{}, err
	}
	return sku, nil
}

func (s *service) AdjustSKUStock(ctx context.Context, skuUUID uuid.UUID, changeBy int32) (repo.Sku, error) {
	log := middleware.LoggerFromContext(ctx)

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return repo.Sku{}, err
	}
	defer tx.Rollback(ctx)

	qtx := repo.New(tx)

	skuRow, err := qtx.GetSKUByUUID(ctx, uuidToPgUUID(skuUUID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Error("SKU not found", "error", err)
			return repo.Sku{}, ErrSKUNotFound
		}
		log.Error("Failed to get SKU by uuid", "error", err)
		return repo.Sku{}, err
	}

	if skuRow.Sku.StockCount+changeBy < 0 {
		log.Error("SKU stock count is negative", "error", err)
		return repo.Sku{}, ErrInsufficientStock
	}

	updatedSKU, err := qtx.AdjustSKUStock(ctx, repo.AdjustSKUStockParams{
		Uuid:     uuidToPgUUID(skuUUID),
		ChangeBy: changeBy,
	})
	if err != nil {
		log.Error("Failed to adjust sku stock", "error", err, "sku_uuid", skuUUID)
		return repo.Sku{}, err
	}

	return updatedSKU, tx.Commit(ctx)
}

func uuidToPgUUID(u uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: u, Valid: true}
}

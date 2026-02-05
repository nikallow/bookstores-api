package stores

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	repo "github.com/nikallow/bookstores-api/internal/adapters/postgres/sqlc"
	"github.com/nikallow/bookstores-api/internal/middleware"
)

var (
	ErrStoreNotFound = errors.New("store not found")
)

type Service interface {
	Create(ctx context.Context, name, address string) (repo.Store, error)
	List(ctx context.Context) ([]repo.Store, error)
	GetByUUID(ctx context.Context, id uuid.UUID) (repo.Store, error)
	Update(ctx context.Context, id uuid.UUID, name, address string) (repo.Store, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type service struct {
	repo repo.Querier
}

func NewService(repo repo.Querier) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, name, address string) (repo.Store, error) {
	log := middleware.LoggerFromContext(ctx)

	store, err := s.repo.CreateStore(ctx, repo.CreateStoreParams{
		Name:    name,
		Address: address,
	})
	if err != nil {
		log.Error("Failed to create store", "error", err)
		return repo.Store{}, fmt.Errorf("failed to create store: %w", err)
	}

	log.Info("Store created successfully", "store_uuid", store.Uuid)
	return store, nil
}

func (s *service) List(ctx context.Context) ([]repo.Store, error) {
	log := middleware.LoggerFromContext(ctx)

	stores, err := s.repo.ListStores(ctx)
	if err != nil {
		log.Error("Failed to list stores", "error", err)
		return nil, fmt.Errorf("failed to list stores: %w", err)
	}

	return stores, nil
}

func (s *service) GetByUUID(ctx context.Context, id uuid.UUID) (repo.Store, error) {
	log := middleware.LoggerFromContext(ctx)

	store, err := s.repo.GetStoreByUUID(ctx, uuidToPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repo.Store{}, ErrStoreNotFound
		}
		log.Error("Failed to get store by UUID", "error", err, "store_uuid", id)
		return repo.Store{}, fmt.Errorf("failed to get store: %w", err)
	}
	return store, nil
}

func (s *service) Update(ctx context.Context, id uuid.UUID, name, address string) (repo.Store, error) {
	log := middleware.LoggerFromContext(ctx)

	store, err := s.repo.UpdateStore(ctx, repo.UpdateStoreParams{
		Uuid:    uuidToPgUUID(id),
		Name:    name,
		Address: address,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repo.Store{}, ErrStoreNotFound
		}
		log.Error("Failed to update store", "error", err, "store_uuid", id)
		return repo.Store{}, fmt.Errorf("failed to update store: %w", err)
	}

	log.Info("Store updated successfully", "store_uuid", store.Uuid)
	return store, nil
}

func (s *service) Delete(ctx context.Context, id uuid.UUID) error {
	log := middleware.LoggerFromContext(ctx)

	err := s.repo.SoftDeleteStore(ctx, uuidToPgUUID(id))
	if err != nil {
		log.Error("Failed to soft delete store", "error", err, "store_uuid", id)
		return fmt.Errorf("failed to delete store: %w", err)
	}

	log.Info("Store soft-deleted successfully", "store_uuid", id)
	return nil
}

func uuidToPgUUID(u uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: u, Valid: true}
}

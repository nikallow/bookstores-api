package books

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	repo "github.com/nikallow/bookstores-api/internal/adapters/postgres/sqlc"
	"github.com/nikallow/bookstores-api/internal/middleware"
)

var ErrBookNotFound = errors.New("book not found")

type Service interface {
	Create(ctx context.Context, params CreateBookRequest) (repo.Book, error)
	List(ctx context.Context) ([]repo.Book, error)
	GetByID(ctx context.Context, id int64) (repo.Book, error)
	Search(ctx context.Context, query string) ([]repo.Book, error)
	GetAvailability(ctx context.Context, bookID int64) ([]repo.ListBookAvailabilityRow, error)
}

type service struct {
	repo repo.Querier
}

func NewService(repo repo.Querier) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, params CreateBookRequest) (repo.Book, error) {
	log := middleware.LoggerFromContext(ctx)

	book, err := s.repo.CreateBook(ctx, repo.CreateBookParams{
		Isbn:            stringToPgTextp(params.ISBN),
		Title:           params.Title,
		Author:          params.Author,
		Description:     stringToPgTextp(params.Description),
		PageCount:       int32ToPgInt4p(params.PageCount),
		PublicationYear: int32ToPgInt4p(params.PublicationYear),
	})
	if err != nil {
		log.Error("Failed to create or update book", "error", err)
		return repo.Book{}, err
	}
	return book, nil
}

func (s *service) List(ctx context.Context) ([]repo.Book, error) {
	return s.repo.ListBooks(ctx)
}

func (s *service) GetByID(ctx context.Context, id int64) (repo.Book, error) {
	book, err := s.repo.GetBookByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repo.Book{}, ErrBookNotFound
		}
		return repo.Book{}, err
	}
	return book, nil
}

func (s *service) Search(ctx context.Context, query string) ([]repo.Book, error) {
	return s.repo.SearchBooks(ctx, pgtype.Text{String: query, Valid: true})
}

func (s *service) GetAvailability(ctx context.Context, bookID int64) ([]repo.ListBookAvailabilityRow, error) {
	if _, err := s.GetByID(ctx, bookID); err != nil {
		return nil, err
	}
	return s.repo.ListBookAvailability(ctx, bookID)
}

func stringToPgTextp(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}
func int32ToPgInt4p(i *int32) pgtype.Int4 {
	if i == nil {
		return pgtype.Int4{Valid: false}
	}
	return pgtype.Int4{Int32: *i, Valid: true}
}

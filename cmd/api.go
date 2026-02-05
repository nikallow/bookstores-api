package main

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nikallow/bookstores-api/internal/stores"
)

type APIDependencies struct {
	Logger           *slog.Logger
	StoreHandler     *stores.Handler
	InventoryHandler *inventory.Handler
}

func MountAPI(deps *APIDependencies) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("OK"))
	})

	r.Route("/stores", func(r chi.Router) {
		r.Post("/", deps.StoreHandler.CreateStore)
		r.Get("/", deps.StoreHandler.ListStores)
		r.Get("/{storeUUID}", deps.StoreHandler.GetStore)
		r.Put("/{storeUUID}", deps.StoreHandler.UpdateStore)
		r.Delete("/{storeUUID}", deps.StoreHandler.DeleteStore)
	})

	return r
}

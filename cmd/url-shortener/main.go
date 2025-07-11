package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"log/slog"
	"net/http"
	"os"
	"url-shortener/internal/config"
	"url-shortener/internal/http-server/handlers/redirect"
	"url-shortener/internal/http-server/handlers/url/delete"
	"url-shortener/internal/http-server/handlers/url/save"
	"url-shortener/internal/http-server/middleware/logger"
	"url-shortener/internal/logging"
	"url-shortener/internal/logging/sl"
	"url-shortener/internal/storage/mysql"
)

func main() {
	cfg := config.MustLoad()
	lg, err := logging.Setup(cfg.Env)
	if err != nil {
		log.Fatal(err)
	}

	lg.Info("starting url-shortener", slog.String("env", cfg.Env))

	storage, err := mysql.New(cfg.DbConnectionString)
	if err != nil {
		lg.Error("failed to init storage", sl.Error(err))
		os.Exit(1)
	}

	lg.Info("storage was initialized")

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(logger.New(lg))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Post("/url", save.New(lg, storage))
	router.Delete("/url/{alias}", delete.New(lg, storage))
	router.Get("/{alias}", redirect.New(lg, storage))

	lg.Info("starting server", slog.String("address", cfg.Address))

	server := http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HttpServerConfig.Timeout,
		WriteTimeout: cfg.HttpServerConfig.Timeout,
		IdleTimeout:  cfg.HttpServerConfig.IdleTimeout,
	}

	if err := server.ListenAndServe(); err != nil {
		lg.Error("failed to start server", sl.Error(err))
	}

	lg.Error("server stopped")
}

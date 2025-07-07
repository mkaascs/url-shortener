package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"log/slog"
	"os"
	"url-shortener/internal/config"
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

	_, err = mysql.New(cfg.DbConnectionString)
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
}

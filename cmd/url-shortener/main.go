package main

import (
	"log"
	"url-shortener/internal/config"
	"url-shortener/internal/logging"
)

func main() {
	cfg := config.MustLoad()
	lg, err := logging.Setup(cfg.Env)
	if err != nil {
		log.Fatal(err)
	}

	lg.Debug(cfg.DbConnectionString)
}

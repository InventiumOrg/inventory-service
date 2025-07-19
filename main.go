package main

import (
	"context"

	"inventory-service/api"
	"inventory-service/config"
	"log/slog"

	// models "inventory-service/models/sqlc"
	"os"

	"github.com/jackc/pgx/v5"
)

func main() {
  router := api.NewServer()

  config, err := config.LoadConfig(".")
  if err != nil {
    slog.Error("Failed to load config: ", slog.Any("ERROR", err))
    os.Exit(1)
  }

  slog.Info("Connecting to database", slog.String("db_source", config.DBSource))
  conn, err := pgx.Connect(context.Background(), config.DBSource)
  if err != nil {
    slog.Error("Unable to connect to database: ", slog.Any("ERROR", err))
    os.Exit(1)
  }
  slog.Info("Connected to database successfully")
  defer conn.Close(context.Background())

  // q := models.New(conn)
  router.Run(":13740")
  

}
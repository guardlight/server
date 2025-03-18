package testcontainers

import (
	"context"
	"time"

	"github.com/guardlight/server/internal/essential/config"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
)

const (
	postgresVersion = "postgres:16-alpine"
)

func NewPostgresContainer(ctx context.Context) (*postgres.PostgresContainer, error) {
	zap.S().Info("testcontainer: Starting postgres container")

	postgresContainer, err := postgres.Run(ctx,
		postgresVersion,
		postgres.WithDatabase(config.Get().Database.Name),
		postgres.WithUsername(config.Get().Database.User),
		postgres.WithPassword(config.Get().Database.Password),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(10*time.Second)),
	)

	if err != nil {
		zap.S().Fatalw("failed to start container", "error", err)
		return nil, err
	}

	return postgresContainer, nil
}

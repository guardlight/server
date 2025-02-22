package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/guardlight/server/internal/essential/config"
	"github.com/guardlight/server/internal/essential/logging"
	"github.com/guardlight/server/internal/essential/testcontainers"
	"github.com/guardlight/server/internal/health"
	"github.com/guardlight/server/pkg/database"
	"github.com/guardlight/server/pkg/router"
	"go.uber.org/zap"
)

func init() {
	env := getEnv("environment", "")
	confFileDir := getEnv("env_file_dir", "")

	getEnvFile := func() string {
		switch env {
		case "development":
			return confFileDir + "env-development.yaml"
		// case "staging":
		// 	return confFileDir + "env-staging.yaml"
		case "production":
			return confFileDir + "env-production.yaml"
		default:
			panic("ENVIRONMENT variable not set")
		}
	}

	// Setup the correct logging
	logging.SetupLogging(env)

	config.SetupConfig(getEnvFile())

}

func getEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = fallback
	}
	return value
}

func main() {

	GLAdapters()

	dbUrl := config.Get().Database.Url
	if config.Get().IsDevelopment() {
		zap.S().Info("Starting staging cockroach database container")
		ctx, ctxCancel := context.WithTimeout(context.Background(), time.Hour)
		defer ctxCancel()
		csqlContainer, err := testcontainers.NewCockroachSQLContainer(ctx)
		if err != nil {
			zap.S().Fatalw("database container cannot start", "error", err)
		}
		dbUrl = csqlContainer.GetDSN()
		zap.S().Infow("starting staging database", "url", dbUrl)
	}

	// Database
	_ = database.InitDatabase(dbUrl)

	// Repositories
	// resultsRepository := results.NewResultsRepository(db)

	// Controller Groups
	mainRouter := router.NewRouter(logging.GetLogger())
	baseGroup := mainRouter.Group("")

	// Controllers
	health.NewHealthController(baseGroup)

	// Start the server
	go router.LiveOrLetDie(mainRouter)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	zap.S().Info("Shutting down api...")

	// Close all database connection etc....

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	router.LetDie(ctx)

	// catching ctx.Done(). timeout of 5 seconds.
	<-ctx.Done()
	zap.S().Info("Shutting down timeout reached")

	zap.S().Info("Server exiting")
}

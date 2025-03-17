package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/guardlight/server/internal/analysismanager"
	"github.com/guardlight/server/internal/api"
	"github.com/guardlight/server/internal/api/analysisapi"
	"github.com/guardlight/server/internal/essential/config"
	"github.com/guardlight/server/internal/essential/logging"
	"github.com/guardlight/server/internal/essential/testcontainers"
	"github.com/guardlight/server/internal/health"
	"github.com/guardlight/server/internal/infrastructure/database"
	"github.com/guardlight/server/internal/infrastructure/messaging"
	"github.com/guardlight/server/internal/jobmanager"
	"github.com/guardlight/server/internal/natsclient"
	"github.com/guardlight/server/internal/orchestrator"
	"github.com/guardlight/server/internal/scheduler"
	"github.com/guardlight/server/servers/natsmessaging"
	"go.uber.org/zap"
)

func init() {
	env := getEnv("environment", "")
	confFileDir := getEnv("env_file_dir", "../")

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
	quit := make(chan os.Signal, 1)

	loc, err := time.LoadLocation("Europe/Amsterdam")
	if err != nil {
		zap.S().Errorw("Could not load timezone", "error", err)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	}

	GlExternalServers()

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

	// Messaging
	ncon := messaging.InitNats(natsmessaging.GetNatsUrl(), natsmessaging.GetServer())
	GLAdapters(ncon)

	// Database
	db := database.InitDatabase(dbUrl)

	// Repositories
	jmr := jobmanager.NewJobManagerRepository(db)
	amr := analysismanager.NewAnalysisManagerRepository(db)

	// Controller Groups
	mainRouter := api.NewRouter(logging.GetLogger())
	baseGroup := mainRouter.Group("")

	// Services
	nc := natsclient.NewNatsClient(ncon)
	jm := jobmanager.NewJobMananger(jmr)
	sch, err := scheduler.NewScheduler(loc)
	if err != nil {
		zap.S().Errorw("Could not create scheduler", "error", err)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	}
	_, err = orchestrator.NewOrchestrator(jm, sch.Gos, nc)
	if err != nil {
		zap.S().Errorw("Could not create orhestrator", "error", err)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	}
	am := analysismanager.NewAnalysisManangerRequester(jm, amr)
	_ = analysismanager.NewAnalysisManagerAllocator(ncon, amr, jm)

	// Controllers
	health.NewHealthController(baseGroup)
	analysisapi.NewAnalysisRequestController(baseGroup, am)

	// Start the server
	go api.LiveOrLetDie(mainRouter)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	zap.S().Info("Shutting down server...")

	// Close all database connection etc....

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	api.LetDie(ctx)
	sch.Gos.Shutdown()

	// catching ctx.Done(). timeout of 5 seconds.
	<-ctx.Done()
	zap.S().Info("Shutting down timeout reached")

	zap.S().Info("Server exiting")
}

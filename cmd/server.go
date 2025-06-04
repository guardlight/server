package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/guardlight/server/internal/analysismanager"
	"github.com/guardlight/server/internal/auth"
	"github.com/guardlight/server/internal/essential/config"
	"github.com/guardlight/server/internal/essential/logging"
	"github.com/guardlight/server/internal/essential/testcontainers"
	"github.com/guardlight/server/internal/health"
	"github.com/guardlight/server/internal/infrastructure/database"
	"github.com/guardlight/server/internal/infrastructure/http"
	"github.com/guardlight/server/internal/infrastructure/messaging"
	"github.com/guardlight/server/internal/jobmanager"
	"github.com/guardlight/server/internal/natsclient"
	"github.com/guardlight/server/internal/orchestrator"
	"github.com/guardlight/server/internal/parser"
	"github.com/guardlight/server/internal/scheduler"
	"github.com/guardlight/server/internal/ssemanager"
	"github.com/guardlight/server/internal/theme"
	"github.com/guardlight/server/servers/natsmessaging"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

func Server() {
	quit := make(chan os.Signal, 1)

	loc, err := time.LoadLocation(config.Get().Timezone)
	if err != nil {
		zap.S().Errorw("Could not load timezone", "error", err)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	}
	os.Setenv("TZ", "")

	dsn := config.Get().GetDbDsn()
	if config.Get().IsDevelopment() {
		ctx, ctxCancel := context.WithTimeout(context.Background(), time.Hour)
		defer ctxCancel()
		csqlContainer, err := testcontainers.NewPostgresContainer(ctx)
		if err != nil {
			zap.S().Fatalw("database container cannot start", "error", err)
		}
		dsn, err = csqlContainer.ConnectionString(ctx, "timezone=Europe/Amsterdam")
		if err != nil {
			zap.S().Fatalw("cannot get connection string", "error", err)
		}
		zap.S().Infow("starting staging database", "url", dsn)
	}
	// Messaging
	var ncon *nats.Conn
	if config.Get().Nats.Server == "" {
		GlNatsServer()
		ncon = messaging.InitNatsInProcess(natsmessaging.GetServer())
	} else {
		ncon = messaging.InitNats()
	}

	GLAdapters(ncon)

	// Database
	db := database.InitDatabase(dsn)

	// Repositories
	jmr := jobmanager.NewJobManagerRepository(db)
	amr := analysismanager.NewAnalysisManagerRepository(db)
	tsr := theme.NewThemeRepository(db)

	// Controller Groups
	mainRouter := http.NewRouter(logging.GetLogger())
	baseGroup := mainRouter.Group("")

	// Services
	nc := natsclient.NewNatsClient(ncon)
	jm := jobmanager.NewJobMananger(jmr)
	sch, err := scheduler.NewScheduler(loc)
	if err != nil {
		zap.S().Errorw("Could not create scheduler", "error", err)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	}
	ssem := ssemanager.NewSseMananger()
	_, err = orchestrator.NewOrchestrator(jm, sch.Gos, nc)
	if err != nil {
		zap.S().Errorw("Could not create orhestrator", "error", err)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	}
	ts := theme.NewThemeService(tsr)
	ars := analysismanager.NewAnalysisResultService(amr, amr, ts)
	am := analysismanager.NewAnalysisManangerRequester(jm, amr, ssem)

	_ = analysismanager.NewAnalysisManagerAllocator(ncon, amr, jm, ssem)

	// Controllers
	health.NewHealthController(baseGroup)
	analysismanager.NewAnalysisRequestController(baseGroup, am, ars)
	parser.NewParserController(baseGroup)
	theme.NewThemeController(baseGroup, ts)
	auth.NewAuthenticationController(baseGroup)

	ssemanager.NewSseController(baseGroup, ssem)

	if config.Get().IsDevelopment() {
		database.LoadMockData(db)
	}

	// Start the server
	go http.LiveOrLetDie(mainRouter)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	zap.S().Info("Shutting down server...")

	// Close all database connection etc....

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	http.LetDie(ctx)
	sch.Gos.Shutdown()

	// catching ctx.Done(). timeout of 5 seconds.
	<-ctx.Done()
	zap.S().Info("Shutting down timeout reached")

	zap.S().Info("Server exiting")
}

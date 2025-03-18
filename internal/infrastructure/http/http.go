package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/guardlight/server/internal/essential/config"
	"github.com/ulule/limiter/v3"
	ginlimiter "github.com/ulule/limiter/v3/drivers/middleware/gin"
	"github.com/ulule/limiter/v3/drivers/store/memory"
	"go.uber.org/zap"
)

var (
	srv http.Server
)

func NewRouter(l *zap.Logger) *gin.Engine {
	router := gin.New()
	router.Use(useCors())
	router.Use(ginzap.Ginzap(l, time.RFC3339, true))
	router.Use(ginzap.RecoveryWithZap(l, true))
	router.Use(UseRateLimiting())

	if config.Get().IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	return router
}

func LiveOrLetDie(engine *gin.Engine) {
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", config.Get().Server.Host, config.Get().Server.Port),
		Handler: engine.Handler(),
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		zap.S().Errorw("Could not start router", "error", err)
		zap.S().Panic("Could not start router")
		return
	}
}

func LetDie(ctx context.Context) {
	if err := srv.Shutdown(ctx); err != nil {
		zap.S().Fatalw("Server Shutdown error", "error", err)
	}
}

func useCors() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     config.Get().Cors.Origins,
		AllowMethods:     []string{"GET, POST, PATCH, PUT, DELETE, OPTIONS"},
		AllowHeaders:     []string{"Accept, Accept-Encoding, Authorization, Cache-Control, Content-Type, Content-Length, Origin, X-Real-IP, X-CSRF-Token, X-Auth-Key"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}

func UseRateLimiting() gin.HandlerFunc {

	rate := limiter.Rate{
		Period: 5 * time.Minute,
		Limit:  100,
	}

	// Initialize the memory store.
	store := memory.NewStore()

	// Create a Gin middleware using the rate limiter.
	middleware := ginlimiter.NewMiddleware(limiter.New(store, rate))

	return middleware
}

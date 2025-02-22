package database

import (
	"time"

	"github.com/guardlight/server/internal/essential/config"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"moul.io/zapgorm2"
)

func InitDatabase(databaseUrl string) *gorm.DB {
	dbName := config.Get().Database.Name
	connectionString := databaseUrl + "&application_name=" + dbName + "&timezone=UTC"

	lgr := zapgorm2.New(zap.L())
	lgr.SetAsDefault()
	lgr.LogLevel = logger.LogLevel(zap.S().Level())
	gormConfig := &gorm.Config{
		Logger: lgr,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	db, err := gorm.Open(postgres.Open(connectionString), gormConfig)
	if err != nil {
		zap.S().DPanicw("Cannot open database connection", "error", err)
	}

	zap.S().Info("database: Database service configured")

	return db
}

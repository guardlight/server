package database

import (
	"github.com/go-testfixtures/testfixtures/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func LoadMockData(db *gorm.DB) {
	sqlDb, _ := db.DB()
	fixtures, err := testfixtures.New(
		testfixtures.Database(sqlDb),
		testfixtures.Dialect("postgres"),
		testfixtures.Directory("testdata/fixtures-livetest"),
		testfixtures.UseDropConstraint(),
	)

	if err != nil {
		zap.S().DPanicw("Could not make new fixtures", "error", err)
	}

	err = fixtures.Load()

	if err != nil {
		zap.S().DPanicw("Could not load fixtures", "error", err)
	}
}

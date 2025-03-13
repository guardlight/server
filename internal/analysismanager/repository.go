package analysismanager

import (
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AnalysisManagerRepository struct {
	db *gorm.DB
}

func NewAnalysisManagerRepository(db *gorm.DB) *AnalysisManagerRepository {
	if err := db.AutoMigrate(
		&AnalysisRequest{},
		&AnalysisRequestStep{},
		&RawData{},
		&Analysis{},
		&AnalysisReport{},
	); err != nil {
		zap.S().DPanicw("Problem automigrating the tables", "error", err)
	}

	return &AnalysisManagerRepository{
		db: db,
	}
}

func (amr AnalysisManagerRepository) createAnalysisRequest(analysisRequest *AnalysisRequest) error {
	if err := amr.db.Create(analysisRequest).Error; err != nil {
		zap.S().Errorw("Could not create analysis request", "error", err)
		return err
	}

	return nil
}

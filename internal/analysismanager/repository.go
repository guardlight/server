package analysismanager

import (
	"errors"

	"github.com/google/uuid"
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

func (amr AnalysisManagerRepository) updateProcessedText(ai uuid.UUID, text string) error {
	res := amr.db.
		Model(RawData{
			AnalysisRequestId: ai,
		}).
		Updates(RawData{ProcessedText: text})

	if res.Error != nil {
		zap.S().Errorw("Could not update processed text", "error", res.Error)
		return res.Error
	}

	if res.RowsAffected == 0 {
		zap.S().Errorw("No records updated", "analysis_request_id", ai)
		return errors.New("no records affected after update")
	}

	return nil
}

func (amr AnalysisManagerRepository) getAllAnalysisByAnalysisRecordId(id uuid.UUID) ([]Analysis, error) {
	var a []Analysis
	if err := amr.db.Where("analysis_request_id = ?", id).Find(&a).Error; err != nil {
		zap.S().Errorw("Could not get analysis records", "error", err)
		return nil, err
	}

	return a, nil

}

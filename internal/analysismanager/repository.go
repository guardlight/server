package analysismanager

import (
	"errors"

	"github.com/google/uuid"
	"github.com/samber/lo"
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

func (amr AnalysisManagerRepository) updateAnalysisJobs(ai uuid.UUID, jbs []SingleJobProgress) error {
	res := amr.db.
		Model(Analysis{
			Id: ai,
		}).
		Updates(Analysis{Jobs: jbs})

	if res.Error != nil {
		zap.S().Errorw("Could not update analysis jobs", "error", res.Error)
		return res.Error
	}

	if res.RowsAffected == 0 {
		zap.S().Errorw("No records updated", "analysis_request_id", ai)
		return errors.New("no records affected after update")
	}
	return nil
}

func (amr AnalysisManagerRepository) updateAnalysisJobProgress(aid uuid.UUID, jid uuid.UUID, status AnalysisStatus, content []string, score float32) error {
	a := Analysis{Id: aid}
	if err := amr.db.First(&a).Error; err != nil {
		return err
	}

	newJs := lo.Map(a.Jobs, func(s SingleJobProgress, _ int) SingleJobProgress {
		if s.JobId == jid {
			return SingleJobProgress{
				JobId:  jid,
				Status: status,
			}
		}
		return s
	})

	newCon := append(a.Content, content...)

	completedJobs := len(lo.Filter(newJs, func(j SingleJobProgress, _ int) bool { return j.Status == AnalysisFinished }))

	newSc := (a.Score*float32(completedJobs-1) + score) / float32(completedJobs)

	newStatus := func() AnalysisStatus {
		if completedJobs == len(newJs) {
			return AnalysisFinished
		} else {
			return AnalysisInprogress
		}
	}()

	resp := amr.db.Model(&a).Updates(Analysis{
		Jobs:    newJs,
		Content: newCon,
		Score:   newSc,
		Status:  newStatus,
	})

	if resp.Error != nil {
		return resp.Error
	}

	return nil
}

func (amr AnalysisManagerRepository) getAnalysesByUserId(id uuid.UUID) ([]AnalysisRequest, error) {
	var ars []AnalysisRequest
	if err := amr.db.Model(AnalysisRequest{UserId: id}).Find(&ars).Error; err != nil {
		zap.S().Errorw("Could not get analyses", "user_id", id)
		return nil, err
	}
	return ars, nil
}

func (amr AnalysisManagerRepository) getAnalysById(uid uuid.UUID, aid uuid.UUID) (AnalysisRequest, error) {
	var ar AnalysisRequest

	resp := amr.db.
		Model(AnalysisRequest{}).
		Where("user_id = ? AND id = ?", uid, aid).
		First(&ar)

	if err := resp.Error; err != nil {
		zap.S().Errorw("Could not get analyses", "user_id", uid, "analysis_id", aid)
		return AnalysisRequest{}, err
	}
	return ar, nil
}

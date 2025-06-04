package analysismanager

import (
	"errors"
	"math"

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

	completedJobs := lo.CountBy(newJs, func(j SingleJobProgress) bool { return j.Status == AnalysisFinished })

	newSc := (a.Score*float32(completedJobs-1) + score) / float32(completedJobs)

	newStatus := func() AnalysisStatus {
		if completedJobs == len(newJs) {
			return AnalysisFinished
		} else {
			return AnalysisInprogress
		}
	}()
	if newStatus == AnalysisFinished {
		zap.S().Infow("Analysis completed", "analysis_id", aid, "job_id", jid)
	}

	resp := amr.db.Model(&a).Updates(Analysis{
		Jobs:    newJs,
		Content: newCon,
		Score:   newSc,
		Status:  newStatus,
	})

	if resp.Error != nil {
		return resp.Error
	}

	if resp.RowsAffected == 0 {
		zap.S().Errorw("No records updated", "analysis_id", aid)
		return errors.New("no records affected after update")
	}

	return nil
}

func (amr AnalysisManagerRepository) getAnalysesByUserId(id uuid.UUID, pag Pagination) (AnalysisResultPaginated, error) {
	var totalRows int64
	amr.db.Model(AnalysisRequest{}).Count(&totalRows)

	totalPages := int(math.Ceil(float64(totalRows) / float64(pag.GetLimit())))

	var ars []AnalysisRequest
	if err := amr.db.Offset(pag.GetOffset()).Limit(pag.GetLimit()).Order("created_at DESC").Preload("Analysis").Model(AnalysisRequest{UserId: id}).Find(&ars).Error; err != nil {
		zap.S().Errorw("Could not get analyses", "user_id", id)
		return AnalysisResultPaginated{}, err
	}
	return AnalysisResultPaginated{
		Limit:      pag.GetLimit(),
		TotalPages: totalPages,
		Page:       pag.GetPage(),
		Requests:   ars,
	}, nil
}

func (amr AnalysisManagerRepository) getAnalysesByAnalysisIdAndUserId(uid, aid uuid.UUID) (AnalysisRequest, error) {
	var ar AnalysisRequest

	resp := amr.db.
		Model(AnalysisRequest{}).
		Preload("Analysis").
		Where("user_id = ? AND id = ?", uid, aid).
		First(&ar)

	if err := resp.Error; err != nil {
		zap.S().Errorw("Could not get analyses", "user_id", uid, "analysis_id", aid)
		return AnalysisRequest{}, err
	}
	return ar, nil
}

func (amr AnalysisManagerRepository) getUserIdByAnalysisId(analysisId uuid.UUID) (uuid.UUID, error) {
	var userId string

	err := amr.db.
		Table("analyses").
		Select("analysis_requests.user_id").
		Joins("LEFT JOIN analysis_requests ON analyses.analysis_request_id = analysis_requests.id").
		Where("analyses.id = ?", analysisId).
		Scan(&userId).Error

	if err != nil {
		zap.S().Errorw("Could not resolve user ID from analysis ID", "analysis_id", analysisId, "error", err)
		return uuid.Nil, err
	}

	return uuid.MustParse(userId), nil
}

func (amr AnalysisManagerRepository) updateScore(analysisId uuid.UUID, score float32) error {

	err := amr.db.
		Model(Analysis{
			Id: analysisId,
		}).
		Updates(Analysis{Score: score}).Error

	if err != nil {
		zap.S().Errorw("Could not resolve user ID from analysis ID", "analysis_id", analysisId, "error", err)
		return err
	}

	return nil
}

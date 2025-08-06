package analysismanager

import (
	"errors"
	"math"

	"github.com/google/uuid"
	"github.com/guardlight/server/internal/theme"
	"github.com/samber/lo"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
		Updates(RawData{ProcessedText: text, Content: []byte{}}) // Remove the raw data and save the processed text

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

func (amr AnalysisManagerRepository) getAllAnalysisById(aid uuid.UUID) (Analysis, error) {
	var a Analysis
	if err := amr.db.Where("id = ?", aid).First(&a).Error; err != nil {
		zap.S().Errorw("Could not get analysis record", "error", err)
		return Analysis{}, err
	}

	return a, nil
}

func (amr AnalysisManagerRepository) getReporterKeyByAnalysisId(aid uuid.UUID) (string, error) {
	a := Analysis{Id: aid}
	if err := amr.db.First(&a).Error; err != nil {
		zap.S().Errorw("Could not get analysis record", "error", err, "id", aid)
		return "", err
	}

	t := theme.Theme{Id: a.ThemeId}
	if err := amr.db.First(&t).Error; err != nil {
		zap.S().Errorw("Could not get theme record", "error", err, "id", a.ThemeId)
		return "", err
	}

	return t.Reporter.Key, nil
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

func (amr AnalysisManagerRepository) updateAllAnalysesStatusByAnalysisRequestId(arid uuid.UUID, status AnalysisStatus) error {
	res := amr.db.
		Model(Analysis{}).
		Where("analysis_request_id = ?", arid).
		Updates(Analysis{Status: status})

	if res.Error != nil {
		zap.S().Errorw("Could not update analysis status", "error", res.Error)
		return res.Error
	}

	if res.RowsAffected == 0 {
		zap.S().Errorw("No records updated", "analysis_request_id", arid)
		return errors.New("no records affected after update")
	}
	return nil
}

func (amr AnalysisManagerRepository) updateAnalysisJobProgress(aid uuid.UUID, jid uuid.UUID, status AnalysisStatus, content []string) (bool, error) {
	a := Analysis{Id: aid}
	if err := amr.db.First(&a).Error; err != nil {
		return false, err
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
		Status:  newStatus,
	})

	if resp.Error != nil {
		return false, resp.Error
	}

	if resp.RowsAffected == 0 {
		zap.S().Errorw("No records updated", "analysis_id", aid)
		return false, errors.New("no records affected after update")
	}

	return newStatus == AnalysisFinished, nil
}

func (amr AnalysisManagerRepository) getAnalysesByUserId(id uuid.UUID, pag Pagination, catType, catCat, query, sc string) (AnalysisResultPaginated, error) {

	var fuzzCatType = "%" + catType + "%"
	var fuzzCatCat = "%" + catCat + "%"
	var fuzzQuery = "%" + query + "%"

	var dbQ = amr.db.Model(AnalysisRequest{UserId: id})

	if len(catType) > 0 {
		dbQ = dbQ.Where("content_type ILIKE ?", fuzzCatType)
	}
	if len(catCat) > 0 {
		dbQ = dbQ.Where("category ILIKE ?", fuzzCatCat)
	}
	if len(query) > 0 {
		dbQ = dbQ.Where("title ILIKE ?", fuzzQuery)
	}

	if sc == "BAD" {
		dbQ = dbQ.Where("NOT EXISTS (SELECT 1 FROM analyses WHERE analyses.analysis_request_id = analysis_requests.id AND analyses.score > ?)", -1)
	}
	if sc == "MIXED" {
		dbQ = dbQ.Where("EXISTS (SELECT 1 FROM analyses WHERE analyses.analysis_request_id = analysis_requests.id AND analyses.score > ?)", 0)
	}
	if sc == "GOOD" {
		dbQ = dbQ.Where("NOT EXISTS (SELECT 1 FROM analyses WHERE analyses.analysis_request_id = analysis_requests.id AND analyses.score < ?)", 1)
	}

	dbQ = dbQ.Session(&gorm.Session{})

	var ars []AnalysisRequest
	if err := dbQ.Offset(pag.GetOffset()).Limit(pag.GetLimit()).Order("created_at DESC").Preload("Analysis").Find(&ars).Error; err != nil {
		zap.S().Errorw("Could not get analyses", "user_id", id)
		return AnalysisResultPaginated{}, err
	}

	var totalRows int64
	if err := dbQ.Debug().Count(&totalRows).Error; err != nil {
		zap.S().Errorw("Could not get rows", "err", err)
	}

	totalPages := int(math.Ceil(float64(totalRows) / float64(pag.GetLimit())))

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

func (amr AnalysisManagerRepository) getAnalysisRequestIdByHash(hash string) (uuid.UUID, error) {
	var rawData RawData

	result := amr.db.Model(&RawData{}).Where("hash = ?", hash).First(&rawData)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return uuid.Nil, nil
		}
		zap.S().Errorw("Could not find raw data by hash", "raw_hash", hash, "error", result.Error)
		return uuid.Nil, result.Error
	}

	return rawData.AnalysisRequestId, nil
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

func (amr AnalysisManagerRepository) deleteAnalysisRequestById(arid, uid uuid.UUID) error {
	var recCount int64
	if err := amr.db.Where("id = ? and user_id = ?", arid, uid).Model(AnalysisRequest{}).Count(&recCount).Error; err != nil {
		zap.S().Errorw("Could not get analysis record for request and user id", "error", err)
		return err
	}

	if recCount == 0 {
		return errors.New("no record found for request id and user id")
	}

	if err := amr.db.Select(clause.Associations).Delete(&AnalysisRequest{Id: arid}).Error; err != nil {
		zap.S().Errorw("Could not delete analysis request", "error", err)
		return err
	}

	return nil
}

package analysismanager

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"github.com/guardlight/server/internal/essential/config"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type RawDataManager struct {
	tc taskCreater
	db *gorm.DB
}

type taskCreater interface {
	NewJob(jobDefinition gocron.JobDefinition, task gocron.Task, options ...gocron.JobOption) (gocron.Job, error)
}

func NewRawDataManager(tc taskCreater, db *gorm.DB) *RawDataManager {
	rdm := &RawDataManager{
		tc: tc,
		db: db,
	}

	zap.S().Infow("Processed text exporting", "status", config.Get().Data.ExportProcessedText)
	if config.Get().Data.ExportProcessedText {
		_, err := tc.NewJob(
			gocron.DurationJob(
				10*time.Second,
			),
			gocron.NewTask(rdm.exportProcessedTextToFile),
			gocron.WithSingletonMode(gocron.LimitModeReschedule),
		)
		if err != nil {
			return nil
		}
	}

	return rdm
}

type exportData struct {
	Id            uuid.UUID `gorm:"column:id"`
	Category      string    `gorm:"column:category"`
	Title         string    `gorm:"column:title"`
	ProcessedText string    `gorm:"column:processed_text"`
}

func (rdm *RawDataManager) exportProcessedTextToFile() {

	zap.S().Infow("Starting exporting raw data processed text to file")
	var exportDatas []exportData

	err := rdm.db.
		Table("analysis_requests").
		Select("analysis_requests.id, analysis_requests.category, analysis_requests.title, raw_data.processed_text").
		Joins("LEFT JOIN raw_data ON raw_data.analysis_request_id = analysis_requests.id").
		Where("raw_data.processed_text <> ?", "EXPORTED").
		Limit(1).
		Scan(&exportDatas).Error
	if err != nil {
		zap.S().Errorw("Problem getting export rawdata", "error", err)
		return
	}

	if len(exportDatas) == 0 {
		zap.S().Infow("No rawdata processed text to export")
	}

	for _, ed := range exportDatas {
		filePath := makeFilePath(config.Get().Data.ExportPath, ed.Category, stripLeadingArticle(ed.Title))
		err := writeToTextToFile(filePath, ed.ProcessedText)
		if err != nil {
			zap.S().Errorw("Could not write text to file", "error", err, "analysis_request_id", ed.Id, "filepath", filePath)
			return
		}

		// Update raw_data to clear processed and content text.
		res := rdm.db.
			Model(&RawData{}).
			Where("analysis_request_id = ?", ed.Id).
			Updates(RawData{ProcessedText: "EXPORTED"})
		if res.Error != nil {
			zap.S().Errorw("Could not update processed text", "error", res.Error)
			return
		}

		if res.RowsAffected == 0 {
			zap.S().Errorw("No records updated", "analysis_request_id", ed.Id)
			return
		}
		zap.S().Infow("Exported and updated text", "analysis_request_id", ed.Id, "filepath", filePath)
	}

}

func writeToTextToFile(path, text string) error {
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		zap.S().Errorw("Error creating directory", "error", err, "filepath", path)
	}

	err = os.WriteFile(path, []byte(text), 0644)
	if err != nil {
		zap.S().Errorw("Error writing to path", "error", err, "filepath", path)
	}
	return nil
}

func makeFilePath(basePath, category, title string) string {
	safeTitle := buildSlugTitle(title)

	firstLetter := string([]rune(safeTitle)[0])
	secondLetter := string([]rune(safeTitle)[1])
	letterDir := fmt.Sprintf("%s%s", firstLetter, secondLetter)

	filename := fmt.Sprintf("%s.txt", safeTitle)

	safeCategory := strings.ReplaceAll(strings.ToLower(category), " ", "_")

	return filepath.Join(basePath, safeCategory, letterDir, filename)
}

func stripLeadingArticle(title string) string {
	trimmedTitle := strings.ToLower(strings.TrimSpace(title))
	for _, prefix := range []string{"the ", "a ", "an "} {
		if after, ok := strings.CutPrefix(trimmedTitle, prefix); ok {
			return after
		}
	}
	return trimmedTitle
}

func buildSlugTitle(title string) string {
	safeTitle := strings.ReplaceAll(strings.ToLower(title), " ", "_")
	var b strings.Builder
	for _, r := range safeTitle {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			b.WriteRune(r)
		}
	}
	slug := b.String()
	if len(slug) == 0 {
		return "zz"
	} else if len(slug) == 1 {
		return fmt.Sprintf("%sz", slug)

	} else {
		return slug
	}
}

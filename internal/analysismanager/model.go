package analysismanager

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/google/uuid"
)

type AnalysisRequestStepType string

const (
	Create  AnalysisRequestStepType = "create"
	Upload  AnalysisRequestStepType = "upload"
	Parse   AnalysisRequestStepType = "parse"
	Analyze AnalysisRequestStepType = "analyze"
	Report  AnalysisRequestStepType = "report"
	Done    AnalysisRequestStepType = "done"
)

type AnalysisRequestStepStatus string

const (
	Waiting    AnalysisRequestStepStatus = "waiting"
	Inprogress AnalysisRequestStepStatus = "inprogress"
	Finished   AnalysisRequestStepStatus = "finished"
	Error      AnalysisRequestStepStatus = "error"
)

type AnalysisRequest struct {
	Id                   uuid.UUID             `gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	UserId               uuid.UUID             `gorm:"column:user_id"`
	Title                string                `gorm:"column:title"`
	AnalysisRequestSteps []AnalysisRequestStep `gorm:"foreignKey:AnalysisRequestId"`
	RawData              RawData               `gorm:"foreignKey:AnalysisRequestId"`
	Analysis             []Analysis            `gorm:"foreignKey:AnalysisRequestId"`
	Report               AnalysisReport        `gorm:"foreignKey:AnalysisRequestId"`
}

type AnalysisRequestStep struct {
	Id                uuid.UUID                 `gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	AnalysisRequestId uuid.UUID                 `gorm:"column:analysis_request_id;primaryKey;type:uuid"`
	Index             int                       `gorm:"column:index"`
	StepType          AnalysisRequestStepType   `gorm:"column:step_type"`
	Status            AnalysisRequestStepStatus `gorm:"column:status"`
	StatusDescription string                    `gorm:"status_decsription"`
}

type RawData struct {
	Id                uuid.UUID `gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	AnalysisRequestId uuid.UUID `gorm:"column:analysis_request_id;primaryKey;type:uuid"`
	Hash              string    `gorm:"column:hash"`
	Content           []byte    `gorm:"column:content;type:bytea"`
	FileType          string    `gorm:"column:file_type"`
	ProcessedText     string    `gorm:"column:processed_text"`
}

type Analysis struct {
	Id                uuid.UUID `gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	AnalysisRequestId uuid.UUID `gorm:"column:analysis_request_id;primaryKey;type:uuid"`
	AnalyzerKey       string    `gorm:"column:analyzer_key"`
	ThemeId           string    `gorm:"column:theme_id"`
	Status            string    `gorm:"column:status"`
	Threshold         int       `gorm:"column:threshold"`
	Score             int       `gorm:"column:score"`
	Content           Content   `gorm:"column:content;type:jsonb"`
}

type AnalysisReport struct {
	Id                uuid.UUID       `gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	AnalysisRequestId uuid.UUID       `gorm:"column:analysis_request_id;primaryKey;type:uuid"`
	Score             int             `gorm:"column:score"`
	AnalysisSummary   AnalysisSummary `gorm:"column:analysis_summary;type:jsonb"`
}

type Content []string

func (c Content) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *Content) Scan(src interface{}) error {
	return json.Unmarshal(src.([]byte), &c)
}

type AnalysisSummary struct{}

func (as AnalysisSummary) Value() (driver.Value, error) {
	return json.Marshal(as)
}

func (as *AnalysisSummary) Scan(src interface{}) error {
	return json.Unmarshal(src.([]byte), &as)
}

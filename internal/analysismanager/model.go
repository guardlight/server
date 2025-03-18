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

type AnalysisStatus string

const (
	AnalysisWaiting    AnalysisStatus = "waiting"
	AnalysisInprogress AnalysisStatus = "inprogress"
	AnalysisFinished   AnalysisStatus = "finished"
	AnalysisError      AnalysisStatus = "error"
)

type AnalysisRequest struct {
	Id                   uuid.UUID             `gorm:"column:id;primaryKey;type:uuid;default:gen_random_uuid()"`
	UserId               uuid.UUID             `gorm:"column:user_id"`
	Title                string                `gorm:"column:title"`
	AnalysisRequestSteps []AnalysisRequestStep `gorm:"foreignKey:AnalysisRequestId"`
	RawData              RawData               `gorm:"foreignKey:AnalysisRequestId"`
	Analysis             []Analysis            `gorm:"foreignKey:AnalysisRequestId"`
}

type AnalysisRequestStep struct {
	Id                uuid.UUID                 `gorm:"column:id;primaryKey;type:uuid;default:gen_random_uuid()"`
	AnalysisRequestId uuid.UUID                 `gorm:"column:analysis_request_id;primaryKey;type:uuid"`
	Index             int                       `gorm:"column:index"`
	StepType          AnalysisRequestStepType   `gorm:"column:step_type"`
	Status            AnalysisRequestStepStatus `gorm:"column:status"`
	StatusDescription string                    `gorm:"status_decsription"`
}

type RawData struct {
	Id                uuid.UUID `gorm:"column:id;primaryKey;type:uuid;default:gen_random_uuid()"`
	AnalysisRequestId uuid.UUID `gorm:"column:analysis_request_id;primaryKey;type:uuid"`
	Hash              string    `gorm:"column:hash"`
	Content           []byte    `gorm:"column:content;type:bytea"`
	FileType          string    `gorm:"column:file_type"`
	ProcessedText     string    `gorm:"column:processed_text"`
}

type Analysis struct {
	Id                uuid.UUID      `gorm:"column:id;primaryKey;type:uuid;default:gen_random_uuid()"`
	AnalysisRequestId uuid.UUID      `gorm:"column:analysis_request_id;primaryKey;type:uuid"`
	AnalyzerKey       string         `gorm:"column:analyzer_key"`
	ThemeId           uuid.UUID      `gorm:"column:theme_id"`
	Status            AnalysisStatus `gorm:"column:status"`
	Threshold         int            `gorm:"column:threshold"`
	Score             float32        `gorm:"column:score"`
	Content           Content        `gorm:"column:content;type:jsonb"`
	Inputs            Inputs         `gorm:"column:inputs;type:jsonb"`
	Jobs              JobsProgress   `gorm:"column:jobs;type:jsonb"`
}

type AnalysisInput struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Inputs []AnalysisInput

func (c Inputs) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *Inputs) Scan(src interface{}) error {
	return json.Unmarshal(src.([]byte), &c)
}

type Content []string

func (c Content) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *Content) Scan(src interface{}) error {
	return json.Unmarshal(src.([]byte), &c)
}

type SingleJobProgress struct {
	JobId  uuid.UUID      `json:"jobId"`
	Status AnalysisStatus `json:"status"`
}
type JobsProgress []SingleJobProgress

func (as JobsProgress) Value() (driver.Value, error) {
	return json.Marshal(as)
}

func (as *JobsProgress) Scan(src interface{}) error {
	return json.Unmarshal(src.([]byte), &as)
}

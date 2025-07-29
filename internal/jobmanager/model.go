package jobmanager

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/guardlight/server/pkg/analyzercontract"
	"github.com/guardlight/server/pkg/parsercontract"
	"github.com/guardlight/server/pkg/reportercontract"
)

type JobStatus string

const (
	Queued     JobStatus = "queued"
	Inprogress JobStatus = "inprogress"
	Finished   JobStatus = "finished"
	Error      JobStatus = "error"
)

type JobType string

const (
	Parse   JobType = "parse"
	Analyze JobType = "analyze"
	Report  JobType = "report"
)

func (jt JobType) Match(s string) bool {
	return strings.HasPrefix(s, string(jt))
}

type Job struct {
	Id                uuid.UUID       `gorm:"column:id;primaryKey;type:uuid"`
	CreatedAt         time.Time       `gorm:"column:created_at"`
	UpdatedAt         time.Time       `gorm:"column:updated_at"`
	Status            JobStatus       `gorm:"column:status"`
	StatusDescription string          `gorm:"column:status_description"`
	RetryCount        int             `gorm:"column:retry_count"`
	Type              JobType         `gorm:"column:type"`
	GroupKey          string          `gorm:"column:group_key"`
	Data              json.RawMessage `gorm:"column:data;type:jsonb"`
}

type ParserJobData struct {
	Type       string                       `json:"type"`
	Topic      string                       `json:"topic"`
	Image      string                       `json:"image"`
	ParserData parsercontract.ParserRequest `json:"parserData"`
}

type AnalyzerJobData struct {
	Type         string                           `json:"type"`
	Topic        string                           `json:"topic"`
	Image        string                           `json:"image"`
	AnalyzerData analyzercontract.AnalyzerRequest `json:"analyzerData"`
}

type ReportJobData struct {
	Type         string                           `json:"type"`
	Topic        string                           `json:"topic"`
	Image        string                           `json:"image"`
	ReporterData reportercontract.ReporterRequest `json:"reporterData"`
}

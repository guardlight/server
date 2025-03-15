package jobmanager

import (
	"encoding/json"
	"strings"

	"github.com/google/uuid"
	"github.com/guardlight/server/pkg/analyzercontract"
	"github.com/guardlight/server/pkg/parsercontract"
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
)

func (jt JobType) Match(s string) bool {
	return strings.HasPrefix(s, string(jt))
}

type Job struct {
	Id                uuid.UUID       `gorm:"column:id;primaryKey;type:uuid"`
	Status            JobStatus       `gorm:"column:status"`
	StatusDescription string          `gorm:"column:status_description"`
	RetryCount        int             `gorm:"column:retry_count"`
	Type              JobType         `gorm:"column:type"`
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

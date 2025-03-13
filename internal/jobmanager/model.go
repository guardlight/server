package jobmanager

import (
	"encoding/json"

	"github.com/google/uuid"
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

type Job struct {
	Id                uuid.UUID       `gorm:"column:id;primaryKey;type:uuid"`
	Status            JobStatus       `gorm:"column:status"`
	StatusDescription string          `gorm:"column:status_description"`
	RetryCount        int             `gorm:"column:retry_count"`
	Type              JobType         `gorm:"column:type"`
	Data              json.RawMessage `gorm:"column:data;type:jsonb"`
}

type ParserJobData struct {
	Type       string      `json:"type"`
	Topic      string      `json:"topic"`
	ParserData interface{} `json:"parserData"`
}

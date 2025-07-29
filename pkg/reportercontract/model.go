package reportercontract

import "github.com/google/uuid"

type ReporterRequest struct {
	JobId      uuid.UUID `json:"jobId"`
	AnalysisId uuid.UUID `json:"analysisId"`
	Contents   []string  `json:"contents"`
}

type ReporterResponseStatus string

const (
	ReportSuccess ReporterResponseStatus = "success"
	ReportError   ReporterResponseStatus = "error"
)

type ReporterResponse struct {
	JobId      uuid.UUID              `json:"jobId"`
	AnalysisId uuid.UUID              `json:"analysisId"`
	Score      float32                `json:"score"`
	Comments   string                 `json:"comments"`
	Status     ReporterResponseStatus `json:"status"`
}

package analyzercontract

import "github.com/google/uuid"

type AnalyzerRequest struct {
	JobId      uuid.UUID       `json:"jobId"`
	AnalysisId uuid.UUID       `json:"analysisId"`
	Content    string          `json:"content"`
	Inputs     []AnalysisInput `json:"inputs"`
}

type AnalysisInput struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type AnalyzerResponseStatus string

const (
	AnalyzerSuccess AnalyzerResponseStatus = "success"
	AnalyzerError   AnalyzerResponseStatus = "error"
)

type AnalyzerResponse struct {
	JobId      uuid.UUID              `json:"jobId"`
	AnalysisId uuid.UUID              `json:"analysisId"`
	Results    []string               `json:"results"`
	Score      float32                `json:"score"`
	Status     AnalyzerResponseStatus `json:"status"`
}

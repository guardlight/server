package analyzercontract

import "github.com/google/uuid"

type AnalyzerRequest struct {
	JobId      uuid.UUID `json:"jobId"`
	AnalysisId uuid.UUID `json:"analysisId"`
	Content    string    `json:"content"`
}

type AnalyzerResponse struct {
	JobId      uuid.UUID `json:"jobId"`
	AnalysisId uuid.UUID `json:"analysisId"`
	Results    []string  `json:"results"`
}

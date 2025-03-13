package parsercontract

import "github.com/google/uuid"

type ParserRequest struct {
	JobId      uuid.UUID `json:"jobId"`
	AnalysisId uuid.UUID `json:"analysisId"`
	Content    []byte    `json:"content"`
}

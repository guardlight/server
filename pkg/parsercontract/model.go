package parsercontract

import "github.com/google/uuid"

type ParserRequest struct {
	JobId      uuid.UUID `json:"jobId"`
	AnalysisId uuid.UUID `json:"analysisId"`
	Content    []byte    `json:"content"`
}

type ParserResponseStatus string

const (
	ParseSuccess ParserResponseStatus = "success"
	ParseError   ParserResponseStatus = "error"
)

type ParserResponse struct {
	JobId      uuid.UUID            `json:"jobId"`
	AnalysisId uuid.UUID            `json:"analysisId"`
	Text       string               `json:"text"`
	Status     ParserResponseStatus `json:"status"`
}

package analysisresult

import (
	"time"

	"github.com/google/uuid"
)

type AnalysisPaginated struct {
	Limit      int        `json:"limit"`
	Page       int        `json:"page"`
	TotalPages int        `json:"totalPages"`
	Analyses   []Analysis `json:"analyses"`
}

type Analysis struct {
	Id            uuid.UUID `json:"id"`
	Title         string    `json:"title"`
	Category      string    `json:"category"`
	RequestOrigin string    `json:"requestOrigin"`
	ContentType   string    `json:"contentType"`
	Themes        []Theme   `json:"themes"`
	CreatedAt     time.Time `json:"createdAt"`
}

type Theme struct {
	Id        uuid.UUID  `json:"id"`
	Title     string     `json:"title"`
	Analyzers []Analyzer `json:"analyzers"`
	Reporter  Reporter   `json:"reporter"`
}

type Reporter struct {
	Key       string  `json:"key"`
	Threshold float32 `json:"threshold"`
}

type Analyzer struct {
	Id      uuid.UUID             `json:"id"`
	Key     string                `json:"key"`
	Name    string                `json:"name"`
	Status  string                `json:"status"`
	Score   float32               `json:"score"`
	Content []string              `json:"content"`
	Inputs  []AnalyzerInput       `json:"inputs"`
	Jobs    []AnalyzerJobProgress `json:"jobs"`
}

type AnalyzerInput struct {
	Key   string `json:"key"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

type AnalyzerJobProgress struct {
	Status string `json:"status"`
}

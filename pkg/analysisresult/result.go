package analysisresult

import "github.com/google/uuid"

type Analysis struct {
	Id          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	ContentType string    `json:"contentType"`
	Themes      []Theme   `json:"themes"`
}

type Theme struct {
	Id        uuid.UUID  `json:"id"`
	Title     string     `json:"title"`
	Analyzers []Analyzer `json:"analyzer"`
}

type Analyzer struct {
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

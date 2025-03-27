package analysisrequest

import "github.com/google/uuid"

type ContentType string

const (
	BOOK   ContentType = "book"
	MOVIE  ContentType = "movie"
	SERIES ContentType = "series"
	LYRICS ContentType = "lyrics"
)

type AnalysisRequest struct {
	Title       string      `json:"title"`
	ContentType ContentType `json:"contentType"`
	File        File        `json:"file"`
	Themes      []Theme     `json:"themes"`
}

type File struct {
	Content  []byte `json:"content"`
	Mimetype string `json:"mimetype"`
}

type Theme struct {
	Title     string     `json:"title"`
	Id        uuid.UUID  `json:"id"`
	Analyzers []Analyzer `json:"analyzers"`
}

type Analyzer struct {
	Key    string          `json:"key"`
	Inputs []AnalyzerInput `json:"inputs"`
}

type AnalyzerInput struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

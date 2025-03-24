package theme

import "github.com/google/uuid"

type Theme struct {
	Id          uuid.UUID  `gorm:"column:id;primaryKey;type:uuid;default:gen_random_uuid()"`
	UserId      uuid.UUID  `gorm:"column:user_id"`
	Title       string     `gorm:"column:title"`
	Description string     `gorm:"column:description"`
	Analyzers   []Analyzer `gorm:"column:analyzers;type:jsonb"`
}

type Analyzer struct {
	Key    string          `json:"key"`
	Inputs []AnalyzerInput `json:"inputs"`
}

type AnalyzerInput struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type ChangeStatus string

const (
	New     ChangeStatus = "new"
	Removed ChangeStatus = "removed"
	Changed ChangeStatus = "changed"
	Same    ChangeStatus = "same"
)

type ThemeDto struct {
	Id          uuid.UUID     `json:"id"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Analyzers   []AnalyzerDto `json:"analyzers"`
}

type AnalyzerDto struct {
	Key          string             `json:"key"`
	Name         string             `json:"name"`
	Description  string             `json:"description"`
	ChangeStatus ChangeStatus       `json:"changeStatus"`
	Inputs       []AnalyzerInputDto `json:"inputs"`
}

type AnalyzerInputDto struct {
	Key          string       `json:"key"`
	Value        string       `json:"value"`
	Name         string       `json:"name"`
	Description  string       `json:"description"`
	Type         string       `json:"type"`
	ChangeStatus ChangeStatus `json:"changeStatus"`
}

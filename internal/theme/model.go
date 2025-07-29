package theme

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/google/uuid"
)

type Theme struct {
	Id          uuid.UUID `gorm:"column:id;primaryKey;type:uuid;default:gen_random_uuid()"`
	UserId      uuid.UUID `gorm:"column:user_id"`
	Title       string    `gorm:"column:title"`
	Description string    `gorm:"column:description"`
	Analyzers   Analyzers `gorm:"column:analyzers;type:jsonb"`
	Reporter    Reporter  `gorm:"column:reporter;type:jsonb"`
}

type Analyzer struct {
	Key    string          `json:"key"`
	Inputs []AnalyzerInput `json:"inputs"`
}

type AnalyzerInput struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (c Reporter) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *Reporter) Scan(src any) error {
	return json.Unmarshal(src.([]byte), &c)
}

type Reporter struct {
	Threshold float32 `json:"threshold"`
	Key       string  `json:"key"`
}

type Analyzers []Analyzer

func (c Analyzers) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *Analyzers) Scan(src any) error {
	return json.Unmarshal(src.([]byte), &c)
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
	Reporters   []ReporterDto `json:"reporters"`
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

type ReporterDto struct {
	Threshold    float32      `json:"threshold"`
	Key          string       `json:"key"`
	Name         string       `json:"name"`
	Description  string       `json:"description"`
	ChangeStatus ChangeStatus `json:"changeStatus"`
}

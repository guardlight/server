package ssemanager

type EventType string

const (
	TypeUpdate    EventType = "update"
	TypeHeartbeat EventType = "heartbeat"
)

type ActionType string

const (
	ActionReportDone        ActionType = "report_done"
	ActionAnalysisDone      ActionType = "analysis_done"
	ActionAnalysisRequested ActionType = "analysis_requested"
	ActionBeat              ActionType = "beat"
)

type SseEvent struct {
	Type   EventType  `json:"type"`
	Action ActionType `json:"action"`
	Data   string     `json:"data"`
}

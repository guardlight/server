package analysismanager

import "github.com/guardlight/server/pkg/analysisrequest"

type AnalysisManager struct{}

func NewAnalysisMananger() *AnalysisManager {
	return &AnalysisManager{}
}

func (am AnalysisManager) RequestAnalysis(ar *analysisrequest.AnalysisRequest) error {
	return nil
}

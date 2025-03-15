package analysismanager

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/guardlight/server/internal/essential/config"
	"github.com/guardlight/server/internal/jobmanager"
	"github.com/guardlight/server/pkg/analyzercontract"
	"github.com/guardlight/server/pkg/parsercontract"
	"github.com/nats-io/nats.go"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

type analysisStore interface {
	updateProcessedText(ai uuid.UUID, text string) error
	getAllAnalysisByAnalysisRecordId(id uuid.UUID) ([]Analysis, error)
}

type subsriber interface {
	Subscribe(subj string, cb nats.MsgHandler) (*nats.Subscription, error)
}

type jobber interface {
	jobmanager.IdCreater
	jobmanager.JobUpdater
	jobmanager.Enqueuer
}

type AnalysisManagerAllocator struct {
	as analysisStore
	ju jobber
}

func NewAnalysisManagerAllocator(s subsriber, as analysisStore, ju jobber) *AnalysisManagerAllocator {
	ama := &AnalysisManagerAllocator{
		as: as,
		ju: ju,
	}

	s.Subscribe("parser.result", ama.processParserResult)
	s.Subscribe("analyzer.result", ama.processAnalyzerResult)

	return ama
}

func (ama *AnalysisManagerAllocator) processParserResult(m *nats.Msg) {
	var pr parsercontract.ParserResponse
	err := json.Unmarshal(m.Data, &pr)
	if err != nil {
		zap.S().Errorw("Could not unmarshal parser response", "error", err)
		// TODO build a clean up function to clean up inprogress tasks that
		//      that are running longer than x minutes.
		//      Update to error status with description, "Task running to long"
	}

	err = ama.as.updateProcessedText(pr.AnalysisId, pr.Text)
	if err != nil {
		zap.S().Errorw("Could not update processed text in raw data", "error", err)
		return
	}
	err = ama.ju.UpdateJobStatus(pr.JobId, jobmanager.Finished, "", 0)
	if err != nil {
		zap.S().Errorw("Could not update job status", "error", err)
		return
	}
	zap.S().Infow("Parser result processed")

	ama.allocateAnalyzeJobs(pr.AnalysisId, pr.Text)
}

func (ama *AnalysisManagerAllocator) allocateAnalyzeJobs(ai uuid.UUID, text string) {
	al, err := ama.as.getAllAnalysisByAnalysisRecordId(ai)
	if err != nil {
		zap.S().Errorw("Could not get analysis from request", "Analysis_request_id", ai)
		return
	}

	for _, a := range al {
		ama.buildJobsForAnalyzer(ai, a, text)
	}

}

func (ama *AnalysisManagerAllocator) buildJobsForAnalyzer(ai uuid.UUID, a Analysis, text string) {
	analyzerFromConfig, ok := config.Get().GetAnalyzer(a.AnalyzerKey)
	if !ok {
		zap.S().Errorw("Could not get analyzer from config", "analyzer_key", a.AnalyzerKey)
		return
	}

	if analyzerFromConfig.Model == "text" {
		chks := lo.ChunkString(text, analyzerFromConfig.ContextWindow)
		for _, ch := range chks {
			jid := ama.ju.CreateId()
			ajd := jobmanager.AnalyzerJobData{
				Type:  analyzerFromConfig.Key,
				Image: analyzerFromConfig.Image,
				Topic: fmt.Sprintf("analyzer.%s", analyzerFromConfig.Key),
				AnalyzerData: analyzercontract.AnalyzerRequest{
					JobId:      jid,
					AnalysisId: ai,
					Content:    ch,
				},
			}
			ama.ju.EnqueueJob(jid, jobmanager.Analyze, ajd)
			zap.S().Infow("Analyzer job submitted", "job_id", jid)
		}
	} else {
		zap.S().Errorw("Model not supported", "model", analyzerFromConfig.Model)
		return
	}

}

func (ama *AnalysisManagerAllocator) processAnalyzerResult(m *nats.Msg) {
	var ar analyzercontract.AnalyzerResponse
	err := json.Unmarshal(m.Data, &ar)
	if err != nil {
		zap.S().Errorw("Could not unmarshal analyzer response", "error", err)
		// TODO build a clean up function to clean up inprogress tasks that
		//      that are running longer than x minutes.
		//      Update to error status with description, "Task running to long"
	}

	// TODO Update analysis
	// Figure out how to know when all jobs are finished for analysis

	err = ama.ju.UpdateJobStatus(ar.JobId, jobmanager.Finished, "", 0)
	if err != nil {
		zap.S().Errorw("Could not update job status", "error", err)
		return
	}
	zap.S().Infow("Analyzer result processed")

	// if allAnalysisFinished
	// make report
}

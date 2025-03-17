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
	updateAnalysisJobs(ai uuid.UUID, jbs []SingleJobProgress) error
	updateAnalysisJobProgress(aid uuid.UUID, jid uuid.UUID, status AnalysisStatus, content []string, score float32) error
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

	if pr.Status == parsercontract.ParseError {
		err = ama.ju.UpdateJobStatus(pr.JobId, jobmanager.Error, pr.Text, 0)
		if err != nil {
			zap.S().Errorw("Could not update job status", "error", err)
			return
		}
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
		jbs := ama.buildJobsForAnalyzer(a, text)
		ama.as.updateAnalysisJobs(a.Id, jbs)
	}

}

func (ama *AnalysisManagerAllocator) buildJobsForAnalyzer(a Analysis, text string) []SingleJobProgress {
	analyzerFromConfig, ok := config.Get().GetAnalyzer(a.AnalyzerKey)
	if !ok {
		zap.S().Errorw("Could not get analyzer from config", "analyzer_key", a.AnalyzerKey)
		return nil
	}
	jobs := make([]SingleJobProgress, 0)

	if analyzerFromConfig.Model == "text" {
		chks := lo.ChunkString(text, analyzerFromConfig.ContextWindow)
		for _, ch := range chks {
			jid := ama.ju.CreateId()
			jobs = append(jobs, SingleJobProgress{
				JobId:  jid,
				Status: AnalysisWaiting,
			})
			ainputs := lo.Map(a.Inputs, func(inp AnalysisInput, _ int) analyzercontract.AnalysisInput {
				return analyzercontract.AnalysisInput{
					Key:   inp.Key,
					Value: inp.Value,
				}
			})
			ajd := jobmanager.AnalyzerJobData{
				Type:  analyzerFromConfig.Key,
				Image: analyzerFromConfig.Image,
				Topic: fmt.Sprintf("analyzer.%s", analyzerFromConfig.Key),
				AnalyzerData: analyzercontract.AnalyzerRequest{
					JobId:      jid,
					AnalysisId: a.Id,
					Content:    ch,
					Inputs:     ainputs,
				},
			}
			gk := fmt.Sprintf("analyzer.%s", analyzerFromConfig.Key)
			ama.ju.EnqueueJob(jid, jobmanager.Analyze, gk, ajd)
			zap.S().Infow("Analyzer job submitted", "job_id", jid)
		}
	} else {
		zap.S().Errorw("Model not supported", "model", analyzerFromConfig.Model)
	}

	return jobs
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

	err = ama.as.updateAnalysisJobProgress(ar.AnalysisId, ar.JobId, AnalysisFinished, ar.Results, ar.Score)
	if err != nil {
		zap.S().Errorw("Could not update analysis progress", "error", err)
		return
	}

	err = ama.ju.UpdateJobStatus(ar.JobId, jobmanager.Finished, "", 0)
	if err != nil {
		zap.S().Errorw("Could not update job status", "error", err)
		return
	}
	zap.S().Infow("Analyzer result processed")

	// if contributeToPublicMADB (MADB=Media Analysis Database)

}

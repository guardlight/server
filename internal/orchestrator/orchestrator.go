package orchestrator

import (
	"encoding/json"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"github.com/guardlight/server/internal/essential/config"
	"github.com/guardlight/server/internal/jobmanager"
	"github.com/guardlight/server/internal/natsclient"
	"go.uber.org/zap"
)

type jobManager interface {
	jobmanager.JobsGetter
	jobmanager.JobUpdater
}

type natsSender interface {
	natsclient.Publisher
}

type taskCreater interface {
	NewJob(jobDefinition gocron.JobDefinition, task gocron.Task, options ...gocron.JobOption) (gocron.Job, error)
}

type Orchestrator struct {
	jm  jobManager
	ns  natsSender
	jcs jobCounts
}

func NewOrchestrator(jm jobManager, tc taskCreater, ns natsSender) (*Orchestrator, error) {
	o := &Orchestrator{
		jm:  jm,
		ns:  ns,
		jcs: jc(),
	}

	_, err := tc.NewJob(
		gocron.CronJob(config.Get().Orchestrator.ScheduleRateCron, true),
		gocron.NewTask(o.checkForJobs),
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
	)
	if err != nil {
		return nil, err
	}

	return o, nil
}

func (o *Orchestrator) checkForJobs() {
	zap.S().Infow("Checking for jobs")
	nfj, _ := o.jm.GetAllNonFinishedJobs()
	o.jcs.build(nfj)

	for _, job := range nfj {
		if job.Status == jobmanager.Queued {
			o.processJob(job)
		}
	}
}

func (o *Orchestrator) processJob(j jobmanager.Job) {
	zap.S().Infow("Found job", "job_id", j.Id)
	if j.Type == jobmanager.Parse {
		o.processParseJob(j)
	} else if j.Type == jobmanager.Analyze {
		o.processAnalyzeJob(j)
	} else {
		zap.S().Errorw("Job type unsupported", "type", j.Type)
	}
}

func (o *Orchestrator) processParseJob(j jobmanager.Job) {
	var f jobmanager.ParserJobData
	err := json.Unmarshal(j.Data, &f)
	if err != nil {
		zap.S().Errorw("Error unmarshaling parser job data", "error", err)
		o.updateJobStatus(j.Id, jobmanager.Queued, err.Error(), j.RetryCount+1)
		return
	}

	p, ok := config.Get().GetParser(f.Type)
	if !ok {
		zap.S().Errorw("Parser type not found", "error", err)
		o.updateJobStatus(j.Id, jobmanager.Error, "Parser type not found", 3)
		return
	}

	if o.jcs[string(j.GroupKey)]+1 <= p.Concurrency {
		if f.Image == "builtin" {
			zap.S().Infow("using builtin parser")
		} else {
			// Start docker container and wait, with max duration.
		}
		o.ns.Publish(f.Topic, f.ParserData)
		zap.S().Infow("Sent data for parsing")
		o.jcs.inc(string(j.Type))
		o.updateJobStatus(j.Id, jobmanager.Inprogress, "", j.RetryCount)
	} else {
		zap.S().Infow("Job still queued")
	}
}

func (o *Orchestrator) processAnalyzeJob(j jobmanager.Job) {
	var f jobmanager.AnalyzerJobData
	err := json.Unmarshal(j.Data, &f)
	if err != nil {
		zap.S().Errorw("Error unmarshaling analyzer job data", "error", err)
		o.updateJobStatus(j.Id, jobmanager.Queued, err.Error(), j.RetryCount+1)
		return
	}

	a, ok := config.Get().GetAnalyzer(f.Type)
	if !ok {
		zap.S().Errorw("Analyzer not found", "error", err)
		o.updateJobStatus(j.Id, jobmanager.Error, "Analyzer not found", 3)
		return
	}

	if o.jcs[string(j.GroupKey)]+1 <= a.Concurrency {
		if f.Image == "builtin" {
			zap.S().Infow("using builtin analyzer")
		} else {
			// Start docker container and wait, with max duration.
		}
		o.ns.Publish(f.Topic, f.AnalyzerData)
		zap.S().Infow("Sent data for analyzing")
		o.jcs.inc(string(j.Type))
		o.updateJobStatus(j.Id, jobmanager.Inprogress, "", j.RetryCount)
	} else {
		zap.S().Infow("Job still queued")
	}
}

func (o *Orchestrator) updateJobStatus(id uuid.UUID, js jobmanager.JobStatus, jsd string, rc int) {
	zap.S().Infow("Updating job status", "status", js, "description", jsd)
	err := o.jm.UpdateJobStatus(id, js, jsd, rc)
	if err != nil {
		zap.S().Errorw("Can really not update job status", "status", js, "description", jsd)
	}
}

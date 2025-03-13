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
	natsclient.Sender
}

type taskCreater interface {
	NewJob(jobDefinition gocron.JobDefinition, task gocron.Task, options ...gocron.JobOption) (gocron.Job, error)
}

type Orchestrator struct {
	jm  jobManager
	ns  natsSender
	jcs map[string]int
}

func NewOrchestrator(jm jobManager, tc taskCreater, ns natsSender) (*Orchestrator, error) {
	o := &Orchestrator{
		jm: jm,
		ns: ns,
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

func (o *Orchestrator) buildJobCounts(j []jobmanager.Job) {
	for _, job := range j {
		if job.Status == jobmanager.Inprogress {
			o.jcs[string(job.Type)]++
		}
	}
}

func (o *Orchestrator) checkForJobs() {
	zap.S().Info("Checking for jobs")
	nfj, _ := o.jm.GetAllNonFinishedJobs()
	o.buildJobCounts(nfj)

	for _, job := range nfj {
		if job.Status == jobmanager.Error || job.Status == jobmanager.Queued {
			o.processJob(job)
		}
	}
}

func (o *Orchestrator) processJob(j jobmanager.Job) {
	zap.S().Info("Found job", "job_id", j.Id)
	if j.Type == jobmanager.Parse {
		o.processParseJob(j)
	}
	o.updateJobStatus(j.Id, jobmanager.Inprogress, "", j.RetryCount)
}

func (o *Orchestrator) processParseJob(j jobmanager.Job) {
	var f jobmanager.ParserJobData
	err := json.Unmarshal(j.Data, &f)
	if err != nil {
		zap.S().Errorw("Error unmarshaling parser job data", "error", err)
		o.updateJobStatus(j.Id, jobmanager.Error, err.Error(), j.RetryCount+1)
		return
	}

	p, ok := config.Get().GetParser(f.Type)
	if !ok {
		zap.S().Errorw("Parser type not found", "error", err)
		o.updateJobStatus(j.Id, jobmanager.Error, "Parser type not found", j.RetryCount+1)
		return
	}

	if o.jcs[string(j.Type)] < p.Concurrency {
		zap.S().Info("Sent data for parsing", "data", f)
		o.ns.Send(f.Topic)
	} else {
		zap.S().Info("Job still queued", "data", f)
	}

}

func (o *Orchestrator) updateJobStatus(id uuid.UUID, js jobmanager.JobStatus, jsd string, rc int) {
	zap.S().Infow("Updating job status", "status", js, "description", jsd)
	err := o.jm.UpdateJobStatus(id, js, jsd, rc)
	if err != nil {
		zap.S().Errorw("Can really not update job status", "status", js, "description", jsd)
	}
}

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
	zap.S().Debugw("Checking For Jobs")
	nfj, err := o.jm.GetAllNonFinishedJobs()
	if err != nil {
		zap.S().Errorw("Could not check for all open jobs", "err", err)
	}
	o.jcs.build(nfj)

	for _, job := range nfj {
		if job.Status == jobmanager.Queued {
			o.processJob(job)
		}
	}
	zap.S().Debugw("Relevant Jobs Processed")
}

func (o *Orchestrator) processJob(j jobmanager.Job) {
	zap.S().Infow("Processing Job", "job_id", j.Id)
	if j.Type == jobmanager.Parse {
		o.processParseJob(j)
	} else if j.Type == jobmanager.Analyze {
		o.processAnalyzeJob(j)
	} else if j.Type == jobmanager.Report {
		o.processReportJob(j)
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
		o.updateJobStatus(j.Id, jobmanager.Queued, "Parser type not found", 3)
		return
	}

	if o.jcs.t(j.GroupKey) <= p.Concurrency {
		if p.External {
			zap.S().Infow("Using external parser container", "image", f.Image, "type", p.Type)
		} else {
			o.updateJobStatus(j.Id, jobmanager.Error, "External=false not supported yet", 0)
			// REMOVE WHEN docker support is implemented.

			// Start docker container and wait, with max duration.
		}
		o.jcs.inc(j.GroupKey)
		err = o.ns.Publish(f.Topic, f.ParserData)
		if err != nil {
			o.jcs.dec(j.GroupKey)
			o.updateJobStatus(j.Id, jobmanager.Queued, "", j.RetryCount+1)
			return
		}
		err = o.updateJobStatus(j.Id, jobmanager.Inprogress, "", j.RetryCount)
		if err != nil {
			o.jcs.dec(j.GroupKey)
			return
		}
	} else {
		zap.S().Infow("Job Still Queued", "job_type", j.Id, "group_key", j.GroupKey)
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
		o.updateJobStatus(j.Id, jobmanager.Queued, "Analyzer not found", 3)
		return
	}

	if o.jcs.t(j.GroupKey) <= a.Concurrency {
		if a.External {
			zap.S().Infow("Using external analyzer container", "image", f.Image, "key", a.Key)
		} else {
			o.updateJobStatus(j.Id, jobmanager.Error, "External=false not supported yet", 0)
			// REMOVE WHEN docker support is implemented.

			// Start docker container and wait, with max duration.
		}
		o.jcs.inc(j.GroupKey)
		err = o.ns.Publish(f.Topic, f.AnalyzerData)
		if err != nil {
			o.jcs.dec(j.GroupKey)
			o.updateJobStatus(j.Id, jobmanager.Queued, "", j.RetryCount+1)
			return
		}
		err = o.updateJobStatus(j.Id, jobmanager.Inprogress, "", j.RetryCount)
		if err != nil {
			o.jcs.dec(j.GroupKey)
			return
		}

	} else {
		zap.S().Infow("Job Still Queued", "job_type", j.Id, "group_key", j.GroupKey)
	}
}

func (o *Orchestrator) processReportJob(j jobmanager.Job) {
	var f jobmanager.ReportJobData
	err := json.Unmarshal(j.Data, &f)
	if err != nil {
		zap.S().Errorw("Error unmarshaling reporter job data", "error", err)
		o.updateJobStatus(j.Id, jobmanager.Queued, err.Error(), j.RetryCount+1)
		return
	}

	r, ok := config.Get().GetReporter(f.Type)
	if !ok {
		zap.S().Errorw("Reporter not found", "error", err)
		o.updateJobStatus(j.Id, jobmanager.Queued, "Reporter not found", 3)
		return
	}

	if o.jcs.t(j.GroupKey) <= r.Concurrency {
		if r.External {
			zap.S().Infow("Using external reporter container", "image", r.Image, "type", j.Type)
		} else {
			o.updateJobStatus(j.Id, jobmanager.Error, "External=false not supported yet", 0)
			// REMOVE WHEN docker support is implemented.

			// Start docker container and wait, with max duration.
		}
		o.jcs.inc(j.GroupKey)
		err = o.ns.Publish(f.Topic, f.ReporterData)
		if err != nil {
			o.jcs.dec(j.GroupKey)
			o.updateJobStatus(j.Id, jobmanager.Queued, "", j.RetryCount+1)
			return
		}
		err = o.updateJobStatus(j.Id, jobmanager.Inprogress, "", j.RetryCount)
		if err != nil {
			o.jcs.dec(j.GroupKey)
			return
		}
	} else {
		zap.S().Infow("Job Still Queued", "job_type", j.Id, "group_key", j.GroupKey)
	}
}

func (o *Orchestrator) updateJobStatus(id uuid.UUID, js jobmanager.JobStatus, jsd string, rc int) error {
	if rc >= 3 {
		js = jobmanager.Error
	}
	err := o.jm.UpdateJobStatus(id, js, jsd, rc)
	if err != nil {
		// if js == jobmanager.Error {
		// 	zap.S().Errorw("Can really not update job status", "status", js, "description", jsd)
		// 	return nil
		// }
		return err

	}
	return nil
}

package jobmanager

import (
	"encoding/json"
	"time"
	"unsafe"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Public Interfaces

type Enqueuer interface {
	EnqueueJob(id uuid.UUID, jType JobType, groupKey string, data interface{}) error
}

type IdCreater interface {
	CreateId() uuid.UUID
}

type JobsGetter interface {
	GetAllNonFinishedJobs() ([]Job, error)
}

type JobUpdater interface {
	UpdateJobStatus(id uuid.UUID, status JobStatus, desc string, retryCount int) error
}

// Private

type jobStore interface {
	saveJob(j *Job) error
	getNotFinishedJobs() ([]Job, error)
	updateJobStatus(id uuid.UUID, s JobStatus, sd string, rc int) error
	deleteJob(id uuid.UUID) error
}

type JobManager struct {
	js jobStore
}

type taskCreater interface {
	NewJob(jobDefinition gocron.JobDefinition, task gocron.Task, options ...gocron.JobOption) (gocron.Job, error)
}

func NewJobMananger(js jobStore, tc taskCreater) *JobManager {
	jm := &JobManager{
		js: js,
	}

	_, err := tc.NewJob(
		gocron.DurationJob(
			30*time.Second,
		),
		gocron.NewTask(jm.stopLongRunningJobs),
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
	)
	if err != nil {
		return nil
	}

	return jm
}

func (jm *JobManager) stopLongRunningJobs() {
	zap.S().Debugw("Stopping long running jobs")
	js, _ := jm.js.getNotFinishedJobs()
	for _, j := range js {
		if j.Status == Inprogress {
			if j.Type == Parse || j.Type == Analyze || j.Type == Report {
				if time.Since(j.CreatedAt) > time.Minute {
					if j.RetryCount > 2 {
						zap.S().Infow("Stopping job", "job_id", j.Id)
						jm.UpdateJobStatus(j.Id, Error, "Timed out", 3)
					} else {
						zap.S().Infow("retrying job", "job_id", j.Id)
						jm.UpdateJobStatus(j.Id, Queued, "long running task", j.RetryCount+1)
					}

				}
			}
		}

	}
}

func (jm *JobManager) CreateId() uuid.UUID {
	return uuid.New()
}

func (jm *JobManager) EnqueueJob(id uuid.UUID, jType JobType, groupKey string, data interface{}) error {
	jData, err := json.Marshal(data)
	if err != nil {
		zap.S().Errorw("Could not marshal job data", "error", err)
		return err
	}

	j := &Job{
		Id:                id,
		Status:            Queued,
		StatusDescription: "",
		RetryCount:        0,
		GroupKey:          groupKey,
		Type:              jType,
		Data:              jData,
	}

	err = jm.js.saveJob(j)
	if err != nil {
		return err
	}

	zap.S().Infow("Job Enqueued", "job_id", id, "group_key", groupKey, "job_type", jType, "data_size", unsafe.Sizeof(data))
	return nil
}

func (jm *JobManager) GetAllNonFinishedJobs() ([]Job, error) {
	j, err := jm.js.getNotFinishedJobs()
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (jm *JobManager) UpdateJobStatus(id uuid.UUID, s JobStatus, sd string, rc int) error {
	err := jm.js.updateJobStatus(id, s, sd, rc)
	if err != nil {
		return err
	}
	zap.S().Infow("Job Status Updated", "job_id", id, "status", s, "dessription", sd)

	if s == Finished {
		jm.js.deleteJob(id)
	}

	return nil
}

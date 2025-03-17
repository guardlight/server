package jobmanager

import (
	"encoding/json"

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
}

type JobManager struct {
	js jobStore
}

func NewJobMananger(js jobStore) *JobManager {
	return &JobManager{
		js: js,
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
	return jm.js.updateJobStatus(id, s, sd, rc)
}

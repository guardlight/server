package jobmanager

import (
	"errors"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type JobManagerRepository struct {
	db *gorm.DB
}

func NewJobManagerRepository(db *gorm.DB) *JobManagerRepository {
	if err := db.AutoMigrate(&Job{}); err != nil {
		zap.S().DPanicw("Problem automigrating the tables", "error", err)
	}

	return &JobManagerRepository{
		db: db,
	}
}

func (jmr JobManagerRepository) saveJob(j *Job) error {
	if err := jmr.db.Create(j).Error; err != nil {
		zap.S().Errorw("Could not save job", "error", err)
		return err
	}

	return nil
}

func (jmr JobManagerRepository) getNotFinishedJobs() ([]Job, error) {
	var js []Job
	if err := jmr.db.Where("status <> ? AND retry_count <= ?", Finished, 3).Find(&js).Error; err != nil {
		zap.S().Errorw("Could not get unfinished jobs", "error", err)
		return nil, err
	}
	return js, nil
}

func (jmr JobManagerRepository) updateJobStatus(id uuid.UUID, s JobStatus, sd string, rc int) error {
	res := jmr.db.Model(Job{Id: id}).Updates(Job{
		Status:            s,
		StatusDescription: sd,
		RetryCount:        rc,
	})

	if res.Error != nil {
		zap.S().Errorw("Could not update job", "error", res.Error)
		return res.Error
	}

	if res.RowsAffected == 0 {
		zap.S().Errorw("No records updated", "job_type", id)
		return errors.New("no records affected after update")
	}

	return nil
}

func (jmr JobManagerRepository) deleteJob(id uuid.UUID) error {
	if err := jmr.db.Delete(&Job{Id: id}).Error; err != nil {
		zap.S().Errorw("Could not delete job", "error", err, "id", id)
		return err
	}
	return nil
}

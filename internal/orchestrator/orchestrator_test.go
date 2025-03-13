package orchestrator

import (
	"testing"

	"github.com/google/uuid"
	"github.com/guardlight/server/internal/essential/config"
	"github.com/guardlight/server/internal/essential/logging"
	"github.com/guardlight/server/internal/jobmanager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAnalysisOrchestrator(t *testing.T) {
	mockJm := NewMockjobManager(t)
	mockTc := NewMocktaskCreater(t)
	mockNs := NewMocknatsSender(t)
	config.SetupConfig("../../env-test.yaml")
	logging.SetupLogging("test")

	mockTc.EXPECT().NewJob(mock.AnythingOfType("cronJobDefinition"), mock.AnythingOfType("Task"), mock.AnythingOfType("JobOption")).Return(nil, nil)

	jobs := []jobmanager.Job{
		{
			Id:                uuid.MustParse("b268c2e9-3a9d-4e36-a17f-33032fa77c72"),
			Status:            jobmanager.Queued,
			StatusDescription: "",
			RetryCount:        0,
			Type:              jobmanager.Parse,
			Data:              []byte("{\"type\":\"freetext\",\"topic\":\"parser.freetext\",\"parserData\":{\"jobId\":\"b268c2e9-3a9d-4e36-a17f-33032fa77c72\",\"analysisId\":\"dcfd5683-bccc-42b0-963a-93fc97ecf67d\",\"Content\":\"UnVubmluZyBhbmQgV2Fsa2luZw==\"}}"),
		},
	}

	mockJm.EXPECT().GetAllNonFinishedJobs().Return(jobs, nil)

	mockJm.EXPECT().UpdateJobStatus(uuid.MustParse("b268c2e9-3a9d-4e36-a17f-33032fa77c72"), jobmanager.Inprogress, "", 0).Return(nil)
	mockNs.EXPECT().Send("parser.freetext").Return(nil)

	o, err := NewOrchestrator(mockJm, mockTc, mockNs)
	assert.NoError(t, err)
	o.checkForJobs()
}

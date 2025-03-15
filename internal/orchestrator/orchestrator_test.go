package orchestrator

import (
	"testing"

	"github.com/google/uuid"
	"github.com/guardlight/server/internal/essential/config"
	"github.com/guardlight/server/internal/essential/logging"
	"github.com/guardlight/server/internal/jobmanager"
	"github.com/guardlight/server/pkg/parsercontract"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAnalysisOrchestratorStart(t *testing.T) {
	mockJm := NewMockjobManager(t)
	mockTc := NewMocktaskCreater(t)
	mockNs := NewMocknatsSender(t)
	config.SetupConfig("../../env-test.yaml")
	logging.SetupLogging("test")

	mockTc.EXPECT().NewJob(mock.AnythingOfType("cronJobDefinition"), mock.AnythingOfType("Task"), mock.AnythingOfType("JobOption")).Return(nil, nil)

	jobId := uuid.MustParse("b268c2e9-3a9d-4e36-a17f-33032fa77c72")
	jobs := []jobmanager.Job{
		{
			Id:                jobId,
			Status:            jobmanager.Queued,
			StatusDescription: "",
			RetryCount:        0,
			Type:              jobmanager.Parse,
			Data:              []byte("{\"image\":\"builtin\",\"type\":\"freetext\",\"topic\":\"parser.freetext\",\"parserData\":{\"jobId\":\"b268c2e9-3a9d-4e36-a17f-33032fa77c72\",\"analysisId\":\"165c0cff-9395-4b10-8636-9d65b3d364ef\",\"Content\":\"UnVubmluZyBhbmQgV2Fsa2luZw==\"}}"),
		},
	}

	mockJm.EXPECT().GetAllNonFinishedJobs().Return(jobs, nil)

	mockJm.EXPECT().UpdateJobStatus(uuid.MustParse("b268c2e9-3a9d-4e36-a17f-33032fa77c72"), jobmanager.Inprogress, "", 0).Return(nil)

	pr := parsercontract.ParserRequest{
		JobId:      jobId,
		AnalysisId: uuid.MustParse("165c0cff-9395-4b10-8636-9d65b3d364ef"),
		Content:    []byte("Running and Walking"),
	}
	mockNs.EXPECT().Publish("parser.freetext", pr).Return(nil)

	o, err := NewOrchestrator(mockJm, mockTc, mockNs)
	assert.NoError(t, err)
	o.checkForJobs()
}

func TestAnalysisOrchestratorDoNothing(t *testing.T) {
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
		{
			Id:                uuid.MustParse("83ee678d-ef1d-4a9d-8264-5079f5696a8f"),
			Status:            jobmanager.Inprogress,
			StatusDescription: "",
			RetryCount:        0,
			Type:              jobmanager.Parse,
			Data:              []byte("{\"type\":\"freetext\",\"topic\":\"parser.freetext\",\"parserData\":{\"jobId\":\"83ee678d-ef1d-4a9d-8264-5079f5696a8f\",\"analysisId\":\"165c0cff-9395-4b10-8636-9d65b3d364ef\",\"Content\":\"UnVubmluZyBhbmQgV2Fsa2luZw==\"}}"),
		},
	}

	mockJm.EXPECT().GetAllNonFinishedJobs().Return(jobs, nil)

	mockJm.AssertNotCalled(t, "UpdateJobStatus")
	mockNs.AssertNotCalled(t, "Publish")

	o, err := NewOrchestrator(mockJm, mockTc, mockNs)
	assert.NoError(t, err)
	o.checkForJobs()
}

func TestAnalysisOrchestratorStartNew(t *testing.T) {
	mockJm := NewMockjobManager(t)
	mockTc := NewMocktaskCreater(t)
	mockNs := NewMocknatsSender(t)
	config.SetupConfig("../../env-test.yaml")
	logging.SetupLogging("test")

	mockTc.EXPECT().NewJob(mock.AnythingOfType("cronJobDefinition"), mock.AnythingOfType("Task"), mock.AnythingOfType("JobOption")).Return(nil, nil)

	jobId := uuid.MustParse("b268c2e9-3a9d-4e36-a17f-33032fa77c72")

	jobs := []jobmanager.Job{
		{
			Id:                jobId,
			Status:            jobmanager.Queued,
			StatusDescription: "",
			RetryCount:        0,
			Type:              jobmanager.Parse,
			Data:              []byte("{\"image\":\"builtin\",\"type\":\"freetext\",\"topic\":\"parser.freetext\",\"parserData\":{\"jobId\":\"b268c2e9-3a9d-4e36-a17f-33032fa77c72\",\"analysisId\":\"dcfd5683-bccc-42b0-963a-93fc97ecf67d\",\"Content\":\"UnVubmluZyBhbmQgV2Fsa2luZw==\"}}"),
		},
		{
			Id:                uuid.MustParse("83ee678d-ef1d-4a9d-8264-5079f5696a8f"),
			Status:            jobmanager.Finished,
			StatusDescription: "",
			RetryCount:        0,
			Type:              jobmanager.Parse,
			Data:              []byte("{\"image\":\"builtin\",\"type\":\"freetext\",\"topic\":\"parser.freetext\",\"parserData\":{\"jobId\":\"83ee678d-ef1d-4a9d-8264-5079f5696a8f\",\"analysisId\":\"165c0cff-9395-4b10-8636-9d65b3d364ef\",\"Content\":\"UnVubmluZyBhbmQgV2Fsa2luZw==\"}}"),
		},
	}

	mockJm.EXPECT().GetAllNonFinishedJobs().Return(jobs, nil)

	mockJm.EXPECT().UpdateJobStatus(jobId, jobmanager.Inprogress, "", 0).Return(nil)

	pr := parsercontract.ParserRequest{
		JobId:      jobId,
		AnalysisId: uuid.MustParse("dcfd5683-bccc-42b0-963a-93fc97ecf67d"),
		Content:    []byte("Running and Walking"),
	}
	mockNs.EXPECT().Publish("parser.freetext", pr).Return(nil)

	o, err := NewOrchestrator(mockJm, mockTc, mockNs)
	assert.NoError(t, err)
	o.checkForJobs()
}

func TestAnalysisOrchestratorFailed(t *testing.T) {
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
			Data:              []byte("{\"type\":\"unknown\",\"topic\":\"parser.freetext\",\"parserData\":{\"jobId\":\"b268c2e9-3a9d-4e36-a17f-33032fa77c72\",\"analysisId\":\"dcfd5683-bccc-42b0-963a-93fc97ecf67d\",\"Content\":\"UnVubmluZyBhbmQgV2Fsa2luZw==\"}}"),
		},
	}

	mockJm.EXPECT().GetAllNonFinishedJobs().Return(jobs, nil)

	mockJm.EXPECT().UpdateJobStatus(uuid.MustParse("b268c2e9-3a9d-4e36-a17f-33032fa77c72"), jobmanager.Error, "Parser type not found", 3).Return(nil)
	mockNs.AssertNotCalled(t, "Publish")

	o, err := NewOrchestrator(mockJm, mockTc, mockNs)
	assert.NoError(t, err)
	o.checkForJobs()
}

func TestAnalysisOrchestratorCannotUnmarshal(t *testing.T) {
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
			Data:              []byte("Wont unmarshal"),
		},
	}

	mockJm.EXPECT().GetAllNonFinishedJobs().Return(jobs, nil)

	mockJm.EXPECT().UpdateJobStatus(uuid.MustParse("b268c2e9-3a9d-4e36-a17f-33032fa77c72"), jobmanager.Queued, "invalid character 'W' looking for beginning of value", 1).Return(nil)
	mockNs.AssertNotCalled(t, "Publish")

	o, err := NewOrchestrator(mockJm, mockTc, mockNs)
	assert.NoError(t, err)
	o.checkForJobs()
}

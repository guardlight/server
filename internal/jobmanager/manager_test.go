package jobmanager

import (
	"testing"

	"github.com/google/uuid"
	"github.com/guardlight/server/internal/essential/config"
	"github.com/guardlight/server/pkg/parsercontract"

	"github.com/stretchr/testify/assert"
)

func TestAnalysisRequestParsersAndAnalyzersSuccess(t *testing.T) {
	mockJs := NewMockjobStore(t)
	config.SetupConfig("../../env-test.yaml")

	jobId := uuid.MustParse("d2efcbb9-c7e0-423c-95c3-a01e7723bedf")
	analysisId := uuid.MustParse("4bc608a3-7f52-4dd4-97dc-ea01975d9f09")

	jm := NewJobMananger(mockJs)

	pjd := &ParserJobData{
		Type:  "test",
		Topic: "test",
		ParserData: parsercontract.ParserRequest{
			JobId:      jobId,
			AnalysisId: analysisId,
			Content:    []byte("Content goes here as byte array"),
		},
	}

	j := &Job{
		Id:                jobId,
		Status:            Queued,
		StatusDescription: "",
		RetryCount:        0,
		Type:              "test",
		Data:              []byte("{\"Name\":\"Nice to know\",\"Content\":\"Q29udGVudCBnb2VzIGhlcmUgYXMgYnl0ZSBhcnJheQ==\"}"),
	}

	mockJs.EXPECT().saveJob(j).Return(nil)

	err := jm.EnqueueJob(jobId, "test", pjd)

	assert.NoError(t, err)
}

package jobmanager

import (
	"encoding/base64"
	"testing"

	"github.com/google/uuid"
	"github.com/guardlight/server/internal/essential/config"
	"github.com/guardlight/server/pkg/parsercontract"

	"github.com/stretchr/testify/assert"
)

func TestAnalysisRequestParsersAndAnalyzersSuccess(t *testing.T) {
	mockJs := NewMockjobStore(t)
	mockTc := NewMocktaskCreater(t)
	config.SetupConfig("../../env-test.yaml")

	jobId := uuid.MustParse("d2efcbb9-c7e0-423c-95c3-a01e7723bedf")
	analysisId := uuid.MustParse("4bc608a3-7f52-4dd4-97dc-ea01975d9f09")

	jm := NewJobMananger(mockJs, mockTc)

	pjd := ParserJobData{
		Image: "test",
		Type:  "test",
		Topic: "test",
		ParserData: parsercontract.ParserRequest{
			JobId:      jobId,
			AnalysisId: analysisId,
			Content:    base64.StdEncoding.EncodeToString([]byte("Content goes here as byte array")),
		},
	}

	j := &Job{
		Id:                jobId,
		Status:            Queued,
		StatusDescription: "",
		RetryCount:        0,
		Type:              "test",
		GroupKey:          "test",
		Data:              []byte("{\"type\":\"test\",\"topic\":\"test\",\"image\":\"test\",\"parserData\":{\"jobId\":\"d2efcbb9-c7e0-423c-95c3-a01e7723bedf\",\"analysisId\":\"4bc608a3-7f52-4dd4-97dc-ea01975d9f09\",\"content\":\"Q29udGVudCBnb2VzIGhlcmUgYXMgYnl0ZSBhcnJheQ==\"}}"),
	}

	mockJs.EXPECT().saveJob(j).Return(nil)

	err := jm.EnqueueJob(jobId, "test", "test", pjd)

	assert.NoError(t, err)
}

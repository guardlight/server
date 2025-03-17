package analysismanager

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/guardlight/server/internal/essential/config"
	"github.com/guardlight/server/internal/jobmanager"
	"github.com/guardlight/server/pkg/analyzercontract"
	"github.com/guardlight/server/pkg/parsercontract"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAnalysisAllocatorParserResult(t *testing.T) {
	config.SetupConfig("../../testdata/envs/analysismanangerallocator.yaml")

	mockS := NewMocksubsriber(t)
	mockAs := NewMockanalysisStore(t)
	mockJu := NewMockjobber(t)

	mockS.EXPECT().Subscribe("parser.result", mock.AnythingOfType("nats.MsgHandler")).Return(nil, nil)
	mockS.EXPECT().Subscribe("analyzer.result", mock.AnythingOfType("nats.MsgHandler")).Return(nil, nil)

	ama := NewAnalysisManagerAllocator(mockS, mockAs, mockJu)

	ai := uuid.MustParse("674e46b6-a4f5-4b4f-bc16-c29ba80971c0")
	jobId := uuid.MustParse("e007bc38-0373-4da6-895e-c76e9ee331e7")

	t.Run("parser_result_success", func(t *testing.T) {
		pr := parsercontract.ParserResponse{
			JobId:      jobId,
			AnalysisId: ai,
			Text:       "This is a book",
			Status:     parsercontract.ParseSuccess,
		}

		dat, err := json.Marshal(pr)
		assert.NoError(t, err)

		n := &nats.Msg{
			Data: dat,
		}

		mockAs.EXPECT().updateProcessedText(ai, "This is a book").Return(nil)
		mockJu.EXPECT().UpdateJobStatus(jobId, jobmanager.Finished, "", 0).Return(nil)

		ans := []Analysis{
			{
				Id:                uuid.MustParse("8e1305f1-3fae-44e5-8a4f-9f815321ae8c"),
				AnalysisRequestId: ai,
				AnalyzerKey:       "word_search",
				ThemeId:           uuid.MustParse("bd2d2784-0fcb-4526-ad3c-76d00964f4de"),
				Status:            AnalysisWaiting,
				Threshold:         2,
				Score:             0,
				Inputs: []AnalysisInput{
					{
						Key:   "strict_words",
						Value: "Running, Walking",
					},
				},
				Content: []string{},
			},
		}

		mockAs.EXPECT().getAllAnalysisByAnalysisRecordId(ai).Return(ans, nil)

		jid := uuid.MustParse("45826a77-8377-4cce-9388-6f8f2154f998")
		ar := jobmanager.AnalyzerJobData{
			Type:  "word_search",
			Image: "builtin",
			Topic: "analyzer.word_search",
			AnalyzerData: analyzercontract.AnalyzerRequest{
				JobId:      jid,
				AnalysisId: ai,
				Content:    "This is a book",
				Inputs: []analyzercontract.AnalysisInput{
					{
						Key:   "strict_words",
						Value: "Running, Walking",
					},
				},
			},
		}

		mockAs.EXPECT().updateAnalysisJobs(ai, []SingleJobProgress{
			{
				JobId:  jid,
				Status: AnalysisWaiting,
			},
		}).Return(nil)
		mockJu.EXPECT().CreateId().Return(jid)
		mockJu.EXPECT().EnqueueJob(jid, jobmanager.Analyze, "analyzer.word_search", ar).Return(nil)

		ama.processParserResult(n)

	})

	t.Run("parser_result_failed", func(t *testing.T) {
		pr := parsercontract.ParserResponse{
			JobId:      jobId,
			AnalysisId: ai,
			Text:       "Error parsing",
			Status:     parsercontract.ParseError,
		}

		dat, err := json.Marshal(pr)
		assert.NoError(t, err)

		n := &nats.Msg{
			Data: dat,
		}

		mockAs.EXPECT().updateProcessedText(ai, "Error parsing").Return(nil)
		mockJu.EXPECT().UpdateJobStatus(jobId, jobmanager.Error, "Error parsing", 0).Return(nil)

		mockAs.AssertNotCalled(t, "getAllAnalysisByAnalysisRecordId")
		mockAs.AssertNotCalled(t, "updateAnalysisJobs")
		mockJu.AssertNotCalled(t, "CreateId")
		mockJu.AssertNotCalled(t, "EnqueueJob")
		ama.processParserResult(n)

	})

}

// TODO Add processAnalyzerResult

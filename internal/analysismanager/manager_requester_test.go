package analysismanager

import (
	"testing"

	"github.com/google/uuid"
	"github.com/guardlight/server/internal/essential/config"
	"github.com/guardlight/server/internal/jobmanager"
	"github.com/guardlight/server/pkg/analysisrequest"
	"github.com/guardlight/server/pkg/parsercontract"
	"github.com/stretchr/testify/assert"
)

var userId = uuid.MustParse("f6bec23c-5106-4805-980f-9c9c1c050af4")

func TestAnalysisRequestParsersAndAnalyzersSuccess(t *testing.T) {
	mockAnalysisRecordSaver := NewMockanalysisRecordSaver(t)
	mockJobManager := NewMockjobManagerRequester(t)
	config.SetupConfig("../../env-test.yaml")

	analyzerRequester := NewAnalysisManangerRequester(mockJobManager, mockAnalysisRecordSaver)

	t.Run("parserFailed", func(t *testing.T) {
		ar := &analysisrequest.AnalysisRequest{
			Title:       "test analysis",
			ContentType: analysisrequest.MOVIE,
			File: analysisrequest.File{
				Content:  []byte("Running and walking"),
				Mimetype: "Unknown File Type",
			},
			Themes: []analysisrequest.Theme{},
		}

		err := analyzerRequester.RequestAnalysis(ar, userId)

		assert.ErrorIs(t, err, ErrInvalidParser)
	})

	t.Run("analyzerFailed", func(t *testing.T) {
		ar := &analysisrequest.AnalysisRequest{
			Title:       "test analysis",
			ContentType: analysisrequest.MOVIE,
			File: analysisrequest.File{
				Content:  []byte("Running and walking"),
				Mimetype: "freetext",
			},
			Themes: []analysisrequest.Theme{
				{
					Title: "Test Theme",
					Id:    uuid.MustParse("2864d1b0-411a-4c6c-932a-61acddd67019"),
					Analyzers: []analysisrequest.Analyzer{
						{
							Key:       "unknown_analyzer",
							Threshold: 2,
							Inputs:    []analysisrequest.AnalyzerInput{},
						},
					},
				},
			},
		}

		err := analyzerRequester.RequestAnalysis(ar, userId)

		assert.ErrorIs(t, err, ErrInvalidAnalyzer)
	})

	t.Run("analyzerValidNoInput", func(t *testing.T) {
		ar := &analysisrequest.AnalysisRequest{
			Title:       "test analysis",
			ContentType: analysisrequest.MOVIE,
			File: analysisrequest.File{
				Content:  []byte("Running and walking"),
				Mimetype: "freetext",
			},
			Themes: []analysisrequest.Theme{
				{
					Title: "Test Theme",
					Id:    uuid.MustParse("2864d1b0-411a-4c6c-932a-61acddd67019"),
					Analyzers: []analysisrequest.Analyzer{
						{
							Key:       "word_search",
							Threshold: 2,
							Inputs:    []analysisrequest.AnalyzerInput{},
						},
					},
				},
			},
		}

		err := analyzerRequester.RequestAnalysis(ar, userId)

		assert.ErrorIs(t, err, ErrInvalidAnalyzer)
	})

}

func TestAnalysisRequestSuccess(t *testing.T) {
	mockAnalysisRecordSaver := NewMockanalysisRecordSaver(t)
	mockJobManager := NewMockjobManagerRequester(t)
	config.SetupConfig("../../env-test.yaml")

	analyzerRequester := NewAnalysisManangerRequester(mockJobManager, mockAnalysisRecordSaver)

	jobId := uuid.MustParse("0e4240a2-a099-4501-b373-7d982b5d5d5d")
	mockJobManager.EXPECT().CreateId().Return(jobId)

	ar := &analysisrequest.AnalysisRequest{
		Title:       "test analysis",
		ContentType: analysisrequest.MOVIE,
		File: analysisrequest.File{
			Content:  []byte("Running and walking"),
			Mimetype: "freetext",
		},
		Themes: []analysisrequest.Theme{
			{
				Title: "Test Theme",
				Id:    uuid.MustParse("2864d1b0-411a-4c6c-932a-61acddd67019"),
				Analyzers: []analysisrequest.Analyzer{
					{
						Key:       "word_search",
						Threshold: 2,
						Inputs: []analysisrequest.AnalyzerInput{
							{
								Key:   "strict_words",
								Value: "Running, Walking",
							},
						},
					},
				},
			},
			{
				Title: "Test Swim Theme",
				Id:    uuid.MustParse("3ab4a569-4de4-4206-a4fe-b4d2ddac3f6c"),
				Analyzers: []analysisrequest.Analyzer{
					{
						Key:       "word_search",
						Threshold: 2,
						Inputs: []analysisrequest.AnalyzerInput{
							{
								Key:   "strict_words",
								Value: "Swimming, Drowning",
							},
						},
					},
				},
			},
		},
	}

	steps := []AnalysisRequestStep{
		{
			Id:                uuid.Nil,
			AnalysisRequestId: uuid.Nil,
			Index:             0,
			StepType:          Create,
			Status:            Finished,
			StatusDescription: "",
		},
		{
			Id:                uuid.Nil,
			AnalysisRequestId: uuid.Nil,
			Index:             1,
			StepType:          Parse,
			Status:            Waiting,
			StatusDescription: "",
		},
		{
			Id:                uuid.Nil,
			AnalysisRequestId: uuid.Nil,
			Index:             2,
			StepType:          Analyze,
			Status:            Waiting,
			StatusDescription: "",
		},
		{
			Id:                uuid.Nil,
			AnalysisRequestId: uuid.Nil,
			Index:             3,
			StepType:          Analyze,
			Status:            Waiting,
			StatusDescription: "",
		},
		{
			Id:                uuid.Nil,
			AnalysisRequestId: uuid.Nil,
			Index:             4,
			StepType:          Report,
			Status:            Waiting,
			StatusDescription: "",
		},
	}

	rawData := RawData{
		Content:  []byte("Running and walking"),
		Hash:     "8640144a9e60ba45c3cea6f444987b41",
		FileType: "freetext",
	}

	as := []Analysis{
		{
			Id:                uuid.Nil,
			AnalysisRequestId: uuid.Nil,
			AnalyzerKey:       "word_search",
			ThemeId:           uuid.MustParse("2864d1b0-411a-4c6c-932a-61acddd67019"),
			Status:            AnalysisWaiting,
			Threshold:         2,
			Score:             0,
			Content:           Content{},
			Inputs: []AnalysisInput{
				{
					Key:   "strict_words",
					Value: "Running, Walking",
				},
			},
			Jobs: []SingleJobProgress{},
		},
		{
			Id:                uuid.Nil,
			AnalysisRequestId: uuid.Nil,
			AnalyzerKey:       "word_search",
			ThemeId:           uuid.MustParse("3ab4a569-4de4-4206-a4fe-b4d2ddac3f6c"),
			Status:            AnalysisWaiting,
			Threshold:         2,
			Score:             0,
			Content:           Content{},
			Inputs: []AnalysisInput{
				{
					Key:   "strict_words",
					Value: "Swimming, Drowning",
				},
			},
			Jobs: []SingleJobProgress{},
		},
	}

	arDb := &AnalysisRequest{
		Title:                "test analysis",
		UserId:               userId,
		AnalysisRequestSteps: steps,
		RawData:              rawData,
		Analysis:             as,
	}

	analysisId := uuid.MustParse("75d25964-6d59-4f88-97f8-dfd3afe96c62")
	mockAnalysisRecordSaver.EXPECT().createAnalysisRequest(arDb).RunAndReturn(func(ar *AnalysisRequest) error {
		ar.Id = analysisId
		return nil
	})

	pData := jobmanager.ParserJobData{
		Image: "builtin",
		Type:  "freetext",
		Topic: "parser.freetext",
		ParserData: parsercontract.ParserRequest{
			JobId:      jobId,
			AnalysisId: analysisId,
			Content:    []byte("Running and walking"),
		},
	}

	mockJobManager.EXPECT().EnqueueJob(jobId, jobmanager.Parse, "parser.freetext", pData).Return(nil)

	err := analyzerRequester.RequestAnalysis(ar, userId)

	assert.NoError(t, err)
}

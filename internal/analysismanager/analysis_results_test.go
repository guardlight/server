package analysismanager

import (
	"testing"

	"github.com/google/uuid"
	"github.com/guardlight/server/internal/essential/config"
	"github.com/guardlight/server/internal/theme"
	"github.com/guardlight/server/pkg/analysisresult"
	"github.com/stretchr/testify/assert"
)

func TestAnalysisGetAllAnalysis(t *testing.T) {
	mars := NewMockanalysisGetter(t)
	marsu := NewMockanalysisUpdater(t)
	mts := NewMockthemeService(t)
	config.SetupConfig("../../testdata/envs/analysisresults.yaml")

	userId := uuid.MustParse("f6bec23c-5106-4805-980f-9c9c1c050af4")
	analyzerResults := NewAnalysisResultService(mars, marsu, mts)

	t.Run("success", func(t *testing.T) {

		as := []AnalysisRequest{
			{
				Id:          uuid.MustParse("78aa39f6-af28-4f01-8809-2e30e7f225d0"),
				UserId:      userId,
				Title:       "Test Analysis Request",
				ContentType: "book",
				RawData:     RawData{},
				Analysis: []Analysis{
					{
						Id:                uuid.MustParse("99d28902-0c4c-40c5-acf0-5d74af24b85b"),
						AnalysisRequestId: uuid.MustParse("ea5d3f89-e5af-49bd-a792-cacf78f7bebe"),
						AnalyzerKey:       "word_search",
						ThemeId:           uuid.MustParse("09a8c66d-d0df-435f-87e2-4f5f17c8c0f1"),
						Status:            AnalysisFinished,
						Score:             1,
						Content: []string{
							"this is a test content",
						},
						Inputs: []AnalysisInput{
							{
								Key:   "threshold",
								Value: "0.24",
							},
						},
						Jobs: []SingleJobProgress{
							{
								JobId:  uuid.MustParse("2d2fae29-dc9a-4d8f-8ae0-f0cd4c35f01b"),
								Status: AnalysisFinished,
							},
						},
					},
				},
			},
		}

		ts := []theme.ThemeDto{
			{
				Id:          uuid.MustParse("09a8c66d-d0df-435f-87e2-4f5f17c8c0f1"),
				Title:       "Test Theme",
				Description: "",
				Analyzers:   []theme.AnalyzerDto{},
			},
		}

		mars.EXPECT().getAnalysesByUserId(userId, Pagination{Limit: 10, Page: 1}).Return(AnalysisResultPaginated{Limit: 10, Page: 1, TotalPages: 1, Requests: as}, nil)
		mts.EXPECT().GetAllThemesByUserId(userId).Return(ts, nil)

		res, err := analyzerResults.GetAnalysesByUserId(userId, 10, 1)
		assert.NoError(t, err)

		asRes := []analysisresult.Analysis{
			{
				Id:          uuid.MustParse("78aa39f6-af28-4f01-8809-2e30e7f225d0"),
				Title:       "Test Analysis Request",
				ContentType: "book",
				Themes: []analysisresult.Theme{
					{
						Id:    uuid.MustParse("09a8c66d-d0df-435f-87e2-4f5f17c8c0f1"),
						Title: "Test Theme",
						Analyzers: []analysisresult.Analyzer{
							{
								Key:    "word_search",
								Name:   "Word Search Analyzer",
								Status: "finished",
								Score:  1,
								Content: []string{
									"this is a test content",
								},
								Inputs: []analysisresult.AnalyzerInput{
									{
										Key:   "threshold",
										Name:  "Threshold",
										Value: "0.24",
									},
								},
								Jobs: []analysisresult.AnalyzerJobProgress{
									{
										Status: "finished",
									},
								},
							},
						},
					},
				},
			},
		}

		assert.Equal(t, analysisresult.AnalysisPaginated{Limit: 10, Page: 1, TotalPages: 1, Analyses: asRes}, res)

	})

}

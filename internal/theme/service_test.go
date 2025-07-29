package theme

import (
	"testing"

	"github.com/google/uuid"
	"github.com/guardlight/server/internal/essential/config"
	"github.com/stretchr/testify/assert"
)

func TestGetTheme(t *testing.T) {
	config.SetupConfig("../../testdata/envs/theme.yaml")

	uid := uuid.MustParse("c5fdf7a6-ec31-49d5-af69-6a77bb6e43fe")

	defaultEmptyThemeDto := ThemeDto{
		Id:    uuid.Nil,
		Title: "",
		Analyzers: []AnalyzerDto{
			{
				Key:          "word_search",
				Name:         "Word Search Analyzer",
				Description:  "Uses a basic word list to scan content for.",
				ChangeStatus: New,
				Inputs: []AnalyzerInputDto{
					{
						Key:          "strict_words",
						Value:        "",
						Name:         "Strict Words",
						Description:  "Words in this list will immediatly flag the content.",
						Type:         "textarea",
						ChangeStatus: New,
					},
				},
			},
		},
	}

	var tests = []struct {
		name   string
		input  []Theme
		expect []ThemeDto
	}{
		{
			name: "analyzer_same",
			input: []Theme{
				{
					Id:     uuid.MustParse("d2ccc491-c8c3-48e8-867a-fbc82b7018b2"),
					UserId: uid,
					Title:  "Test Theme",
					Analyzers: []Analyzer{
						{
							Key: "word_search",
							Inputs: []AnalyzerInput{
								{
									Key:   "strict_words",
									Value: "Running, Walking",
								},
							},
						},
					},
					Reporter: Reporter{
						Key:       "word_count",
						Threshold: 1,
					},
				},
			},
			expect: []ThemeDto{
				{
					Id:    uuid.MustParse("d2ccc491-c8c3-48e8-867a-fbc82b7018b2"),
					Title: "Test Theme",
					Analyzers: []AnalyzerDto{
						{
							Key:          "word_search",
							Name:         "Word Search Analyzer",
							Description:  "Uses a basic word list to scan content for.",
							ChangeStatus: Same,
							Inputs: []AnalyzerInputDto{
								{
									Key:          "strict_words",
									Value:        "Running, Walking",
									Name:         "Strict Words",
									Description:  "Words in this list will immediatly flag the content.",
									Type:         "textarea",
									ChangeStatus: Same,
								},
							},
						},
					},
					Reporters: []ReporterDto{
						{
							Key:          "word_count",
							Threshold:    1,
							Name:         "Word Count",
							Description:  "Basic reporter that will check basic counts.",
							ChangeStatus: Same,
						},
					},
				},
				defaultEmptyThemeDto,
			},
		},
		{
			name: "analyzer_added",
			input: []Theme{
				{
					Id:        uuid.MustParse("d2ccc491-c8c3-48e8-867a-fbc82b7018b2"),
					UserId:    uid,
					Title:     "Test Theme",
					Analyzers: []Analyzer{},
					Reporter: Reporter{
						Key:       "word_count",
						Threshold: 1,
					},
				},
			},
			expect: []ThemeDto{
				{
					Id:    uuid.MustParse("d2ccc491-c8c3-48e8-867a-fbc82b7018b2"),
					Title: "Test Theme",
					Analyzers: []AnalyzerDto{
						{
							Key:          "word_search",
							Name:         "Word Search Analyzer",
							Description:  "Uses a basic word list to scan content for.",
							ChangeStatus: New,
							Inputs: []AnalyzerInputDto{
								{
									Key:          "strict_words",
									Value:        "",
									Name:         "Strict Words",
									Description:  "Words in this list will immediatly flag the content.",
									Type:         "textarea",
									ChangeStatus: New,
								},
							},
						},
					},
					Reporters: []ReporterDto{
						{
							Key:          "word_count",
							Threshold:    1,
							Name:         "Word Count",
							Description:  "Basic reporter that will check basic counts.",
							ChangeStatus: Same,
						},
					},
				},
				defaultEmptyThemeDto,
			},
		},
		{
			name: "analyzer_removed",
			input: []Theme{
				{
					Id:     uuid.MustParse("d2ccc491-c8c3-48e8-867a-fbc82b7018b2"),
					UserId: uid,
					Title:  "Test Theme",
					Analyzers: []Analyzer{
						{
							Key: "super_ai",
							Inputs: []AnalyzerInput{
								{
									Key:   "command_prompt",
									Value: "Is there walking?",
								},
							},
						},
					},
					Reporter: Reporter{
						Key:       "word_count",
						Threshold: 1,
					},
				},
			},
			expect: []ThemeDto{
				{
					Id:    uuid.MustParse("d2ccc491-c8c3-48e8-867a-fbc82b7018b2"),
					Title: "Test Theme",
					Analyzers: []AnalyzerDto{
						{
							Key:          "super_ai",
							Name:         "",
							Description:  "",
							ChangeStatus: Removed,
							Inputs: []AnalyzerInputDto{
								{
									Key:          "command_prompt",
									Value:        "Is there walking?",
									Name:         "",
									Description:  "",
									Type:         "",
									ChangeStatus: Removed,
								},
							},
						},
						{
							Key:          "word_search",
							Name:         "Word Search Analyzer",
							Description:  "Uses a basic word list to scan content for.",
							ChangeStatus: New,
							Inputs: []AnalyzerInputDto{
								{
									Key:          "strict_words",
									Value:        "",
									Name:         "Strict Words",
									Description:  "Words in this list will immediatly flag the content.",
									Type:         "textarea",
									ChangeStatus: New,
								},
							},
						},
					},
					Reporters: []ReporterDto{
						{
							Key:          "word_count",
							Threshold:    1,
							Name:         "Word Count",
							Description:  "Basic reporter that will check basic counts.",
							ChangeStatus: Same,
						},
					},
				},
				defaultEmptyThemeDto,
			},
		},
		{
			name: "analyzer_changed_input_new_and_removed",
			input: []Theme{
				{
					Id:     uuid.MustParse("d2ccc491-c8c3-48e8-867a-fbc82b7018b2"),
					UserId: uid,
					Title:  "Test Theme",
					Analyzers: []Analyzer{
						{
							Key: "word_search",
							Inputs: []AnalyzerInput{
								{
									Key:   "another_text_input",
									Value: "Something else",
								},
							},
						},
					},
					Reporter: Reporter{
						Key:       "word_count",
						Threshold: 1,
					},
				},
			},
			expect: []ThemeDto{
				{
					Id:    uuid.MustParse("d2ccc491-c8c3-48e8-867a-fbc82b7018b2"),
					Title: "Test Theme",
					Analyzers: []AnalyzerDto{
						{
							Key:          "word_search",
							Name:         "Word Search Analyzer",
							Description:  "Uses a basic word list to scan content for.",
							ChangeStatus: Changed,
							Inputs: []AnalyzerInputDto{
								{
									Key:          "another_text_input",
									Value:        "Something else",
									Name:         "",
									Description:  "",
									Type:         "",
									ChangeStatus: Removed,
								},
								{
									Key:          "strict_words",
									Value:        "",
									Name:         "Strict Words",
									Description:  "Words in this list will immediatly flag the content.",
									Type:         "textarea",
									ChangeStatus: New,
								},
							},
						},
					},
					Reporters: []ReporterDto{
						{
							Key:          "word_count",
							Threshold:    1,
							Name:         "Word Count",
							Description:  "Basic reporter that will check basic counts.",
							ChangeStatus: Same,
						},
					},
				},
				defaultEmptyThemeDto,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mts := NewMockthemeStore(t)
			mts.EXPECT().getAllThemesByUserId(uid).Return(tt.input, nil)
			ts := NewThemeService(mts)

			res, err := ts.GetAllThemesByUserId(uid)
			assert.NoError(t, err)
			assert.Equal(t, tt.expect, res)
		})
	}

}

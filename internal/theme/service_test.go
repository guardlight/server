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
							Key:         "word_search",
							Name:        "Word Search",
							Description: "Description of analyzer",
							Inputs: []AnalyzerInput{
								{
									Key:   "strict_words",
									Value: "Running, Walking",
								},
							},
						},
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
							Name:         "Word Search",
							Description:  "Description of analyzer",
							ChangeStatus: Same,
							Inputs: []AnalyzerInputDto{
								{
									Key:          "strict_words",
									Value:        "Running, Walking",
									ChangeStatus: Same,
								},
							},
						},
					},
				},
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
									ChangeStatus: New,
								},
							},
						},
					},
				},
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
							Key:         "super_ai",
							Name:        "Super AI analyzer",
							Description: "Description of super AI analyzer.",
							Inputs: []AnalyzerInput{
								{
									Key:   "command_prompt",
									Value: "Is there walking?",
								},
							},
						},
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
							Name:         "Super AI analyzer",
							Description:  "Description of super AI analyzer.",
							ChangeStatus: Removed,
							Inputs: []AnalyzerInputDto{
								{
									Key:          "command_prompt",
									Value:        "Is there walking?",
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
									ChangeStatus: New,
								},
							},
						},
					},
				},
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
							Key:         "word_search",
							Name:        "Word Search",
							Description: "Description of analyzer",
							Inputs: []AnalyzerInput{
								{
									Key:   "another_text_input",
									Value: "Something else",
								},
							},
						},
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
							Name:         "Word Search",
							Description:  "Description of analyzer",
							ChangeStatus: Changed,
							Inputs: []AnalyzerInputDto{
								{
									Key:          "another_text_input",
									Value:        "Something else",
									ChangeStatus: Removed,
								},
								{
									Key:          "strict_words",
									Value:        "",
									ChangeStatus: New,
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mts := NewMockthemeStore(t)
			mts.EXPECT().getAllThemesByUserId(uid).Return(tt.input, nil)
			ts := NewThemeService(mts)

			res, err := ts.getAllThemesByUserId(uid)
			assert.NoError(t, err)
			assert.Equal(t, tt.expect, res)
		})
	}

}

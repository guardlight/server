package analysismanager

import (
	"github.com/google/uuid"
	"github.com/guardlight/server/internal/essential/config"
	"github.com/guardlight/server/internal/theme"
	"github.com/guardlight/server/pkg/analysisresult"
	"github.com/samber/lo"
)

type analysisGetter interface {
	getAnalysesByUserId(id uuid.UUID) ([]AnalysisRequest, error)
}

type themeService interface {
	GetAllThemesByUserId(id uuid.UUID) ([]theme.ThemeDto, error)
}

type AnalysisResultService struct {
	ag analysisGetter
	ts themeService
}

func NewAnalysisResultService(ag analysisGetter, ts themeService) *AnalysisResultService {
	return &AnalysisResultService{
		ag: ag,
		ts: ts,
	}
}

func (ars *AnalysisResultService) GetAnalysesByUserId(id uuid.UUID) ([]analysisresult.Analysis, error) {
	as, err := ars.ag.getAnalysesByUserId(id)
	if err != nil {
		return nil, err
	}

	if len(as) == 0 {
		return []analysisresult.Analysis{}, nil
	}

	ts, err := ars.ts.GetAllThemesByUserId(id)
	if err != nil {
		return nil, err
	}

	return lo.Map(as, func(ar AnalysisRequest, _ int) analysisresult.Analysis {
		return mapToAnalysisResult(ar, ts)
	}), nil

}

func mapToAnalysisResult(ar AnalysisRequest, ts []theme.ThemeDto) analysisresult.Analysis {
	themes := []analysisresult.Theme{}

	themeMap := make(map[uuid.UUID][]analysisresult.Analyzer)
	for _, a := range ar.Analysis {
		themeMap[a.ThemeId] = append(themeMap[a.ThemeId], mapToAnalyzerToResult(a))
	}

	for tk, tv := range themeMap {
		t, ok := lo.Find(ts, func(t theme.ThemeDto) bool { return t.Id == tk })

		themes = append(themes, analysisresult.Theme{
			Id:        tk,
			Title:     lo.If(ok, t.Title).Else("Theme unknown"),
			Analyzers: tv,
		})
	}

	a := analysisresult.Analysis{
		Id:          ar.Id,
		Title:       ar.Title,
		ContentType: ar.ContentType,
		Themes:      themes,
	}

	return a
}

func mapToAnalyzerToResult(a Analysis) analysisresult.Analyzer {
	aName := a.AnalyzerKey
	ac, ok := config.Get().GetAnalyzer(a.AnalyzerKey)
	if ok {
		aName = ac.Name
	}

	return analysisresult.Analyzer{
		Key:     a.AnalyzerKey,
		Name:    aName,
		Status:  string(a.Status),
		Score:   a.Score,
		Content: a.Content,
		Inputs: lo.Map(a.Inputs, func(i AnalysisInput, _ int) analysisresult.AnalyzerInput {
			iName := i.Key
			if in, ok := lo.Find(ac.Inputs, func(inp config.AnalyzerInput) bool { return inp.Key == i.Key }); ok {
				iName = in.Name
			}
			return analysisresult.AnalyzerInput{Key: i.Key, Name: iName, Value: i.Value}
		}),
		Jobs: lo.Map(a.Jobs, func(j SingleJobProgress, _ int) analysisresult.AnalyzerJobProgress {
			return analysisresult.AnalyzerJobProgress{Status: string(j.Status)}
		}),
	}
}

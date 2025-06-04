package analysismanager

import (
	"github.com/google/uuid"
	"github.com/guardlight/server/internal/essential/config"
	"github.com/guardlight/server/internal/theme"
	"github.com/guardlight/server/pkg/analysisresult"
	"github.com/samber/lo"
)

type analysisGetter interface {
	getAnalysesByUserId(id uuid.UUID, pag Pagination) (AnalysisResultPaginated, error)
	getAnalysesByAnalysisIdAndUserId(id, arid uuid.UUID) (AnalysisRequest, error)
}

type analysisUpdater interface {
	updateScore(id uuid.UUID, score float32) error
}

type themeService interface {
	GetAllThemesByUserId(id uuid.UUID) ([]theme.ThemeDto, error)
}

type AnalysisResultService struct {
	ag analysisGetter
	au analysisUpdater
	ts themeService
}

func NewAnalysisResultService(ag analysisGetter, au analysisUpdater, ts themeService) *AnalysisResultService {
	return &AnalysisResultService{
		ag: ag,
		au: au,
		ts: ts,
	}
}

func (ars *AnalysisResultService) GetAnalysesByUserId(id uuid.UUID, limit, page int) (analysisresult.AnalysisPaginated, error) {
	as, err := ars.ag.getAnalysesByUserId(id, Pagination{Limit: limit, Page: page})
	if err != nil {
		return analysisresult.AnalysisPaginated{}, err
	}

	if len(as.Requests) == 0 {
		return analysisresult.AnalysisPaginated{
			Limit:      as.Limit,
			Page:       as.Page,
			TotalPages: as.TotalPages,
			Analyses:   []analysisresult.Analysis{},
		}, nil
	}

	ts, err := ars.ts.GetAllThemesByUserId(id)
	if err != nil {
		return analysisresult.AnalysisPaginated{}, err
	}

	mpAns := lo.Map(as.Requests, func(ar AnalysisRequest, _ int) analysisresult.Analysis {
		return mapToAnalysisResult(ar, ts)
	})

	return analysisresult.AnalysisPaginated{
		Limit:      as.Limit,
		Page:       as.Page,
		TotalPages: as.TotalPages,
		Analyses:   mpAns,
	}, nil

}

func (ars *AnalysisResultService) GetAnalysesByAnalysisIdAndUserId(uid, arid uuid.UUID) (analysisresult.Analysis, error) {
	ar, err := ars.ag.getAnalysesByAnalysisIdAndUserId(uid, arid)
	if err != nil {
		return analysisresult.Analysis{}, err
	}

	ts, err := ars.ts.GetAllThemesByUserId(uid)
	if err != nil {
		return analysisresult.Analysis{}, err
	}

	return mapToAnalysisResult(ar, ts), nil
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
		CreatedAt:   ar.CreatedAt,
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
		Id:      a.Id,
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

func (ars *AnalysisResultService) UpdateScore(id uuid.UUID, score float32) error {
	return ars.au.updateScore(id, score)
}

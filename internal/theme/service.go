package theme

import (
	"github.com/google/uuid"
	"github.com/guardlight/server/internal/essential/config"
	"github.com/samber/lo"
)

type themeStore interface {
	getAllThemesByUserId(id uuid.UUID) ([]Theme, error)
	updateTheme(t *Theme, uid uuid.UUID) error
}

type ThemeService struct {
	ts themeStore
}

func NewThemeService(ts themeStore) *ThemeService {
	return &ThemeService{
		ts: ts,
	}
}

func (ts *ThemeService) updateTheme(tDto ThemeDto, uid uuid.UUID) error {
	t := mapDtoToEntity(tDto, uid)
	err := ts.ts.updateTheme(&t, uid)
	if err != nil {
		return err
	}

	return nil
}

func mapDtoToEntity(tDto ThemeDto, uid uuid.UUID) Theme {
	rep := lo.FindOrElse(tDto.Reporters, ReporterDto{Key: "none", Threshold: 0}, func(rDto ReporterDto) bool {
		return rDto.ChangeStatus == Same || rDto.ChangeStatus == Changed
	})

	return Theme{
		Id:          tDto.Id,
		UserId:      uid,
		Title:       tDto.Title,
		Description: tDto.Description,
		Reporter: Reporter{
			Key:       rep.Key,
			Threshold: rep.Threshold,
		},
		Analyzers: lo.Map(tDto.Analyzers, mapAnalyzerDtoToEntity),
	}
}

func mapAnalyzerDtoToEntity(aDto AnalyzerDto, _ int) Analyzer {
	return Analyzer{
		Key:    aDto.Key,
		Inputs: lo.Map(aDto.Inputs, mapInputDtoToEntity),
	}
}

func mapInputDtoToEntity(iDto AnalyzerInputDto, _ int) AnalyzerInput {
	return AnalyzerInput{
		Key:   iDto.Key,
		Value: iDto.Value,
	}
}

func (ts *ThemeService) GetAllThemesByUserId(id uuid.UUID) ([]ThemeDto, error) {
	dbThemes, err := ts.ts.getAllThemesByUserId(id)
	if err != nil {
		return nil, err
	}

	// Add an empty theme to populate it with all new fields to the UI
	dbThemes = append(dbThemes, Theme{
		Id:          uuid.Nil,
		UserId:      id,
		Title:       "",
		Description: "",
		Analyzers:   []Analyzer{},
		Reporter:    Reporter{},
	})

	return lo.Map(dbThemes, mergeThemesFromConfig), nil
}

func mergeThemesFromConfig(t Theme, _ int) ThemeDto {
	tDto := ThemeDto{
		Id:          t.Id,
		Title:       t.Title,
		Description: t.Description,
		Reporters:   mergeReporterFromConfig(t.Reporter),
		// Contains all Same, Removed and Changed analyzers
		Analyzers: lo.Map(t.Analyzers, mergeAnalyzerFromConfig),
	}

	// Adds all new analyzers
	for _, aConf := range config.Get().Analyzers {
		if ok := lo.ContainsBy(tDto.Analyzers, func(a AnalyzerDto) bool { return a.Key == aConf.Key }); !ok {
			tDto.Analyzers = append(tDto.Analyzers, AnalyzerDto{
				Key:          aConf.Key,
				Name:         aConf.Name,
				Description:  aConf.Description,
				ChangeStatus: New,
				Inputs:       mergeAnalyzerInputsFromConfig([]AnalyzerInput{}, aConf.Inputs),
			})
		}
	}

	// Adds all new reporters
	for _, rConf := range config.Get().Reporters {
		if ok := lo.ContainsBy(tDto.Reporters, func(r ReporterDto) bool { return r.Key == rConf.Key }); !ok {
			tDto.Reporters = append(tDto.Reporters, ReporterDto{
				Key:          rConf.Key,
				Name:         rConf.Name,
				Description:  rConf.Description,
				ChangeStatus: New,
				Threshold:    0,
			})
		}
	}

	return tDto
}

func mergeReporterFromConfig(r Reporter) []ReporterDto {
	rConf, ok := config.Get().GetReporter(r.Key)

	// Default key if none is empty string
	if len(r.Key) == 0 {
		return []ReporterDto{}
	}

	// Reporter is not in config anymore
	if !ok {
		return []ReporterDto{
			{
				Key:          r.Key,
				ChangeStatus: Removed,
				Threshold:    r.Threshold,
				Name:         "",
				Description:  "",
			},
		}
	}

	return []ReporterDto{
		{
			Key:          rConf.Key,
			Name:         rConf.Name,
			Threshold:    r.Threshold,
			Description:  rConf.Description,
			ChangeStatus: Same,
		},
	}
}

func mergeAnalyzerFromConfig(a Analyzer, _ int) AnalyzerDto {
	ad := AnalyzerDto{
		Key:          a.Key,
		ChangeStatus: Same,
	}

	aConf, ok := config.Get().GetAnalyzer(a.Key)
	// Analyzer is not in Config anymore
	if !ok {
		ad.ChangeStatus = Removed
		ad.Inputs = mergeAnalyzerInputsFromConfig(a.Inputs, []config.AnalyzerInput{})
		return ad
	}

	ad.Inputs = mergeAnalyzerInputsFromConfig(a.Inputs, aConf.Inputs)
	ad.Name = aConf.Name
	ad.Description = aConf.Description

	// Merged inputs contains a Removed or New input
	if ok := lo.ContainsBy(ad.Inputs, func(a AnalyzerInputDto) bool { return a.ChangeStatus == Removed || a.ChangeStatus == New }); ok {
		ad.ChangeStatus = Changed
	}

	return ad

}

func mergeAnalyzerInputsFromConfig(in []AnalyzerInput, inConf []config.AnalyzerInput) []AnalyzerInputDto {
	aiDtos := make([]AnalyzerInputDto, 0)

	// Check saved inputs are Same or Removed
	for _, i := range in {
		aid := AnalyzerInputDto{
			Key:          i.Key,
			Value:        i.Value,
			ChangeStatus: Same,
		}
		if inp, ok := lo.Find(inConf, func(a config.AnalyzerInput) bool { return a.Key == i.Key }); !ok {
			aid.ChangeStatus = Removed
		} else {
			aid.Name = inp.Name
			aid.Description = inp.Description
			aid.Type = inp.Type
		}
		aiDtos = append(aiDtos, aid)
	}

	// Check if new inputs where added
	for _, ic := range inConf {
		if ok := lo.ContainsBy(aiDtos, func(a AnalyzerInputDto) bool { return a.Key == ic.Key }); !ok {
			aiDtos = append(aiDtos, AnalyzerInputDto{
				Key:          ic.Key,
				Value:        "",
				ChangeStatus: New,
				Name:         ic.Name,
				Description:  ic.Description,
				Type:         ic.Type,
			})
		}
	}

	return aiDtos
}

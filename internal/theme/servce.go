package theme

import (
	"github.com/google/uuid"
	"github.com/guardlight/server/internal/essential/config"
	"github.com/samber/lo"
)

type themeStore interface {
	getAllThemesByUserId(id uuid.UUID) ([]Theme, error)
	updateTheme(t Theme, uid uuid.UUID) error
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
	// TODO Validate theme

	t := mapDtoToEntity(tDto, uid)
	err := ts.ts.updateTheme(t, uid)
	if err != nil {
		return err
	}

	return nil
}

func mapDtoToEntity(tDto ThemeDto, uid uuid.UUID) Theme {
	return Theme{
		Id:        tDto.Id,
		UserId:    uid,
		Title:     tDto.Title,
		Analyzers: lo.Map(tDto.Analyzers, mapAnalyzerDtoToEntity),
	}
}

func mapAnalyzerDtoToEntity(aDto AnalyzerDto, _ int) Analyzer {
	return Analyzer{
		Key:         aDto.Key,
		Name:        aDto.Name,
		Description: aDto.Description,
		Inputs:      lo.Map(aDto.Inputs, mapInputDtoToEntity),
	}
}

func mapInputDtoToEntity(iDto AnalyzerInputDto, _ int) AnalyzerInput {
	return AnalyzerInput{
		Key:   iDto.Key,
		Value: iDto.Value,
	}
}

func (ts *ThemeService) getAllThemesByUserId(id uuid.UUID) ([]ThemeDto, error) {
	dbThemes, err := ts.ts.getAllThemesByUserId(id)
	if err != nil {
		return nil, err
	}

	return lo.Map(dbThemes, mergeThemesFromConfig), nil
}

func mergeThemesFromConfig(t Theme, _ int) ThemeDto {
	tDto := ThemeDto{
		Id:    t.Id,
		Title: t.Title,
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

	return tDto
}

func mergeAnalyzerFromConfig(a Analyzer, _ int) AnalyzerDto {
	ad := AnalyzerDto{
		Key:          a.Key,
		Name:         a.Name,
		Description:  a.Description,
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
		if ok := lo.ContainsBy(inConf, func(a config.AnalyzerInput) bool { return a.Key == i.Key }); !ok {
			aid.ChangeStatus = Removed
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
			})
		}
	}

	return aiDtos
}

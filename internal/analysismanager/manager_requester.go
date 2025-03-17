package analysismanager

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/guardlight/server/internal/essential/config"
	"github.com/guardlight/server/internal/jobmanager"
	"github.com/guardlight/server/pkg/analysisrequest"
	"github.com/guardlight/server/pkg/parsercontract"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

var (
	ErrInvalidParser   = errors.New("invalid parser selected")
	ErrInvalidAnalyzer = errors.New("invalid analyzer selected")
	ErrParserMarshal   = errors.New("error marshaling parser data")
)

type analysisRequestStore interface {
	createAnalysisRequest(analysisRequest *AnalysisRequest) error
	getAnalysesByUserId(id uuid.UUID) ([]AnalysisRequest, error)
	getAnalysById(uid uuid.UUID, aid uuid.UUID) (AnalysisRequest, error)
}

type jobManagerRequester interface {
	jobmanager.Enqueuer
	jobmanager.IdCreater
}

type AnalysisManagerRequester struct {
	jobMananger jobManagerRequester
	ars         analysisRequestStore
}

func NewAnalysisManangerRequester(jobMananger jobManagerRequester, ars analysisRequestStore) *AnalysisManagerRequester {
	return &AnalysisManagerRequester{
		jobMananger: jobMananger,
		ars:         ars,
	}
}

func (am *AnalysisManagerRequester) RequestAnalysis(arDto *analysisrequest.AnalysisRequest, ui uuid.UUID) error {

	p, ok := config.Get().GetParser(arDto.File.Mimetype)
	if !ok {
		zap.S().Errorw("Invalid parser specified", "parser_type", arDto.File.Mimetype)
		return ErrInvalidParser
	}

	if !hasValidAnalyzers(arDto) {
		return ErrInvalidAnalyzer
	}

	steps := createSteps(arDto)

	rawData := createRawData(arDto)

	analysisParts := createAnalysis(arDto)

	ar := &AnalysisRequest{
		Id:                   uuid.Nil,
		UserId:               ui,
		Title:                arDto.Title,
		AnalysisRequestSteps: steps,
		RawData:              rawData,
		Analysis:             analysisParts,
	}
	err := am.ars.createAnalysisRequest(ar)
	if err != nil {
		return err
	}

	jobId := am.jobMananger.CreateId()

	jd := jobmanager.ParserJobData{
		Type:  p.Type,
		Topic: fmt.Sprintf("parser.%s", p.Type),
		Image: p.Image,
		ParserData: parsercontract.ParserRequest{
			JobId:      jobId,
			AnalysisId: ar.Id,
			Content:    arDto.File.Content,
		},
	}
	gk := fmt.Sprintf("parser.%s", p.Type)
	am.jobMananger.EnqueueJob(jobId, jobmanager.Parse, gk, jd)

	return nil
}

func createAnalysis(arDto *analysisrequest.AnalysisRequest) []Analysis {
	as := make([]Analysis, 0)

	for _, t := range arDto.Themes {
		for _, a := range t.Analyzers {
			ainputs := lo.Map(a.Inputs, func(inp analysisrequest.AnalyzerInput, _ int) AnalysisInput {
				return AnalysisInput{
					Key:   inp.Key,
					Value: inp.Value,
				}
			})
			as = append(as, Analysis{
				Id:                uuid.Nil,
				AnalysisRequestId: uuid.Nil,
				AnalyzerKey:       a.Key,
				ThemeId:           t.Id,
				Status:            AnalysisWaiting,
				Threshold:         a.Threshold,
				Score:             0,
				Inputs:            ainputs,
				Content:           []string{},
				Jobs:              []SingleJobProgress{},
			})
		}
	}

	return as
}

func createRawData(arDto *analysisrequest.AnalysisRequest) RawData {
	hash := md5.Sum(arDto.File.Content)

	return RawData{
		Id:                uuid.Nil,
		AnalysisRequestId: uuid.Nil,
		Content:           arDto.File.Content,
		FileType:          arDto.File.Mimetype,
		Hash:              hex.EncodeToString(hash[:]),
	}
}

func createSteps(arDto *analysisrequest.AnalysisRequest) []AnalysisRequestStep {

	steps := make([]AnalysisRequestStep, 0)

	steps = append(steps, AnalysisRequestStep{
		Id:                uuid.Nil,
		AnalysisRequestId: uuid.Nil,
		Index:             0,
		StepType:          Create,
		Status:            Finished,
		StatusDescription: "",
	})

	steps = append(steps, AnalysisRequestStep{
		Id:                uuid.Nil,
		AnalysisRequestId: uuid.Nil,
		Index:             1,
		StepType:          Parse,
		Status:            Waiting,
		StatusDescription: "",
	})

	for iTheme, theme := range arDto.Themes {
		for iAnalyzer := range theme.Analyzers {
			steps = append(steps, AnalysisRequestStep{
				Id:                uuid.Nil,
				AnalysisRequestId: uuid.Nil,
				Index:             2 + iTheme + iAnalyzer,
				StepType:          Analyze,
				Status:            Waiting,
				StatusDescription: "",
			})
		}
	}

	steps = append(steps, AnalysisRequestStep{
		Id:                uuid.Nil,
		AnalysisRequestId: uuid.Nil,
		Index:             len(steps),
		StepType:          Report,
		Status:            Waiting,
		StatusDescription: "",
	})

	return steps
}

func hasValidAnalyzers(arDto *analysisrequest.AnalysisRequest) bool {
	analyzersFromRequest := lo.FlatMap(arDto.Themes, func(t analysisrequest.Theme, _ int) []analysisrequest.Analyzer {
		return t.Analyzers
	})

	for _, a := range analyzersFromRequest {
		analyzerFromConfig, ok := config.Get().GetAnalyzer(a.Key)
		if !ok {
			zap.S().Errorw("Invalid analyzer specified", "analyzer_key", a.Key)
			return false
		}

		for _, i := range analyzerFromConfig.Inputs {
			if !lo.ContainsBy(a.Inputs, func(azi analysisrequest.AnalyzerInput) bool { return azi.Key == i.Key }) {
				zap.S().Errorw("Analyzer Defined Input does not exist in Request Analyzer Input", "analyzer_key", a.Key)
				return false
			}

		}
	}

	return true
}

func (am *AnalysisManagerRequester) GetAnalysesByUserId(id uuid.UUID) ([]AnalysisRequest, error) {
	return am.ars.getAnalysesByUserId(id)
}

func (am *AnalysisManagerRequester) GetAnalysById(uid uuid.UUID, aid uuid.UUID) (AnalysisRequest, error) {
	return am.ars.getAnalysById(uid, aid)
}

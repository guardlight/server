package analysismanager

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/guardlight/server/internal/essential/config"
	"github.com/guardlight/server/internal/jobmanager"
	"github.com/guardlight/server/internal/ssemanager"
	"github.com/guardlight/server/internal/theme"
	"github.com/guardlight/server/pkg/analysisrequest"
	"github.com/guardlight/server/pkg/parsercontract"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

var (
	ErrInvalidParser    = errors.New("invalid parser selected")
	ErrInvalidAnalyzer  = errors.New("invalid analyzer selected")
	ErrParserMarshal    = errors.New("error marshaling parser data")
	ErrHashAlreadyExist = errors.New("hash already exist")
)

type analysisRequestStore interface {
	createAnalysisRequest(analysisRequest *AnalysisRequest) error
	getAnalysisRequestIdByHash(hash string) (uuid.UUID, error)
}

type jobManagerRequester interface {
	jobmanager.Enqueuer
	jobmanager.IdCreater
}

type AnalysisManagerRequester struct {
	jobMananger jobManagerRequester
	ars         analysisRequestStore
	ts          themeService
	sse         sseEventSender
}

func NewAnalysisManangerRequester(jobMananger jobManagerRequester, ars analysisRequestStore, sse sseEventSender, ts themeService) *AnalysisManagerRequester {
	return &AnalysisManagerRequester{
		jobMananger: jobMananger,
		ars:         ars,
		sse:         sse,
		ts:          ts,
	}
}

func (am *AnalysisManagerRequester) RequestAnalysisDataloom(ardDto *analysisrequest.AnalysisRequestDataloom, ui uuid.UUID) (uuid.UUID, error) {
	ard := &analysisrequest.AnalysisRequest{
		Title:       ardDto.Title,
		ContentType: ardDto.ContentType,
		Category:    ardDto.Category,
		File:        ardDto.File,
		Themes:      []analysisrequest.Theme{},
	}

	userThemes, err := am.ts.GetAllThemesByUserId(ui)
	if err != nil {
		return uuid.Nil, err
	}

	ard.Themes = lo.FilterMap(userThemes, func(ut theme.ThemeDto, _ int) (analysisrequest.Theme, bool) {
		if lo.ContainsBy(ardDto.ThemeIds, func(ti uuid.UUID) bool { return ti == ut.Id }) {
			return analysisrequest.Theme{
				Id:    ut.Id,
				Title: ut.Title,
				Analyzers: lo.FilterMap(ut.Analyzers, func(ta theme.AnalyzerDto, _ int) (analysisrequest.Analyzer, bool) {
					if ta.ChangeStatus == theme.Same || ta.ChangeStatus == theme.Changed {
						return analysisrequest.Analyzer{
							Key: ta.Key,
							Inputs: lo.FilterMap(ta.Inputs, func(tai theme.AnalyzerInputDto, _ int) (analysisrequest.AnalyzerInput, bool) {
								if tai.ChangeStatus == theme.Same || tai.ChangeStatus == theme.Changed {
									return analysisrequest.AnalyzerInput{
										Key:   tai.Key,
										Value: tai.Value,
									}, true
								}
								return analysisrequest.AnalyzerInput{}, false
							}),
						}, true
					}
					return analysisrequest.Analyzer{}, false
				}),
			}, true
		}
		return analysisrequest.Theme{}, false
	})

	return am.RequestAnalysis(ard, ui, string(RequestOriginDataloom))
}

func (am *AnalysisManagerRequester) RequestAnalysis(arDto *analysisrequest.AnalysisRequest, ui uuid.UUID, requestOrigin string) (uuid.UUID, error) {

	p, ok := config.Get().GetParser(arDto.File.Mimetype)
	if !ok {
		zap.S().Errorw("Invalid parser specified", "parser_type", arDto.File.Mimetype)
		return uuid.Nil, ErrInvalidParser
	}

	if !hasValidAnalyzers(arDto) {
		return uuid.Nil, ErrInvalidAnalyzer
	}

	bContent, err := base64.StdEncoding.DecodeString(arDto.File.Content)
	if err != nil {
		return uuid.Nil, err
	}
	rawData := createRawData(arDto, bContent)

	arid, err := am.ars.getAnalysisRequestIdByHash(rawData.Hash)
	if err != nil {
		return uuid.Nil, err
	}

	if arid != uuid.Nil {
		return arid, ErrHashAlreadyExist
	}

	analysisParts := createAnalysis(arDto)

	ar := &AnalysisRequest{
		Id:            uuid.Nil,
		UserId:        ui,
		Title:         arDto.Title,
		RequestOrigin: requestOrigin,
		Category:      arDto.Category,
		ContentType:   string(arDto.ContentType),
		RawData:       rawData,
		Analysis:      analysisParts,
	}
	err = am.ars.createAnalysisRequest(ar)
	if err != nil {
		return uuid.Nil, err
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
	err = am.jobMananger.EnqueueJob(jobId, jobmanager.Parse, gk, jd)
	if err != nil {
		return uuid.Nil, err
	}

	am.sse.SendEvent(ui, ssemanager.SseEvent{
		Type:   ssemanager.TypeUpdate,
		Action: ssemanager.ActionAnalysisRequested,
		Data:   ar.Id.String(),
	})

	return ar.Id, nil
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
				Score:             0,
				Inputs:            ainputs,
				Content:           []string{},
				Jobs:              []SingleJobProgress{},
			})
		}
	}

	return as
}

func createRawData(arDto *analysisrequest.AnalysisRequest, bdata []byte) RawData {
	hash := md5.Sum(bdata)

	return RawData{
		Id:                uuid.Nil,
		AnalysisRequestId: uuid.Nil,
		Content:           bdata,
		FileType:          arDto.File.Mimetype,
		Hash:              hex.EncodeToString(hash[:]),
	}
}

func hasValidAnalyzers(arDto *analysisrequest.AnalysisRequest) bool {
	// Check if Analyzer:Input:Threshold is part of request
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

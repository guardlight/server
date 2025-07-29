package parsers

import (
	"encoding/json"

	"github.com/guardlight/server/pkg/parsercontract"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type subripSubtitleParser struct {
	ncon *nats.Conn
}

func NewSubripSubtitleParser(ncon *nats.Conn) *subripSubtitleParser {
	srtp := &subripSubtitleParser{
		ncon: ncon,
	}
	ncon.Subscribe("parser.srt", srtp.parseSubripSubtitle)
	return srtp
}

func (srtp *subripSubtitleParser) parseSubripSubtitle(m *nats.Msg) {
	var pr parsercontract.ParserRequest
	err := json.Unmarshal(m.Data, &pr)
	if err != nil {
		srtp.makeParserErrorResponse(&pr, err)
		return
	}

	var sc = "EMPTY_ALPHA"

	presp := parsercontract.ParserResponse{
		JobId:      pr.JobId,
		AnalysisId: pr.AnalysisId,
		Text:       sc,
		Status:     parsercontract.ParseSuccess,
	}

	dat, err := json.Marshal(presp)
	if err != nil {
		srtp.makeParserErrorResponse(&pr, err)
		return
	}

	srtp.ncon.Publish("parser.result", dat)
}

func (fp *subripSubtitleParser) makeParserErrorResponse(pr *parsercontract.ParserRequest, err error) {
	presp := parsercontract.ParserResponse{
		JobId:      pr.JobId,
		AnalysisId: pr.AnalysisId,
		Text:       err.Error(),
		Status:     parsercontract.ParseError,
	}
	dat, err := json.Marshal(presp)
	if err != nil {
		zap.S().Errorw("Could not marshal parser error response", "error", err)
		return
	}

	fp.ncon.Publish("parser.result", dat)
}

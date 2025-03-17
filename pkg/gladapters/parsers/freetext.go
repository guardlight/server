package parsers

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/guardlight/server/pkg/parsercontract"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type freetextParser struct {
	ncon *nats.Conn
}

func NewFreetextParser(ncon *nats.Conn) *freetextParser {
	fp := &freetextParser{
		ncon: ncon,
	}
	ncon.Subscribe("parser.freetext", fp.parseFreetext)
	return fp
}

func (fp *freetextParser) parseFreetext(m *nats.Msg) {
	var pr parsercontract.ParserRequest
	err := json.Unmarshal(m.Data, &pr)
	if err != nil {
		fp.makeParserErrorResponse(&pr, err)
		return
	}

	sc := parse(pr.Content)

	presp := parsercontract.ParserResponse{
		JobId:      pr.JobId,
		AnalysisId: pr.AnalysisId,
		Text:       sc,
		Status:     parsercontract.ParseSuccess,
	}

	dat, err := json.Marshal(presp)
	if err != nil {
		fp.makeParserErrorResponse(&pr, err)
		return
	}

	fp.ncon.Publish("parser.result", dat)
}

func parse(data []byte) string {
	cleanText := strings.ReplaceAll(string(data), "\n", " ")
	cleanText = strings.ReplaceAll(cleanText, "\r", " ")

	// Use regex to replace multiple spaces with a single space
	re := regexp.MustCompile(`\s+`)
	cleanText = re.ReplaceAllString(cleanText, " ")

	return cleanText
}

func (fp *freetextParser) makeParserErrorResponse(pr *parsercontract.ParserRequest, err error) {
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

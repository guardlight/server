package analyzers

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"

	"github.com/guardlight/server/pkg/analyzercontract"
	"github.com/nats-io/nats.go"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

const (
	INPUT_KEY_STRICT_WORDS = "strict_words"
)

type wordsearchAnalyzer struct {
	ncon *nats.Conn
}

func NewWordsearchAnalyzer(ncon *nats.Conn) *wordsearchAnalyzer {
	wa := &wordsearchAnalyzer{
		ncon: ncon,
	}
	ncon.Subscribe("analyzer.word_search", wa.analyze)
	return wa
}

func (wa *wordsearchAnalyzer) analyze(m *nats.Msg) {
	var ar analyzercontract.AnalyzerRequest
	err := json.Unmarshal(m.Data, &ar)
	if err != nil {
		wa.makeParserErrorResponse(&ar, err)
		return
	}
	zap.S().Info(ar.Content)
	res, score, err := analyze(ar.Content, ar.Inputs)
	if err != nil {
		wa.makeParserErrorResponse(&ar, err)
		return
	}

	aresp := analyzercontract.AnalyzerResponse{
		JobId:      ar.JobId,
		AnalysisId: ar.AnalysisId,
		Results:    res,
		Status:     analyzercontract.AnalyzerSuccess,
		Score:      score,
	}
	dat, err := json.Marshal(aresp)
	if err != nil {
		zap.S().Errorw("Could not marshal analyzer response", "error", err)
		return
	}

	wa.ncon.Publish("analyzer.result", dat)
}

func (wa *wordsearchAnalyzer) makeParserErrorResponse(ar *analyzercontract.AnalyzerRequest, err error) {
	aresp := analyzercontract.AnalyzerResponse{
		JobId:      ar.JobId,
		AnalysisId: ar.AnalysisId,
		Results: []string{
			err.Error(),
		},
		Status: analyzercontract.AnalyzerError,
	}
	dat, err := json.Marshal(aresp)
	if err != nil {
		zap.S().Errorw("Could not marshal analyzer error response", "error", err)
		return
	}

	wa.ncon.Publish("analyzer.result", dat)
}

func analyze(text string, ins []analyzercontract.AnalysisInput) ([]string, float32, error) {
	in, ok := lo.Find(ins, func(item analyzercontract.AnalysisInput) bool {
		return item.Key == INPUT_KEY_STRICT_WORDS
	})
	if !ok {
		return nil, 0, errors.New("strict_words key not found in data")
	}

	strWordsMapper := func(str string, _ int) string {
		return strings.ToLower(strings.TrimSpace(str))
	}
	splStrictWords := lo.Map(strings.Split(in.Value, ","), strWordsMapper)

	// match sentence-ending punctuation (including optional closing quote)
	re := regexp.MustCompile(`([.!?][")]?)(\s+|$)`)

	// Find matches
	matches := re.FindAllStringIndex(text, -1)

	var splText []string
	start := 0

	for _, match := range matches {
		end := match[1]
		splText = append(splText, strings.TrimSpace(text[start:end]))
		start = end
	}

	sents := make([]string, 0)

	for i, sentance := range splText {
		for _, sw := range splStrictWords {
			pattern := `\b` + regexp.QuoteMeta(strings.ToLower(sw)) + `\b`
			re := regexp.MustCompile(pattern)
			if re.MatchString(strings.ToLower(sentance)) {
				sents = append(sents, buildSent(splText, i))
			}
		}
	}

	score := func() float32 {
		if len(sents) == 0 {
			return -1
		}
		return 2*float32(len(sents))/float32(len(splText)) - 1
	}()

	return sents, score, nil
}

func buildSent(splText []string, ind int) string {
	sents := make([]string, 0)
	// if ind != 0 {
	// 	sents = append(sents, fmt.Sprintf("...%s", lo.Substring(splText[ind-1], 0, 20)))
	// }
	sents = append(sents, splText[ind])
	// if ind != len(splText) {
	// 	sents = append(sents, fmt.Sprintf("%s...", lo.Substring(splText[ind+1], 0, 20)))
	// }
	// tr := lo.Map(sents, func(s string, _ int) string {
	// 	return strings.Trim(strings.Trim(s, "\n"), "\n")
	// })
	return strings.Join(sents, " ")
}

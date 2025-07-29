package reporters

import (
	"encoding/json"

	"github.com/guardlight/server/pkg/reportercontract"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type wordcountReporter struct {
	ncon *nats.Conn
}

func NewWordcountReporter(ncon *nats.Conn) *wordcountReporter {
	wr := &wordcountReporter{
		ncon: ncon,
	}
	ncon.Subscribe("reporter.word_count", wr.report)
	return wr
}

func (wr *wordcountReporter) report(m *nats.Msg) {
	var rr reportercontract.ReporterRequest
	err := json.Unmarshal(m.Data, &rr)
	if err != nil {
		wr.makeReporterErrorResponse(&rr, err)
		return
	}

	var score float32
	score = 0

	if len(rr.Contents) == 0 {
		score = 1
	} else {
		score = -1
	}

	aresp := reportercontract.ReporterResponse{
		JobId:      rr.JobId,
		AnalysisId: rr.AnalysisId,
		Score:      score,
		Comments:   "",
		Status:     reportercontract.ReportSuccess,
	}
	dat, err := json.Marshal(aresp)
	if err != nil {
		zap.S().Errorw("Could not marshal analyzer response", "error", err)
		return
	}

	wr.ncon.Publish("reporter.result", dat)
}

func (wr *wordcountReporter) makeReporterErrorResponse(rr *reportercontract.ReporterRequest, err error) {
	aresp := reportercontract.ReporterResponse{
		JobId:      rr.JobId,
		AnalysisId: rr.AnalysisId,
		Score:      -1,
		Comments:   err.Error(),
		Status:     reportercontract.ReportError,
	}
	dat, err := json.Marshal(aresp)
	if err != nil {
		zap.S().Errorw("Could not marshal reporter error response", "error", err)
		return
	}

	wr.ncon.Publish("reporter.result", dat)
}

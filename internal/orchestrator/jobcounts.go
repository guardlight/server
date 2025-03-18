package orchestrator

import "github.com/guardlight/server/internal/jobmanager"

type jobCounts map[string]int

func jc() map[string]int {
	return make(map[string]int)
}

func (j jobCounts) build(js []jobmanager.Job) {
	clear(j)
	for _, job := range js {
		if job.Status == jobmanager.Inprogress {
			j[job.GroupKey]++
		}
	}
}

func (j jobCounts) t(t string) int {
	return j[t] + 1
}

func (j jobCounts) inc(t string) {
	j[t]++
}

func (j jobCounts) dec(t string) {
	j[t]--
}

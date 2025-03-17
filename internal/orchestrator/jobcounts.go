package orchestrator

import "github.com/guardlight/server/internal/jobmanager"

type jobCounts map[string]int

func jc() map[string]int {
	return make(map[string]int)
}

func (jc jobCounts) build(j []jobmanager.Job) {
	for _, job := range j {
		if job.Status == jobmanager.Inprogress {
			jc[job.GroupKey]++
		}
	}
}

func (jc jobCounts) inc(t string) {
	jc[t]++
}

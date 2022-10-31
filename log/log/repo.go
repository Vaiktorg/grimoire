package log

import (
	"sort"
)

type LogCache []Log

func (m LogCache) FilterByLevel(level string) (ret []Log) {
	for _, t := range m {
		if t.Level == level {
			ret = append(ret, t)
		}
	}
	sort.Slice(ret, func(i, j int) bool {
		return ret[i].Level < ret[j].Level
	})
	return
}

func (m LogCache) FilterByService(service string) (ret []Log) {
	for _, t := range m {
		if t.Service == service {
			ret = append(ret, t)
		}
	}
	sort.Slice(m, func(i, j int) bool {
		return m[i].Service < m[j].Service
	})
	return
}

func (m *LogCache) Write(msg Log) {
	*m = append(*m, msg)
}

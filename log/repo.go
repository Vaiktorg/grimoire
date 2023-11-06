package log

import (
	"sort"
)

type Cache []Log

func (m *Cache) FilterByLevel(level string) (ret []Log) {
	for _, t := range *m {
		if t.Level == level {
			ret = append(ret, t)
		}
	}
	sort.Slice(ret, func(i, j int) bool {
		return ret[i].Level < ret[j].Level
	})
	return
}

func (m *Cache) FilterByService(service string) (ret []Log) {
	for _, t := range *m {
		if t.Service == service {
			ret = append(ret, t)
		}
	}
	sort.Slice(ret, func(i, j int) bool {
		return ret[i].Service < ret[j].Service
	})
	return
}

func (m *Cache) Write(msg Log) {
	*m = append(*m, msg)
}

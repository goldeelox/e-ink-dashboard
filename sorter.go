package main

import "sort"

type lessFunc func(p1, p2 *AgendaEvent) bool

type multiSorter struct {
	processedEvents []*AgendaEvent
	less            []lessFunc
}

func OrderedBy(less ...lessFunc) *multiSorter {
	return &multiSorter{
		less: less,
	}
}

func (ms *multiSorter) Len() int {
	return len(ms.processedEvents)
}

func (ms *multiSorter) Less(i, j int) bool {
	p, q := ms.processedEvents[i], ms.processedEvents[j]
	var k int
	for k = 0; k < len(ms.less)-1; k++ {
		less := ms.less[k]
		switch {
		case less(p, q):
			return true
		case less(q, p):
			return false
		}
	}
	return ms.less[k](p, q)
}

func (ms *multiSorter) Sort(processedEvents []*AgendaEvent) {
	ms.processedEvents = processedEvents
	sort.Sort(ms)
}

func (ms *multiSorter) Swap(i, j int) {
	ms.processedEvents[i], ms.processedEvents[j] = ms.processedEvents[j], ms.processedEvents[i]
}

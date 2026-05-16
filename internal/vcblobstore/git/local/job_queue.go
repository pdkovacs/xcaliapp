package local

import (
	"sync"
)

var in = make(chan func())

func queueProcessor() {
	for job := range in {
		job()
	}
}

func Enqueue(job func()) {
	var wg sync.WaitGroup
	wg.Add(1)
	in <- func() {
		job()
		wg.Done()
	}
	wg.Wait()
}

func init() {
	go queueProcessor()
}

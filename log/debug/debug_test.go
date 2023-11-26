package main

import (
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
)

func TestMain(m *testing.M) {
	m.Run()
}

const numLogs = 10000

func BenchmarkLogger(b *testing.B) {
	receivedMain := uint64(0)
	receivedServ := uint64(0)

	l := NewLogger()
	defer l.Close()

	s := l.NewService()

	var wg sync.WaitGroup
	go l.Output(func(log Log) {
		atomic.AddUint64(&receivedMain, 1)
		wg.Done()
	})

	go s.Output(func(log Log) {
		atomic.AddUint64(&receivedServ, 1)
		wg.Done()
	})

	wg.Add(numLogs * 2)

	b.ResetTimer() // Start benchmark timing
	for i := 0; i < numLogs; i++ {
		s.Log(strconv.Itoa(i))
	}
	wg.Wait()     // Wait for all logs to be processed
	b.StopTimer() // Stop benchmark timing

	// Optionally, print out results for verification
	b.Logf("Received Main: %d, Received Serv: %d", receivedMain, receivedServ)
}

func TestLogger_Output(t *testing.T) {
	t.Run("Both Services", func(t *testing.T) {
		received := uint64(0)

		l := NewLogger()
		sl := l.NewService()

		defer l.Close()

		wg := new(sync.WaitGroup)
		go l.Output(func(log Log) {
			atomic.AddUint64(&received, 1)
			wg.Done()
		})

		go sl.Output(func(log Log) {
			atomic.AddUint64(&received, 1)
			wg.Done()
		})

		wg.Add(numLogs * 2)
		for i := 0; i < numLogs; i++ {
			sl.Log(strconv.Itoa(i))
		}

		wg.Wait()

		if received != uint64(numLogs*2) {
			t.Fatalf("%d (sum all services) / %d", numLogs, received)
		}
	})
}

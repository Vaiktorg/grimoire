package tests

import (
	"github.com/vaiktorg/grimoire/log"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestLoggerOutput(t *testing.T) {
	t.Cleanup(cleanup)
	t.Run("MainLoggerOutputOnly", func(t *testing.T) {
		mainLogger := log.NewLogger(log.Config{ServiceName: "MainService", CanOutput: true})
		defer mainLogger.Close()
		serviceLogger := mainLogger.NewServiceLogger(log.Config{ServiceName: "ServiceLogger", CanOutput: true})

		testMessage := "Test log message"
		received := make(chan bool, 1) // Use buffered channel to avoid blocking

		go mainLogger.Output(func(logEntry log.Log) {
			if logEntry.Msg == testMessage {
				received <- true
			}
		})

		serviceLogger.INFO(testMessage) // This should not appear in mainLogger's output

		select {
		case <-received:
		case <-time.After(sleepTime):
			t.Errorf("Log from service logger was not outputted")
			// Test passes as no log is received
		}
	})

	t.Run("ServiceLoggerOutputOnly", func(t *testing.T) {
		mainLogger := log.NewLogger(log.Config{ServiceName: "MainService", CanOutput: true})
		defer mainLogger.Close()

		serviceLogger := mainLogger.NewServiceLogger(log.Config{ServiceName: "ServiceLogger", CanOutput: true})

		testMessage := "Test log message"
		received := make(chan bool, 1) // Use buffered channel to avoid blocking

		go serviceLogger.Output(func(logEntry log.Log) {
			if logEntry.Msg == testMessage {
				received <- true
			}
		})

		serviceLogger.INFO(testMessage) // This should appear in serviceLogger's output only

		select {
		case <-received:
			// Test passes as log is received
		case <-time.After(sleepTime):
			t.Errorf("Log from service logger should be outputted")
		}
	})

	t.Run("BothLoggersCanOutput", func(t *testing.T) {
		mainLogger := log.NewLogger(log.Config{ServiceName: "MainService", CanOutput: true})
		defer mainLogger.Close()

		serviceLogger := mainLogger.NewServiceLogger(log.Config{ServiceName: "ServiceLogger", CanOutput: true})

		testMessage := "Test log message"
		mainReceived := make(chan bool, 1)
		serviceReceived := make(chan bool, 1)

		// OnMessage to main logger output
		go mainLogger.Output(func(logEntry log.Log) {
			if logEntry.Msg == testMessage {
				mainReceived <- true
			}
		})

		// OnMessage to service logger output
		go serviceLogger.Output(func(logEntry log.Log) {
			if logEntry.Msg == testMessage {
				serviceReceived <- true
			}
		})

		serviceLogger.INFO(testMessage) // Log should appear in both outputs

		select {
		case <-mainReceived:
			// Main logger received the log
		case <-time.After(sleepTime):
			t.Errorf("Main logger did not output the log")
		}

		select {
		case <-serviceReceived:
			// Service logger received the log
		case <-time.After(sleepTime):
			t.Errorf("Service logger did not output the log")
		}
	})

	t.Run("MainLoggerBatchOutput", func(t *testing.T) {
		mainLogger := log.NewLogger(log.Config{ServiceName: "MainService", CanOutput: true})
		defer mainLogger.Close()

		batchCount := totalLogAmount
		receivedCount := int64(0)

		wg := new(sync.WaitGroup)
		go mainLogger.Output(func(log log.Log) {
			defer wg.Done()
			atomic.AddInt64(&receivedCount, 1)
		})

		wg.Add(batchCount)
		for i := 0; i < batchCount; i++ {
			mainLogger.INFO("Test batch log message")
		}

		// Allow some time for processing
		wg.Wait()

		if atomic.LoadInt64(&receivedCount) != int64(batchCount) {
			t.Errorf("Expected %d logs, got %d", batchCount, receivedCount)
		}
	})

	t.Run("ServiceLoggerBatchOutput", func(t *testing.T) {
		mainLogger := log.NewLogger(log.Config{ServiceName: "MainService", CanOutput: true})
		defer mainLogger.Close()

		serviceLogger := mainLogger.NewServiceLogger(log.Config{ServiceName: "ServiceLogger", CanOutput: true})

		batchCount := totalLogAmount
		receivedCount := uint64(0)

		wg := new(sync.WaitGroup)
		go serviceLogger.Output(func(log log.Log) {
			atomic.AddUint64(&receivedCount, 1)
			wg.Done()
		})

		wg.Add(batchCount)
		for i := 0; i < batchCount; i++ {
			serviceLogger.INFO("Test batch log message")
		}

		// Allow some time for processing
		wg.Wait()

		if atomic.LoadUint64(&receivedCount) != uint64(batchCount) {
			t.Errorf("Expected %d logs, got %d", batchCount, receivedCount)
		}
	})

	t.Run("MainLoggerWithServiceLoggerBatchOutput", func(t *testing.T) {
		mainLogger := log.NewLogger(log.Config{ServiceName: "MainService", CanOutput: true})
		defer mainLogger.Close()

		serviceLogger := mainLogger.NewServiceLogger(log.Config{ServiceName: "ServiceLogger", CanOutput: true})

		batchCount := totalLogAmount
		totalBatchCount := batchCount * 2 // Since both main and service logger will log
		receivedCount := int64(0)

		wg := new(sync.WaitGroup)
		go mainLogger.Output(func(log log.Log) {
			defer wg.Done()
			atomic.AddInt64(&receivedCount, 1)
		})

		wg.Add(totalBatchCount)
		for i := 0; i < batchCount; i++ {
			mainLogger.INFO("Test main logger batch log message")
			serviceLogger.INFO("Test service logger batch log message")
		}

		wg.Wait()

		// Allow some time for processing
		if atomic.LoadInt64(&receivedCount) != int64(totalBatchCount) {
			t.Errorf("Expected %d logs (from both main and service logger), got %d", totalBatchCount, receivedCount)
		}
	})
}

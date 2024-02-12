package tests

import (
	"errors"
	"github.com/vaiktorg/grimoire/log"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestLogger(t *testing.T) {
	logger := log.NewLogger(&log.Config{ServiceName: "MainLogger", CanOutput: true, Persist: true})

	t.Cleanup(cleanup)

	var serviceLogger log.ILogger
	t.Run("NewServiceLogger", func(t *testing.T) {
		serviceLogger = logger.NewServiceLogger(&log.Config{ServiceName: "ServiceLogger", CanOutput: true, Persist: true})
		serviceName := serviceLogger.ServiceName()

		if serviceLogger.ServiceName() != serviceName {
			t.Errorf("ServiceName() = %v, want %v", serviceLogger.ServiceName(), serviceName)
		}
	})

	t.Run("LogGeneration", func(t *testing.T) {
		serviceLogger.INFO("Test Info", "Test Data")

		time.Sleep(sleepTime)

		msgs := serviceLogger.Messages(log.Pagination{Page: 1, Amount: 1})
		l := len(msgs)
		if l == 0 {
			t.Error("Expected at least one log entry")
		}
	})

	t.Run("ConcurrentLogging", func(t *testing.T) {
		wg := new(sync.WaitGroup)
		numMessages := 100

		wg.Add(numMessages)
		for i := 0; i < numMessages; i++ {
			go serviceLogger.DEBUG("Concurrent Log", i)
			wg.Done()
		}

		wg.Wait()

		time.Sleep(sleepTime)
		messages := serviceLogger.Messages(log.Pagination{Page: 1, Amount: numMessages})

		if len(messages) < numMessages {
			t.Errorf("Expected 100 log entries from concurrent logging; got %d", len(messages))
		}
	})

	t.Run("LoggerOutput", func(t *testing.T) {
		wg := new(sync.WaitGroup)

		wg.Add(102)
		rec := uint64(0)

		go serviceLogger.Output(func(l log.Log) error {
			if l.Service != serviceLogger.ServiceName() {
				return errors.New("received l from unexpected service")
			}

			atomic.AddUint64(&rec, 1)
			wg.Done()

			return nil
		})

		serviceLogger.INFO("Test for Output Channel", "Test Data")

		wg.Wait()

		if atomic.LoadUint64(&rec) != 102 {
			t.Fatalf("not received total sent logs in test")
		}
	})

	t.Run("TotalSent", func(t *testing.T) {
		totalSentBefore := logger.TotalSent()

		logger.DEBUG("Test Total Sent", "Test Data")

		time.Sleep(50 * time.Millisecond) // Allow time for log processing
		if logger.TotalSent() <= totalSentBefore {
			t.Error("Expected TotalSent to increase after logging")
		}
	})

	logger.Close()
}

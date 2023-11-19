package tests

import (
	"errors"
	"github.com/vaiktorg/grimoire/log"
	"sync"
	"testing"
	"time"
)

func TestLogger(t *testing.T) {
	logger := log.NewLogger(log.Config{ServiceName: "MainLogger"})
	defer logger.Close()

	t.Cleanup(cleanup)

	var serviceLogger log.ILogger
	t.Run("NewServiceLogger", func(t *testing.T) {
		serviceName := "ServiceLogger"
		serviceLogger = logger.NewServiceLogger(log.Config{ServiceName: serviceName, CanOutput: true})

		if serviceLogger.ServiceName() != serviceName {
			t.Errorf("ServiceName() = %v, want %v", serviceLogger.ServiceName(), serviceName)
		}
	})

	t.Run("LogGeneration", func(t *testing.T) {
		serviceLogger.INFO("Test Info", "Test Data")

		time.Sleep(sleepTime) // Allow time for log processing

		if len(serviceLogger.Messages(log.Pagination{Page: 1, Amount: 10})) == 0 {
			t.Error("Expected at least one log entry")
		}
	})

	t.Run("ConcurrentLogging", func(t *testing.T) {
		var wg sync.WaitGroup

		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				serviceLogger.DEBUG("Concurrent Log", idx)
			}(i)
		}

		wg.Wait()

		time.Sleep(sleepTime) // Allow time for log processing

		if len(serviceLogger.Messages(log.Pagination{Page: 1, Amount: 100})) < 100 {
			t.Error("Expected 100 log entries from concurrent logging")
		}
	})

	t.Run("LoggerOutput", func(t *testing.T) {
		wg := sync.WaitGroup{}
		go func() {
			wg.Add(1)
			serviceLogger.Output(func(l log.Log) error {
				if l.Service != "ServiceLogger" {
					t.Errorf("Received l from unexpected service: %v", l.Service)
					return errors.New("received l from unexpected service")
				}

				return nil
			})
			wg.Done()
		}()
		wg.Wait()
		serviceLogger.INFO("Test for Output Channel", "Test Data")
		time.Sleep(sleepTime)

	})

	t.Run("TotalSent", func(t *testing.T) {
		t.Log("TotalSent")
		totalSentBefore := logger.TotalSent()

		logger.DEBUG("Test Total Sent", "Test Data")

		time.Sleep(50 * time.Millisecond) // Allow time for log processing
		if logger.TotalSent() <= totalSentBefore {
			t.Error("Expected TotalSent to increase after logging")
		}
	})
}

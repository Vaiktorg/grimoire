package tests

import (
	"errors"
	"github.com/vaiktorg/grimoire/log"
	"github.com/vaiktorg/grimoire/store"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const (
	totalLogAmount = store.CurrentLen
	sleepTime      = 10 * time.Millisecond
)

func TestMain(m *testing.M) {
	m.Run()
}

func TestLoggingCaching(t *testing.T) {
	t.Cleanup(cleanup)

	logger := log.NewLogger(&log.Config{ServiceName: "TestLogger", CanOutput: true, Persist: true})
	defer logger.Close()

	servLogger := logger.NewServiceLogger(&log.Config{ServiceName: "TestLoggerService", CanOutput: true, Persist: true})

	t.Run("ServiceLogger", func(t *testing.T) {
		// Number of messages to test caching with.
		numMessages := totalLogAmount
		receivedMessages := uint64(0)

		wg := new(sync.WaitGroup)
		wg.Add(numMessages)

		go servLogger.Output(func(l log.Log) error {
			if l.Service == servLogger.ServiceName() {
				defer wg.Done()
				atomic.AddUint64(&receivedMessages, 1)
			} else if l.Service == "TestLogger" {
				return errors.New("service logger should not receive from main logger")
			}

			return nil
		})

		// Send multiple log messages.
		for i := 0; i < numMessages; i++ {
			go func(msgNum int) {
				// Use different logging levels for diversity.
				switch msgNum % 5 {
				case 0:
					servLogger.TRACE("test TRACE message")
				case 1:
					servLogger.DEBUG("test DEBUG message")
				case 2:
					servLogger.INFO("test INFO message")
				case 3:
					servLogger.WARN("test WARN message")
				case 4:
					servLogger.ERROR("test ERROR message")
				}
			}(i)
		}

		// Wait for all logging operations to complete.
		wg.Wait()

		// Retrieve the cached messages.
		messages := servLogger.Messages(log.Pagination{
			Page:   1,
			Amount: int(receivedMessages),
		})

		// Check if the messages have been cached.
		if int(receivedMessages) != numMessages {
			t.Errorf("Expected %d cached messages, found %d", numMessages, receivedMessages)
			return
		}

		// Optionally, verify the content of each message.
		for i, msg := range messages {
			expectedMsg := "test " + msg.Level + " message"
			if msg.Msg != expectedMsg {
				t.Errorf("Message %d does not match expected content: got '%s', want '%s'", i, msg.Msg, expectedMsg)
				return
			}
		}
	})

	t.Run("MainLogger", func(t *testing.T) {
		// Number of messages to test caching with.
		numMessages := totalLogAmount
		receivedMessages := uint64(0)

		wg := new(sync.WaitGroup)
		go logger.Output(func(l log.Log) error {
			defer wg.Done()
			atomic.AddUint64(&receivedMessages, 1)

			return nil
		})

		wg.Add(numMessages * 2)
		// Send multiple log messages.
		for i := 0; i < numMessages; i++ {
			func(msgNum int) {
				// Use different logging levels for diversity.
				switch msgNum % 5 {
				case 0:
					logger.TRACE("test TRACE message")
				case 1:
					logger.DEBUG("test DEBUG message")
				case 2:
					logger.INFO("test INFO message")
				case 3:
					logger.WARN("test WARN message")
				case 4:
					logger.ERROR("test ERROR message")
				}
			}(i)
		}

		// Wait for all logging operations to complete.
		wg.Wait()

		// Retrieve the cached messages.
		messages := logger.Messages(log.Pagination{
			Page:   1,
			Amount: numMessages,
		})

		// Check if the messages have been cached.
		if int(receivedMessages/2) != numMessages {
			t.Errorf("Expected %d cached messages, found %d", numMessages, receivedMessages/2)
			return
		}

		// Optionally, verify the content of each message.
		for i, msg := range messages {
			expectedMsg := "test " + msg.Level + " message"
			if msg.Msg != expectedMsg {
				t.Errorf("Message %d does not match expected content: got '%s', want '%s'", i, msg.Msg, expectedMsg)
				return
			}
		}
	})
}

func cleanup() {
	var logPaths []string
	_ = filepath.Walk(".", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		parts := regexp.MustCompile(`__([^_]+)__([^_]+)__`).FindStringSubmatch(filepath.Base(path))
		if len(parts) != 3 {
			return nil
		}

		//parts[0] : BasePath
		//parts[1] : ServiceName
		//parts[2] : RunID

		if filepath.Ext(info.Name()) == ".log" {
			logPaths = append(logPaths, path)
		}

		return nil
	})
	for _, path := range logPaths {
		err := os.RemoveAll(path)
		if err != nil {
			panic(err)
		}
	}
}

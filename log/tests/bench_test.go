package tests

import (
	"github.com/vaiktorg/grimoire/log"
	"testing"
)

func BenchmarkLogger(b *testing.B) {
	b.Cleanup(cleanup)

	b.Run("BenchmarkLogCreation", func(b *testing.B) {
		logger := log.NewLogger(&log.Config{ServiceName: "MainService", Persist: true})

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.INFO("Test log message")
		}

		logger.Close()
	})
	b.Run("BenchmarkChannelCommunication", func(b *testing.B) {
		mainLogger := log.NewLogger(&log.Config{ServiceName: "MainService", Persist: true})
		defer mainLogger.Close()

		serviceLogger := mainLogger.NewServiceLogger(&log.Config{ServiceName: "ServiceLogger", Persist: true})

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			serviceLogger.INFO("Test log message")
		}
	})
	b.Run("BenchmarkConcurrentLogging", func(b *testing.B) {
		logger := log.NewLogger(&log.Config{ServiceName: "MainService", Persist: true})
		defer logger.Close()

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.INFO("Test log message")
			}
		})
	})

	b.Run("BenchmarkConcurrentServiceLogging", func(b *testing.B) {
		mainLogger := log.NewLogger(&log.Config{ServiceName: "MainService", Persist: true})
		defer mainLogger.Close()

		serviceLogger := mainLogger.NewServiceLogger(&log.Config{ServiceName: "ServiceLogger", Persist: true})

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				mainLogger.INFO("Test log message")
			}
		})
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				serviceLogger.INFO("Test log message")
			}
		})
	})
}

package log

import (
	"fmt"
	"github.com/vaiktorg/grimoire/store"
	"github.com/vaiktorg/grimoire/uid"
	"os"
	"sync/atomic"
	"time"
)

type SimLogger struct {
	Service   string
	Cache     *store.Cache[Log] // Assuming a simpler cache implementation.
	totalSent *uint64
	Level     Level
	outChan   chan Log
	services  *store.Repo[string, ILogger]
}

func NewSimLogger(service string) *SimLogger {
	totalSent := uint64(0)
	return &SimLogger{
		Service:   service,
		Cache:     store.NewCache[Log](service), // Assuming a simpler cache implementation.
		outChan:   make(chan Log, 100),          // Adjust buffer size as needed.
		totalSent: &totalSent,
	}
}

func (l *SimLogger) NewServiceLogger(config *Config) ILogger {
	return &SimLogger{
		Service:   config.ServiceName,
		Cache:     l.Cache,
		totalSent: l.totalSent,
		Level:     l.Level,
		outChan:   l.outChan,
		services:  store.NewRepo[string, ILogger](),
	}
}

func (l *SimLogger) ServiceName() string {
	return l.Service
}

func (l *SimLogger) Messages(pagination Pagination) []Log {
	start := int64((pagination.Page - 1) * pagination.Amount)
	end := start + int64(pagination.Amount)
	total := int64(l.Cache.FlushLen()*store.CurrentLen) + l.Cache.Len()

	if end > total {
		end = total
	}

	if start < 0 || start >= total {
		return nil
	}

	return l.Cache.ReadAll(l.Service)
}

func (l *SimLogger) Services() map[string]ILogger {
	return l.services.All()
}

func (l *SimLogger) Output(f func(log Log) error) {
	for log := range l.outChan {
		if err := f(log); err != nil {
			_ = l.ERROR(err.Error())
			return
		}
	}
}

func (l *SimLogger) Println(in ...any) {
	fmt.Println(in)
}

func (l *SimLogger) Printf(str string, data ...any) {
	fmt.Printf(str, data)
}

func (l *SimLogger) Log(level Level, msg string, data ...any) {
	if level < l.Level {
		return
	}

	log := Log{
		ID:        atomic.AddUint64(l.totalSent, 1),
		SourceId:  []byte(uid.NewUID(8)), // Assuming a simpler UID generation function.
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     level.String(),
		Service:   l.Service,
		Msg:       msg,
		Data:      data,
	}

	_, _ = os.Stdout.WriteString(log.String() + "\n")

	l.Cache.Write(log) // Assuming a simpler cache implementation.
	l.outChan <- log
}

func (l *SimLogger) TRACE(info string, obj ...any) { l.Log(LevelTrace, info, obj...) }
func (l *SimLogger) INFO(info string, obj ...any)  { l.Log(LevelInfo, info, obj...) }
func (l *SimLogger) DEBUG(info string, obj ...any) { l.Log(LevelDebug, info, obj...) }
func (l *SimLogger) WARN(info string, obj ...any)  { l.Log(LevelWarn, info, obj...) }
func (l *SimLogger) ERROR(err string, obj ...any) string {
	l.Log(LevelError, err, obj...)
	return err
}
func (l *SimLogger) FATAL(info string) { l.Log(LevelFatal, info) }

func (l *SimLogger) TotalSent() uint64 {
	return atomic.LoadUint64(l.totalSent)
}
func (l *SimLogger) BatchLogs(logs ...Log) {
	for _, log := range logs {
		l.Log(LevelFromString(log.Level), log.Msg, log.Data)
	}
}
func (l *SimLogger) Close() {
	l.services.Iterate(func(s string, logger ILogger) {
		logger.Close()
		l.services.Delete(s)
	})
	close(l.outChan)
}

package log

import (
	"fmt"
	"github.com/vaiktorg/grimoire/store"
	"github.com/vaiktorg/grimoire/uid"
	"sync/atomic"
	"time"
)

type SimLogger struct {
	Service   string
	Cache     *store.Cache[Log] // Assuming a simpler cache implementation.
	totalSent uint64
	Level     Level
	outChan   chan Log
	services  store.Repo[string, ILogger]
}

func NewSimLogger(service string, level Level) *SimLogger {
	return &SimLogger{
		Service:   service,
		Cache:     store.NewCache[Log](service), // Assuming a simpler cache implementation.
		Level:     level,
		outChan:   make(chan Log, 100), // Adjust buffer size as needed.
		totalSent: 0,
	}
}

func (l *SimLogger) NewServiceLogger(config Config) ILogger {
	return &SimLogger{
		Service:   config.ServiceName,
		Cache:     l.Cache,
		totalSent: l.totalSent,
		Level:     l.Level,
		outChan:   l.outChan,
	}
}

func (l *SimLogger) ServiceName() string {
	return l.Service
}

func (l *SimLogger) Messages(pagination Pagination) []Log {
	start := (pagination.Page - 1) * pagination.Amount
	end := start + pagination.Amount
	total := l.Cache.FlushLen()*store.DefaultLen + l.Cache.Len()

	if end > total {
		end = total
	}

	if start < 0 || start >= total {
		return nil
	}

	return l.Cache.ReadAll(l.Service)
}

func (l *SimLogger) Services() map[string]ILogger {
	//TODO implement me
	return l.services.All()
}

func (l *SimLogger) Output(f func(log Log)) {
	for log := range l.outChan {
		f(log)
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
		ID:        atomic.AddUint64(&l.totalSent, 1),
		SourceId:  uid.NewUID(8).String(), // Assuming a simpler UID generation function.
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     level.String(),
		Service:   l.Service,
		Msg:       msg,
		Data:      data,
	}

	l.Cache.Write(log) // Assuming a simpler cache implementation.
	l.outChan <- log
}

func (l *SimLogger) TRACE(info string, obj ...any) { l.Log(LevelTrace, info, obj...) }
func (l *SimLogger) INFO(info string, obj ...any)  { l.Log(LevelInfo, info, obj...) }
func (l *SimLogger) DEBUG(info string, obj ...any) { l.Log(LevelDebug, info, obj...) }
func (l *SimLogger) WARN(info string, obj ...any)  { l.Log(LevelWarn, info, obj...) }
func (l *SimLogger) ERROR(info string, obj ...any) { l.Log(LevelError, info, obj...) }
func (l *SimLogger) FATAL(info string)             { l.Log(LevelFatal, info) }

func (l *SimLogger) TotalSent() uint64 {
	return atomic.LoadUint64(&l.totalSent)
}

func (l *SimLogger) Close() {
	close(l.outChan)
}

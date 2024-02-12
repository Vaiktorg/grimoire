package log

import (
	"fmt"
	"github.com/vaiktorg/grimoire/store"
	"github.com/vaiktorg/grimoire/uid"
	"os"
	"sync/atomic"
	"time"
)

type StdOutLogger struct {
	Service   string
	totalSent uint64
	outChan   chan Log
	services  *store.Repo[string, ILogger]
	logs      []Log
}

func NewStdOutLogger(service string) ILogger {
	return &StdOutLogger{
		Service:   service,
		outChan:   make(chan Log, 100), // Adjust buffer size as needed.
		totalSent: 0,
		services:  store.NewRepo[string, ILogger](),
	}
}

func (l *StdOutLogger) Log(level Level, msg string, data ...any) {
	log := Log{
		ID:        atomic.AddUint64(&l.totalSent, 1),
		SourceId:  []byte(uid.NewUID(8)), // Assuming a simpler UID generation function.
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     level.String(),
		Service:   l.Service,
		Msg:       msg,
		Data:      data,
	}

	_, _ = os.Stdout.WriteString(log.String() + "\n")
}

func (l *StdOutLogger) TRACE(info string, obj ...any) { l.Log(LevelTrace, info, obj...) }
func (l *StdOutLogger) INFO(info string, obj ...any)  { l.Log(LevelInfo, info, obj...) }
func (l *StdOutLogger) DEBUG(info string, obj ...any) { l.Log(LevelDebug, info, obj...) }
func (l *StdOutLogger) WARN(info string, obj ...any)  { l.Log(LevelWarn, info, obj...) }
func (l *StdOutLogger) ERROR(err string, obj ...any) string {
	l.Log(LevelError, err, obj...)
	return err
}
func (l *StdOutLogger) FATAL(info string) { l.Log(LevelFatal, info) }

// ====================================================================================================

func (l *StdOutLogger) ServiceName() string       { return l.Service }
func (l *StdOutLogger) Messages(Pagination) []Log { return l.logs }
func (l *StdOutLogger) NewServiceLogger(config *Config) ILogger {
	l.services.Add(config.ServiceName, &StdOutLogger{
		Service:   l.Service,
		totalSent: l.totalSent,
		outChan:   l.outChan,
		services:  store.NewRepo[string, ILogger](),
		logs:      l.logs,
	})

	return l.services.Get(config.ServiceName)
}
func (l *StdOutLogger) Services() map[string]ILogger   { return l.services.All() }
func (l *StdOutLogger) Println(in ...any)              { fmt.Println(in...) }
func (l *StdOutLogger) Printf(str string, data ...any) { fmt.Printf(str, data...) }
func (l *StdOutLogger) TotalSent() uint64              { return l.totalSent }
func (l *StdOutLogger) BatchLogs(...Log)               { return }
func (l *StdOutLogger) Output(func(log Log) error)     { return }
func (l *StdOutLogger) Close()                         { return }

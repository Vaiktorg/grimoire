package log

import (
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
	services  store.Repo[string, ILogger]
	logs      []Log
}

func NewStdOutLogger(service string) ISimLogger {
	return &StdOutLogger{
		Service:   service,
		outChan:   make(chan Log, 100), // Adjust buffer size as needed.
		totalSent: 0,
	}
}

func (l *StdOutLogger) Log(level Level, msg string, data ...any) {
	log := Log{
		ID:        atomic.AddUint64(&l.totalSent, 1),
		SourceId:  uid.NewUID(8).String(), // Assuming a simpler UID generation function.
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
func (l *StdOutLogger) ERROR(info string, obj ...any) { l.Log(LevelError, info, obj...) }
func (l *StdOutLogger) FATAL(info string)             { l.Log(LevelFatal, info) }

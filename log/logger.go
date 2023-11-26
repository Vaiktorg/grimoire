package log

import (
	"errors"
	"fmt"
	"github.com/vaiktorg/grimoire/store"
	"github.com/vaiktorg/grimoire/uid"
	"os"
	"runtime/debug"
	_ "runtime/pprof"
	"sync"
	"sync/atomic"
	"time"
)

type Log struct {
	ID        uint64 `json:"id"`             // Incremental ID iota
	SourceId  string `json:"sid"`            // ID of where it comes from.
	Service   string `json:"serv"`           // Service appName for isolated logging
	Level     string `json:"lvl"`            //Log Severity Level
	Msg       string `json:"msg"`            // Message / Description
	Data      any    `json:"data,omitempty"` // Data if any for inspecting
	Timestamp string `json:"time"`           // When did we get the log
}

func (l *Log) String() string {
	dataFmt := "[%d] %s [ %s ] %s ==> %s %+v\n"
	msgFmt := "[%d] %s [ %s ] %s ==> %s\n"

	if d, ok := l.Data.([]interface{}); ok && d != nil {
		return fmt.Sprintf(dataFmt, l.ID, l.Timestamp, l.Service, l.Level, l.Msg, l.Data)
	}
	return fmt.Sprintf(msgFmt, l.ID, l.Timestamp, l.Service, l.Level, l.Msg)
}

type Logger struct {
	wg sync.WaitGroup

	closeChan chan struct{}
	runId     string

	logLevels Level
	Cache     *store.Cache[Log]
	size      uint
	totalSent *uint64

	canOutput atomic.Bool
	canPrint  bool
	Service   string

	services *store.Repo[ILogger]
	inChan   chan Log

	outChan  chan Log
	pOutChan chan Log
}

type Config struct {
	CanPrint    bool
	CanOutput   bool
	ServiceName string
}

func NewLogger(config Config) ILogger {
	runId := uid.NewUID(8).String()
	totalSent := uint64(0)

	l := &Logger{
		runId:     runId,
		totalSent: &totalSent,

		logLevels: LevelTrace | LevelInfo | LevelDebug | LevelError,

		inChan:    make(chan Log, store.DefaultLen),
		outChan:   make(chan Log, store.DefaultLen),
		pOutChan:  nil,
		closeChan: make(chan struct{}),

		Cache:    store.NewIDCache[Log](config.ServiceName, runId),
		services: store.NewRepo[ILogger](),

		Service:  config.ServiceName,
		canPrint: config.CanPrint,
	}

	l.canOutput.Store(config.CanOutput)

	go l.inputLogs()

	return l
}

// ==================================================

func (l *Logger) NewServiceLogger(config Config) IServiceLogger {
	if l.services.Has(config.ServiceName) {
		l.ERROR(errors.New("service logger " + config.ServiceName + " already existed"))
		return nil
	}

	service := &Logger{
		runId:     l.runId,
		logLevels: l.logLevels,
		totalSent: l.totalSent,

		inChan:   make(chan Log, store.DefaultLen), // Log Input -> Proc
		outChan:  make(chan Log, store.DefaultLen), // Proc 	  -> Output(func(Log))
		pOutChan: l.outChan,

		closeChan: make(chan struct{}),

		Cache:    store.NewIDCache[Log](config.ServiceName, l.runId),
		services: store.NewRepo[ILogger](),

		Service:  config.ServiceName,
		canPrint: config.CanPrint,
	}

	service.canOutput.Store(config.CanOutput)

	go service.inputLogs()

	l.services.Add(config.ServiceName, service)

	return service
}

func (l *Logger) ServiceName() string {
	return l.Service
}

func (l *Logger) Services() map[string]ILogger {
	return l.services.All()
}

// ==================================================

// TRACE Used for debugging, should not exist after production
func (l *Logger) TRACE(info string, obj ...any) {
	l.newMsg(info, obj, LevelTrace)
}

// INFO Used to tell users of things going on in their process steps
func (l *Logger) INFO(info string, obj ...any) {
	l.newMsg(info, obj, LevelInfo)
}

// DEBUG Used to communicate processes to other developers
func (l *Logger) DEBUG(procStep string, obj ...any) {
	l.newMsg(procStep, obj, LevelDebug)
}

// WARN  Possible breaking scenarios: If you do this, this could happen, keep it in mind, etc.
func (l *Logger) WARN(warn string, obj ...any) {
	l.newMsg(warn, obj, LevelWarn)
}

// ERROR Something broke, print out the error message and possible entries
func (l *Logger) ERROR(errMsg error, obj ...any) {
	l.newMsg(errMsg.Error(), obj, LevelError)
}

// FATAL This should not have happened, very system critical, total breakage risk
func (l *Logger) FATAL(breakage string) {
	l.newMsg(breakage, debug.Stack(), LevelFatal)
}

func (l *Logger) Println(in ...any) {
	_, _ = fmt.Fprintln(os.Stdout, in...)
}

func (l *Logger) Printf(str string, data ...any) {
	_, _ = fmt.Fprintf(os.Stdout, str, data...)
}

// ==================================================

type Pagination struct {
	Page   int
	Amount int
}

func (l *Logger) Messages(p Pagination) []Log {
	start := (p.Page - 1) * p.Amount
	end := start + p.Amount
	total := l.Cache.FlushLen()*store.DefaultLen + l.Cache.Len()

	if end > total {
		end = total
	}

	if start < 0 || start >= total {
		return nil
	}

	return l.Cache.ReadAll(l.Service)[start:end]
}

func (l *Logger) BatchLogs(logs ...Log) {
	for _, log := range logs {
		l.inChan <- log
	}
}
func (l *Logger) Output(handler func(log Log)) {
	for log := range l.outChan {
		if handler != nil {
			handler(log)
		}
	}
}
func (l *Logger) TotalSent() uint64 {
	return atomic.LoadUint64(l.totalSent)
}

func (l *Logger) Close() {
	l.services.Iterate(func(servName string, logger ILogger) {
		logger.Close()
		l.services.Delete(servName)
	})
	l.canOutput.Swap(false)

	close(l.inChan)
	l.wg.Wait()
	close(l.outChan)

	l.Cache.Flush()
}

// ==================================================

func (l *Logger) newMsg(msg string, data any, level Level) {
	if !l.HasLevel(level) {
		return // Skip if logger is closeChan or level is not enabled
	}

	l.wg.Add(1)
	l.inChan <- Log{
		ID:        atomic.AddUint64(l.totalSent, 1),
		SourceId:  uid.NewUID(8).String(),
		Timestamp: time.Now().Format("01-02-2006_03-04-05"),
		Level:     level.String(),
		Service:   l.Service,
		Msg:       msg,
		Data:      data,
	}
}

func (l *Logger) inputLogs() {
	for in := range l.inChan {
		l.Cache.Write(in)

		if l.Cache.IsFull() {
			go l.Cache.Flush()
		}

		if l.canPrint {
			_, _ = os.Stdout.WriteString(in.String())
		}

		l.internalOutputLogs(in) // Send it to output proc
	}
}

func (l *Logger) internalOutputLogs(out Log) {
	defer l.wg.Done()
	if l.canOutput.Load() {
		l.outChan <- out

		if l.pOutChan != nil {
			l.pOutChan <- out
		}
	}
}

// ==================================================

func (l *Logger) HasLevel(flag Level) bool {
	return l.logLevels.Has(flag)
}
func (l *Logger) AddLevel(flag Level) {
	l.logLevels.Set(flag)
}
func (l *Logger) ClearLevel(flag Level) {
	l.logLevels.Clear(flag)
}
func (l *Logger) ToggleLevel(flag Level) {
	l.logLevels.Toggle(flag)
}

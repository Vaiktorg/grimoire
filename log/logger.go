package log

import (
	"fmt"
	"github.com/vaiktorg/grimoire/names"
	"github.com/vaiktorg/grimoire/store"
	"github.com/vaiktorg/grimoire/uid"
	"os"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
)

type Log struct {
	ID        uint64 `json:"id"`             // Incremental ID iota
	SourceId  []byte `json:"sid"`            // ID of where it comes from.
	Service   string `json:"serv"`           // Service appName for isolated logging
	Level     string `json:"lvl"`            //Log Severity Level
	Msg       string `json:"msg"`            // Message / Description
	Data      any    `json:"data,omitempty"` // Data if any for inspecting
	Timestamp string `json:"time"`           // When did we get the log
}

func (l *Log) String() string {
	if d, ok := l.Data.([]interface{}); !ok && d != nil {
		return fmt.Sprintf("[%d] %s %s [ %s ] ==> %s %v", l.ID, l.Timestamp, l.Level, l.Service, l.Msg, l.Data)
	}
	return fmt.Sprintf("[%d] %s %s [ %s ] ==> %s", l.ID, l.Timestamp, l.Level, l.Service, l.Msg)
}

type Logger struct {
	wg sync.WaitGroup

	Service   string
	closeChan chan struct{}

	runId uid.UID

	Cache *store.Cache[Log]

	totalSent *uint64

	persist   atomic.Bool
	canOutput atomic.Bool
	canPrint  atomic.Bool

	services *store.Repo[string, ILogger]
	inChan   chan Log

	outChan  chan Log
	pOutChan chan Log
}

type Config struct {
	CanPrint    bool
	CanOutput   bool
	Persist     bool
	ServiceName string
}

func NewLogger(config *Config) ILogger {
	runId := uid.NewUID(8)
	totalSent := uint64(0)

	l := &Logger{
		runId:     runId,
		totalSent: &totalSent,

		inChan:    make(chan Log, store.CurrentLen),
		outChan:   make(chan Log, store.CurrentLen),
		pOutChan:  nil,
		closeChan: make(chan struct{}),

		Service:  config.ServiceName + "_" + string(runId),
		Cache:    store.NewIDCache[Log](config.ServiceName, []byte(runId)),
		services: store.NewRepo[string, ILogger](),
	}

	l.canOutput.Store(config.CanOutput)
	l.canPrint.Store(config.CanPrint)
	l.persist.Store(config.Persist)

	go l.inputLogs()

	return l
}

// ==================================================

func (l *Logger) NewServiceLogger(config *Config) ILogger {
	if l.services.Has(config.ServiceName) {
		_ = l.ERROR("service logger " + config.ServiceName + " already existed")
		return nil
	}

	service := &Logger{
		runId:     l.runId,
		totalSent: l.totalSent,

		inChan:    make(chan Log, store.CurrentLen), // Log Input -> Proc
		outChan:   make(chan Log, store.CurrentLen), // Proc 	  -> Output(func(Log))
		pOutChan:  l.outChan,
		closeChan: make(chan struct{}),

		Service:  names.NewLastName(config.ServiceName),
		Cache:    store.NewIDCache[Log](config.ServiceName, []byte(l.runId)),
		services: store.NewRepo[string, ILogger](),
	}

	service.canOutput.Store(config.CanOutput)
	service.canPrint.Store(config.CanPrint)
	service.persist.Store(config.Persist)

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
	l.newMsg(info, LevelTrace, obj)
}

// INFO Used to tell users of things going on in their process steps
func (l *Logger) INFO(info string, obj ...any) {
	l.newMsg(info, LevelInfo, obj)
}

// DEBUG Used to communicate processes to other developers
func (l *Logger) DEBUG(procStep string, obj ...any) {
	l.newMsg(procStep, LevelDebug, obj)
}

// WARN  Possible breaking scenarios: If you do this, this could happen, keep it in mind, etc.
func (l *Logger) WARN(warn string, obj ...any) {
	l.newMsg(warn, LevelWarn, obj)
}

// ERROR Something broke, print out the error message and possible entries
func (l *Logger) ERROR(errMsg string, obj ...any) string {
	l.newMsg(errMsg, LevelError, obj)
	return errMsg
}

// FATAL This should not have happened, very system critical, total breakage risk
func (l *Logger) FATAL(breakage string) {
	l.newMsg(breakage+"\n"+string(debug.Stack()), LevelFatal)
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
	start := int64((p.Page - 1) * p.Amount)
	end := start + int64(p.Amount)
	total := int64(l.Cache.FlushLen()*store.CurrentLen) + l.Cache.Len()

	if end > total {
		end = total
	}

	if start < 0 || start >= total {
		return nil
	}

	msgs := l.Cache.ReadAll(l.Service)[start:end]

	return msgs
}

func (l *Logger) BatchLogs(logs ...Log) {
	for _, log := range logs {
		l.inChan <- log
	}
}
func (l *Logger) Output(handler func(log Log) error) {
	for log := range l.outChan {
		if handler != nil {
			if err := handler(log); err != nil {
				_ = l.ERROR(err.Error())
				return
			}
		}
	}
}
func (l *Logger) TotalSent() uint64 {
	return atomic.LoadUint64(l.totalSent)
}

func (l *Logger) Close() {
	l.services.Iterate(func(servName string, logger ILogger) {
		logger.Close()
		defer l.services.Delete(servName)
	})

	close(l.inChan)
	l.canOutput.Swap(false)

	l.wg.Wait()
	close(l.outChan)

	l.Cache.Flush()
}

// ==================================================

func (l *Logger) newMsg(msg string, level Level, data ...any) {
	l.wg.Add(1)
	l.inChan <- Log{
		ID:        atomic.AddUint64(l.totalSent, 1),
		SourceId:  []byte(uid.NewUID(8)),
		Timestamp: time.Now().Format("01-02-2006_03-04-05"),
		Level:     level.String(),
		Service:   l.Service,
		Msg:       msg,
		Data:      data,
	}
}
func (l *Logger) inputLogs() {
	for in := range l.inChan {
		if l.persist.Load() {
			// persist to disk
			if l.Cache.IsFull() {
				go l.Cache.Flush()
			}

			// Write to memory
			l.Cache.Write(in)
		}

		// Print to os.Stdout
		if l.canPrint.Load() {
			_, _ = os.Stdout.WriteString(in.String() + "\n")
		}

		//
		if l.canOutput.Load() {
			l.internalOutputLogs(in) // Send it to output proc
		}

		l.wg.Done()
	}
}
func (l *Logger) internalOutputLogs(out Log) {
	l.outChan <- out

	if l.pOutChan != nil {
		l.pOutChan <- out
	}
}

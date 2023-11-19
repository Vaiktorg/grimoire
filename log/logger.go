package log

import (
	"errors"
	"fmt"
	"github.com/vaiktorg/grimoire/uid"
	"os"
	"runtime/debug"
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
	mu  sync.RWMutex
	iwg sync.WaitGroup
	owg sync.WaitGroup

	closed atomic.Bool

	runId string

	LogLevels Level
	Cache     *Cache[Log]
	size      uint
	totalSent *uint64

	canPrint bool

	Service  string
	services *Repo[ILogger]

	inChan  chan Log
	outChan chan Log

	canOutput bool
	outInt    chan Log
}

type Config struct {
	CanPrint    bool
	CanOutput   bool
	ServiceName string
}

func NewLogger(config Config) ILogger {
	runId := UID(8)
	totalSent := uint64(0)

	l := &Logger{
		runId:     runId,
		totalSent: &totalSent,

		LogLevels: LevelTrace | LevelInfo | LevelDebug | LevelError,
		inChan:    make(chan Log, DefaultLogLen),

		outInt:  make(chan Log, DefaultLogLen),
		outChan: make(chan Log, DefaultLogLen),

		Cache: NewCache[Log](config.ServiceName, runId),

		services: NewRepo[ILogger](),

		canOutput: config.CanOutput,
		Service:   config.ServiceName,
		canPrint:  config.CanPrint,
	}

	l.iwg.Add(1)
	go l.writeLogInput()

	return l
}

// ==================================================

func (l *Logger) ServiceName() string {
	return l.Service
}

func (l *Logger) NewServiceLogger(config Config) ILogger {
	if l.services.Has(config.ServiceName) {
		l.ERROR(errors.New("could not create services logger " + config.ServiceName))
		return nil
	}

	service := &Logger{
		runId:     l.runId,
		totalSent: l.totalSent,

		LogLevels: l.LogLevels,
		inChan:    make(chan Log, DefaultLogLen), // Log Input -> Proc

		outChan: make(chan Log, DefaultLogLen), // Proc 	  -> Output(func(Log))
		outInt:  l.outChan,                     // Proc 	  -> Parent Logger

		Cache: NewCache[Log](config.ServiceName, l.runId),

		services: NewRepo[ILogger](),

		canOutput: config.CanOutput,
		Service:   config.ServiceName,
		canPrint:  config.CanPrint,
	}

	l.services.Add(config.ServiceName, service)

	l.iwg.Add(1)
	go service.writeLogInput()

	return service
}
func (l *Logger) Services() *Repo[ILogger] {
	return l.services
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
	l.mu.Lock()
	defer l.mu.Unlock()

	start := (p.Page - 1) * p.Amount
	end := start + p.Amount
	total := int(l.Cache.flushTotal)*DefaultLogLen + l.Cache.Len()

	if end > total {
		end = total
	}

	if start < 0 || start >= total {
		return nil
	}

	return l.Cache.Collection(l.Service)[start:end]
}

func (l *Logger) BatchLogs(logs ...Log) {
	for _, log := range logs {
		l.inChan <- log
	}
}
func (l *Logger) Output(handler func(log Log) error) {
	for log := range l.outChan {
		if err := handler(log); err != nil {
			l.ERROR(err)
			break
		}
	}
}
func (l *Logger) TotalSent() uint64 {
	l.mu.Lock()
	defer l.mu.Unlock()

	return *l.totalSent
}
func (l *Logger) Len() int {
	return l.Cache.Len()
}

func (l *Logger) Close() {
	if l.closed.CompareAndSwap(false, true) {
		return
	}

	l.services.Iterate(func(logger ILogger) {
		logger.Close()
		l.Services().Delete(logger.ServiceName())
	})

	close(l.inChan)
	if l.outInt != nil {
		close(l.outInt)
	}

	l.iwg.Wait()

	close(l.outChan)

	go l.must(l.Cache.Flush())
}

// ==================================================

func (l *Logger) newMsg(msg string, data any, level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.HasLevel(level) {
		return
	}

	l.iwg.Add(1)
	l.inChan <- Log{
		ID:        atomic.AddUint64(l.totalSent, 1),
		Timestamp: time.Now().Format("01-02-2006_03-04-05"),
		Level:     level.String(),
		Service:   l.Service,
		Msg:       msg,
		SourceId:  uid.NewUID(8).String(),
		Data:      data,
	}
}
func (l *Logger) writeLogInput() {
	for log := range l.inChan {
		l.Cache.Write(log)

		if l.Cache.Len() >= DefaultLogLen {
			go l.must(l.Cache.Flush())
		}
		if l.canPrint {
			_, _ = os.Stdout.WriteString(log.String())
		}

		go l.forwardLog(log)

		l.iwg.Done()
	}
}
func (l *Logger) forwardLog(log Log) {
	if l.outInt != nil {
		l.outInt <- log
	}

	if l.canOutput {
		l.outChan <- log
	}

}

// ==================================================

func (l *Logger) HasLevel(flag Level) bool {
	return l.LogLevels.Has(flag)
}
func (l *Logger) AddLevel(flag Level) {
	l.LogLevels.Set(flag)
}
func (l *Logger) ClearLevel(flag Level) {
	l.LogLevels.Clear(flag)
}
func (l *Logger) ToggleLevel(flag Level) {
	l.LogLevels.Toggle(flag)

}

// ==================================================
func (l *Logger) must(err error) {
	defer func() {
		if e := recover(); e != nil {
			_, _ = fmt.Fprint(os.Stdout, e.(error))
		}
	}()

	if err != nil {
		l.ERROR(err)
		panic(err)
	}
}

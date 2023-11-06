package log

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

type Log struct {
	ID        uint64 `json:"order"` // Incremental ID
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Service   string `json:"service"`
	Msg       string `json:"msg"`
	SourceId  string `json:"sourceId"`
	Data      any    `json:"data,omitempty"`
}

func (l *Log) String() string {
	if _, ok := l.Data.([]any); ok {
		return fmt.Sprintf("%s\t[ %s ] %s ==> %s\n %+v \n", l.Timestamp, l.Level, l.Service, l.Msg, l.Data)
	}
	return fmt.Sprintf("%s\t[ %s ] %s ==> %s\n", l.Timestamp, l.Level, l.Service, l.Msg)
}

type Logger struct {
	mu    sync.Mutex
	cache Cache

	Service   string
	services  []string
	LogLevels Level
	totalLogs uint64
	size      int

	inChan chan Log
	Output chan Log
}

const (
	DefaultLogSize = 20 * MB
	MaxStoreSize   = DefaultLogSize
)

func NewLogger() (*Logger, error) {
	l := &Logger{
		totalLogs: 0,
		Service:   "global",
		inChan:    make(chan Log, 10),
		Output:    make(chan Log, 10),
		LogLevels: LevelTrace | LevelInfo | LevelDebug | LevelError,
	}

	go l.monitorPersistence()

	return l, nil
}

func (l *Logger) NewService(servName string) *Logger {
	l.services = append(l.services, servName)
	return &Logger{
		inChan:    l.inChan,
		Output:    l.Output,
		Service:   servName,
		services:  l.services,
		totalLogs: l.totalLogs,
		LogLevels: l.LogLevels,
	}
}

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
	for _, data := range in {
		println(data)
	}
}

func (l *Logger) Printf(str string, data ...any) {
	fmt.Printf(str, data)
}

// ==================================================

func (l *Logger) Messages() []Log {
	return l.cache
}
func (l *Logger) Close() {
	close(l.inChan)
	close(l.Output)
	l.toFile(l.cache)
}

func (l *Logger) newMsg(msg string, data any, level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.HasLevel(level) {
		logMsg := Log{
			ID:        atomic.AddUint64(&l.totalLogs, 1),
			Timestamp: time.Now().Format("01-02-2006_03-04-05"),
			Level:     level.String(),
			Msg:       msg,
			Data:      data,
		}

		l.cache.Write(logMsg)

		l.inChan <- logMsg

		_, err := os.Stdout.Write([]byte(logMsg.String()))
		if err != nil {
			panic(err)
		}
	}
}

// ==================================================

func (l *Logger) monitorPersistence() {
	go func() {
		for msg := range l.inChan {
			l.size += int(unsafe.Sizeof(l.cache) + unsafe.Sizeof(""))
			if l.size >= MaxStoreSize.Val() {
				l.dumpLog()
			}

			l.Output <- msg
		}
	}()
}

func (l *Logger) dumpLog() {
	l.toFile(l.cache)
	l.cache = Cache{}
}

func (l *Logger) toFile(msg []Log) {
	filename := time.Now().Format("2006-01-02__15-04") + ".log"
	filename = l.Service + "__" + filename

	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}

	enc := json.NewEncoder(file)

	l.must(enc.Encode(msg))
	l.must(file.Close())
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
		stack := debug.Stack()
		l.FATAL(string(stack))
		panic(err)
	}
}

func (l *Logger) canWriteToChannel(ch chan Log) bool {
	select {
	case <-ch:
		return true
	default:
		return false
	}
}

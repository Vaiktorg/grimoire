package log

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime/debug"
	"sync/atomic"
	"time"
	"unsafe"
)

type Log struct {
	ID        int    `json:"order"` // Incremental ID
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Service   string `json:"service"`
	Msg       string `json:"msg"`
	SourceId  string `json:"sourceId"`
	Data      any    `json:"data,omitempty"`
}

func (l *Log) String() string {
	if l.Data.([]any) != nil {
		return fmt.Sprintf("%s\t[ %s ] %s ==> %s\n %+v \n", l.Timestamp, l.Level, l.Service, l.Msg, l.Data)
	}
	return fmt.Sprintf("%s\t[ %s ] %s ==> %s\n", l.Timestamp, l.Level, l.Service, l.Msg)
}

type Logger struct {
	//store  strings.Builder
	store     LogCache
	LogLevels Levels
	enc       json.Encoder

	inChan chan Log

	//network//
	Handler   *WebSocketViewer
	totalLogs uint64
}

const (
	DefaultLogSize = 5 * MB
	MaxStoreSize   = DefaultLogSize
)

func NewLogger(appId string) (*Logger, error) {
	viewer, err := NewWebSocketViewer(appId)
	if err != nil {
		return nil, err
	}

	l := &Logger{
		store:     LogCache{},
		LogLevels: LevelTrace | LevelInfo | LevelDebug | LevelError,
		inChan:    make(chan Log),

		//network//
		Handler: viewer,
	}

	go l.handlePersistence()

	return l, nil
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
func (l *Logger) FATAL(breakage string, obj ...any) {
	l.newMsg(breakage, obj, LevelFatal)
}

// INBOUND Generic Log msg injection
func (l *Logger) INBOUND(msg Log) {
	if l.HasLevel(l.LogLevels.Level(msg.Level)) {
		l.inChan <- msg
	}
}

// ==================================================

func (l *Logger) Messages() []Log {
	return l.store
}
func (l *Logger) Close() {
	close(l.inChan)
	defer l.toFile(l.store)

	fmt.Println("Logger Closed!")
}

func (l *Logger) newMsg(msg string, data any, level Levels) {
	if l.HasLevel(level) {
		l.inChan <- Log{
			Timestamp: time.Now().Format("20060102150405"),
			Level:     level.String(),
			Msg:       msg,
			Data:      data,
			ID:        int(atomic.AddUint64(&l.totalLogs, 1)),
		}
	}
}

// ==================================================

func (l *Logger) handlePersistence() {
	go func() {
		for msg := range l.inChan {
			mem := int(unsafe.Sizeof(l.store) + unsafe.Sizeof(""))
			if mem >= MaxStoreSize.Val() {
				l.dumpLog()
			}

			l.store.Write(msg)
			os.Stdout.WriteString(msg.String())
		}
	}()
}

func (l *Logger) dumpLog() {
	l.toFile(l.store)
	l.store = LogCache{}
}

func (l *Logger) toFile(msg []Log) {
	filename := time.Now().Format("20060102150405") + ".log"

	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}

	enc := json.NewEncoder(file)
	enc.SetIndent("", "	")
	err = enc.Encode(msg)
	l.must(err)

	err = file.Close()
	l.must(err)
}

// ==================================================

func (l *Logger) HasLevel(flag Levels) bool {
	return l.LogLevels.Has(flag)
}
func (l *Logger) AddLevel(flag Levels) {
	l.LogLevels.Set(flag)
}
func (l *Logger) ClearLevel(flag Levels) {
	l.LogLevels.Clear(flag)
}
func (l *Logger) ToggleLevel(flag Levels) {
	l.LogLevels.Toggle(flag)
}

// ==================================================
func (l *Logger) must(e error) {
	defer func() {
		if e, ok := recover().(error); ok {
			_, _ = fmt.Fprint(os.Stdout, e)
		}
	}()

	if e != nil {
		stack := debug.Stack()
		l.FATAL(string(stack))
		panic(e)
	}
}

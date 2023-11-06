package tests

import (
	"errors"
	"github.com/vaiktorg/grimoire/log"
	"os"
	"testing"
	"time"
)

var logger *log.Logger
var loggerServices *log.Logger
var msgTotal uint32

func TestMain(m *testing.M) {
	l, err := log.NewLogger()
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}

	l.LogLevels = log.LevelTrace | log.LevelDebug | log.LevelInfo | log.LevelWarn | log.LevelError | log.LevelFatal
	logger = l
	loggerServices = l.NewService("ServiceNameABC")
	defer l.Close()

	//Start Test
	m.Run()
}

func TestTrace(t *testing.T) {
	time.Sleep(time.Second / 4)

	logger.TRACE("TRACE MSG")
	messages := logger.Messages()

	trace := log.LevelTrace
	if messages[msgTotal].Level != trace.String() {
		t.FailNow()
	}

	println(messages[0].Level)
}
func TestDebug(t *testing.T) {
	time.Sleep(time.Second / 4)

	logger.DEBUG("DEBUG MSG")
	messages := logger.Messages()

	debug := log.LevelDebug
	if messages[len(messages)-1].Level != debug.String() {
		t.FailNow()
	}

	println(messages[msgTotal].Level)
}
func TestInfo(t *testing.T) {
	time.Sleep(time.Second / 4)

	logger.INFO("INFO MSG")
	messages := logger.Messages()

	info := log.LevelInfo
	if messages[len(messages)-1].Level != info.String() {
		t.FailNow()
	}

	println(messages[msgTotal].Level)
}
func TestWarn(t *testing.T) {
	time.Sleep(time.Second / 4)

	logger.WARN("ERROR MSG")
	messages := logger.Messages()

	warn := log.LevelWarn
	if messages[len(messages)-1].Level != warn.String() {
		t.FailNow()
	}

	println(messages[msgTotal].Level)
}
func TestError(t *testing.T) {
	time.Sleep(time.Second / 4)

	logger.ERROR(errors.New("ERROR MSG"))
	messages := logger.Messages()

	levelError := log.LevelError
	if messages[len(messages)-1].Level != levelError.String() {
		t.FailNow()
	}

	println(messages[msgTotal].Level)
}
func TestFatal(t *testing.T) {
	time.Sleep(time.Second / 4)

	logger.FATAL("FATAL MSG")
	messages := logger.Messages()

	fatal := log.LevelFatal
	if messages[len(messages)-1].Level != fatal.String() {
		t.FailNow()
	}

	println(messages[msgTotal].Level)
}

// ------------------
func TestServiceTrace(t *testing.T) {
	time.Sleep(time.Second / 4)

	loggerServices.TRACE("TRACE MSG")
	messages := loggerServices.Messages()

	trace := log.LevelTrace
	if messages[len(messages)-1].Level != trace.String() {
		t.FailNow()
	}

	println(messages[0].Level)
}
func TestServiceDebug(t *testing.T) {
	time.Sleep(time.Second / 4)

	loggerServices.DEBUG("DEBUG MSG")
	messages := loggerServices.Messages()

	debug := log.LevelDebug
	if messages[len(messages)-1].Level != debug.String() {
		t.FailNow()
	}

	println(messages[msgTotal].Level)
}
func TestServiceInfo(t *testing.T) {
	time.Sleep(time.Second / 4)

	loggerServices.INFO("INFO MSG")
	messages := loggerServices.Messages()

	info := log.LevelInfo
	if messages[len(messages)-1].Level != info.String() {
		t.FailNow()
	}

	println(messages[msgTotal].Level)
}
func TestServiceWarn(t *testing.T) {
	time.Sleep(time.Second / 4)

	loggerServices.WARN("ERROR MSG")
	messages := loggerServices.Messages()

	warn := log.LevelWarn
	if messages[len(messages)-1].Level != warn.String() {
		t.FailNow()
	}

	println(messages[msgTotal].Level)
}
func TestServiceError(t *testing.T) {
	time.Sleep(time.Second / 4)

	loggerServices.ERROR(errors.New("ERROR MSG"))
	messages := loggerServices.Messages()

	levelError := log.LevelError
	if messages[len(messages)-1].Level != levelError.String() {
		t.FailNow()
	}

	println(messages[msgTotal].Level)
}
func TestServiceFatal(t *testing.T) {
	time.Sleep(time.Second / 4)

	loggerServices.FATAL("FATAL MSG")
	messages := loggerServices.Messages()

	fatal := log.LevelFatal
	if messages[len(messages)-1].Level != fatal.String() {
		t.FailNow()
	}

	println(messages[msgTotal].Level)
}

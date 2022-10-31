package tests

import (
	"errors"
	"fmt"
	"github.com/vaiktorg/grimoire/log/log"
	"os"
	"sync/atomic"
	"testing"
	"time"
)

var logger *log.Logger
var msgTotal uint32

func TestMain(m *testing.M) {
	l, err := log.NewLogger("TestLoggerV0.1")
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
	l.LogLevels = log.LevelTrace | log.LevelDebug | log.LevelInfo | log.LevelWarn | log.LevelError | log.LevelFatal
	logger = l

	//Start Test
	defer logger.Close()
	m.Run()
}

func TestTrace(t *testing.T) {
	logger.TRACE("TRACE MSG")

	time.Sleep(time.Second / 4)
	msgs := logger.Messages()

	if msgs[msgTotal].Level != log.LevelTrace.String() {
		t.FailNow()
	}

	fmt.Println(msgs[0].Level)
}
func TestDebug(t *testing.T) {
	logger.DEBUG("DEBUG MSG")

	time.Sleep(time.Second / 4)
	msgs := logger.Messages()

	if msgs[atomic.AddUint32(&msgTotal, 1)].Level != log.LevelDebug.String() {
		t.FailNow()
	}

	fmt.Println(msgs[msgTotal].Level)
}
func TestInfo(t *testing.T) {
	logger.INFO("INFO MSG")

	time.Sleep(time.Second / 4)
	msgs := logger.Messages()

	if msgs[atomic.AddUint32(&msgTotal, 1)].Level != log.LevelInfo.String() {
		t.FailNow()
	}

	fmt.Println(msgs[msgTotal].Level)
}
func TestWarn(t *testing.T) {
	logger.WARN("ERROR MSG")

	time.Sleep(time.Second / 4)
	msgs := logger.Messages()

	if msgs[atomic.AddUint32(&msgTotal, 1)].Level != log.LevelWarn.String() {
		t.FailNow()
	}

	fmt.Println(msgs[msgTotal].Level)
}
func TestError(t *testing.T) {
	logger.ERROR(errors.New("ERROR MSG"))

	time.Sleep(time.Second / 4)
	msgs := logger.Messages()

	if msgs[atomic.AddUint32(&msgTotal, 1)].Level != log.LevelError.String() {
		t.FailNow()
	}

	fmt.Println(msgs[msgTotal].Level)
}
func TestFatal(t *testing.T) {
	logger.FATAL("FATAL MSG")

	time.Sleep(time.Second / 4)
	msgs := logger.Messages()

	if msgs[atomic.AddUint32(&msgTotal, 1)].Level != log.LevelFatal.String() {
		t.FailNow()
	}

	fmt.Println(msgs[msgTotal].Level)
}

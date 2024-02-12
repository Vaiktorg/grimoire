package main

import (
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type Log struct {
	service string
	msg     string
	id      string
}

type Logger struct {
	iwg sync.WaitGroup
	owg sync.WaitGroup

	service  string
	services map[string]*Logger

	in    chan Log
	out   chan Log
	inout chan Log

	forwarder func(Log)
}

func main() {
	nLogs := 1000000
	receivedMain := uint64(0)
	receivedServ := uint64(0)

	l := NewLogger()
	defer l.Close()

	s := l.NewService()

	wg := new(sync.WaitGroup)
	go l.Output(func(log Log) { // If you log 100 through Main, it will return 1M
		atomic.AddUint64(&receivedMain, 1)
		wg.Done()
	})

	go s.Output(func(log Log) { // If you send 100 through Service, it will return x2 in Main.
		atomic.AddUint64(&receivedServ, 1)
		wg.Done()
	})

	wg.Add(nLogs * 2)
	for i := 0; i < nLogs; i++ {
		s.Log(strconv.Itoa(i))
	}

	wg.Wait()

	println(receivedMain, receivedServ)
}

func NewLogger() *Logger {
	l := &Logger{
		in:    make(chan Log),
		inout: nil,
		out:   make(chan Log),

		service: "MainLogger",

		services: make(map[string]*Logger),
	}

	l.forwarder = func(log Log) {
		defer l.owg.Done()
		l.out <- log
	}

	go l.procLogs()

	return l
}

func (l *Logger) NewService() *Logger {
	s := &Logger{
		in:       make(chan Log),
		out:      make(chan Log),
		inout:    make(chan Log),
		service:  "ServiceLogger",
		services: make(map[string]*Logger),
	}

	s.forwarder = func(log Log) {
		defer s.owg.Done()

		s.inout <- log
		s.out <- log
	}

	go s.forwardLogs(l.out)
	go s.procLogs()

	l.services[s.service] = s

	return s
}

func (l *Logger) Log(msg string) {
	l.iwg.Add(1)
	l.in <- Log{
		service: l.service,
		msg:     msg,
		id:      time.Now().String(),
	}
}
func (l *Logger) Output(h func(Log)) {
	for log := range l.out {
		h(log)
	}
}
func (l *Logger) Close() {
	for serv, logger := range l.services {
		logger.Close()
		delete(l.services, serv)
	}

	close(l.in)
	l.iwg.Wait()

	if l.inout != nil {
		close(l.inout)
	}
	close(l.out)

	l.owg.Wait()
}

func (l *Logger) forwardLogs(parentChan chan Log) {
	for log := range l.inout {
		parentChan <- log
	}
}

func (l *Logger) procLogs() {
	for log := range l.in {
		l.owg.Add(1)
		go l.forwarder(log)

		l.iwg.Done()
	}
}

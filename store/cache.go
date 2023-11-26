package store

import (
	"encoding/json"
	"fmt"
	"github.com/vaiktorg/grimoire/uid"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"sync"
	"time"
)

type Cache[T any] struct {
	mu sync.Mutex

	appName       string
	runId         string
	size          uint64
	flushTotal    int
	logFilesPaths map[string][]string //Map [Service_RunID] []Paths

	buff   []T
	cap    int
	wPos   int
	rPos   int
	isFull bool
}

func NewCache[T any](name string) *Cache[T] {
	return &Cache[T]{
		appName:       name,
		runId:         uid.NewUID(8).String(),
		logFilesPaths: make(map[string][]string),
	}
}

func NewIDCache[T any](cacheName string, runId string) *Cache[T] {

	return &Cache[T]{
		appName:       cacheName,
		runId:         runId,
		logFilesPaths: make(map[string][]string),
		buff:          make([]T, DefaultLen),
		cap:           DefaultLen,
	}
}

func (c *Cache[T]) Size() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return int(c.size)
}
func (c *Cache[T]) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isFull {
		return c.cap // The buffer is full
	}
	return (c.wPos - c.rPos + c.cap) % c.cap
}
func (c *Cache[T]) IsFull() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.isFull
}
func (c *Cache[T]) Write(msg T) {
	c.mu.Lock()
	defer c.mu.Unlock()

	logBytes, err := json.Marshal(msg)
	if err != nil {
		fmt.Printf("error marshaling log message: %v", err)
		return
	}
	logSize := uint64(len(logBytes))

	c.size += logSize

	c.buff[c.wPos] = msg
	c.wPos = (c.wPos + 1) % c.cap

	if c.isFull {
		c.rPos = (c.rPos + 1) % c.cap
	}

	if c.wPos == c.rPos {
		c.isFull = true
	}
}

func (c *Cache[T]) FlushLen() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.flushTotal
}

func (c *Cache[T]) ReadAll(runName string) []T {
	c.mu.Lock()
	defer c.mu.Unlock()

	var logs []T
	if c.isFull {
		logs = append(logs, c.buff[c.rPos:]...)
		logs = append(logs, c.buff[:c.wPos]...)
	} else {
		logs = append(logs, c.buff[:c.wPos]...)
	}

	buffLen := 0
	if c.isFull {
		buffLen = c.cap // The buffer is full
	} else {
		buffLen = (c.wPos - c.rPos + c.cap) % c.cap
	}

	if c.flushTotal > 0 && buffLen < DefaultLen {
		key := runName + "_" + c.runId
		logs = append(logs, c.fromFiles(key)...)
		return logs
	}

	return c.buff
}

func (c *Cache[T]) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.wPos == 0 && !c.isFull {
		return
	}

	date := time.Now().Format("20060102150405")
	filename := date + "__" + c.appName + "__" + c.runId + "__" + strconv.Itoa(c.flushTotal) + ".log"

	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}

	defer file.Close()

	var toFlush []T
	if c.isFull {
		toFlush = append(toFlush, c.buff[c.rPos:]...)
		toFlush = append(toFlush, c.buff[:c.wPos]...)
	} else {
		toFlush = c.buff[:c.wPos]
	}

	if err = json.NewEncoder(file).Encode(toFlush); err != nil {
		panic(err)
	}

	c.flushTotal++
	c.Clear()
}

func (c *Cache[T]) Clear() {
	c.size = 0
	c.wPos = 0
	c.rPos = 0
	c.isFull = false
}

func (c *Cache[T]) fromFiles(key string) []T {
	if err := c.readLogFiles(); err != nil {
		return nil
	}

	var logs []T
	for _, path := range c.logFilesPaths[key] {
		file, er := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
		if er != nil {
			return nil
		}

		var fileLogs []T
		er = json.NewDecoder(file).Decode(&fileLogs)
		if er != nil {
			return nil
		}

		logs = append(logs, fileLogs...)

		file.Close()
	}

	return logs
}

func (c *Cache[T]) readLogFiles() error {
	return filepath.Walk(".", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		parts := regexp.MustCompile(`__([^_]+)__([^_]+)__`).FindStringSubmatch(filepath.Base(path))
		if len(parts) != 3 {
			return nil
		}

		//parts[0] : BasePath
		//parts[1] : ServiceName
		//parts[2] : RunID

		if filepath.Ext(info.Name()) == ".log" {
			key := parts[1] + "_" + parts[2]
			c.logFilesPaths[key] = append(c.logFilesPaths[key], path)
		}

		return nil
	})
}

const (
	DefaultSize = 1 * MB
	DefaultLen  = 10000
)

type Size int

func (s Size) Val() int { return int(s) }

const (
	_ Size = 1.0 << (10 * iota) // ignore first value by assigning to blank identifier
	KB
	MB
	GB
)

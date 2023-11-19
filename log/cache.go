package log

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"sync"
	"time"
)

const (
	DefaultLogSize = 1 * MB
	DefaultLogLen  = 10000
)

type Cache[T any] struct {
	mu sync.Mutex

	appName       string
	RunId         string
	size          uint64
	flushTotal    uint64
	logFilesPaths map[string][]string //Map [Service_RunID] []Paths

	c []T
}

func NewCache[T any](cacheName string, runId string) *Cache[T] {
	return &Cache[T]{
		appName:       cacheName,
		RunId:         runId,
		logFilesPaths: make(map[string][]string),
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

	return len(c.c)
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

	c.c = append(c.c, msg)
}

func (c *Cache[T]) Collection(serviceName string) []T {
	c.mu.Lock()
	defer c.mu.Unlock()

	var logs []T
	if len(c.c) < DefaultLogLen && c.flushTotal > 0 {
		key := serviceName + "_" + c.RunId
		return append(append(logs, c.FromFiles(key)...), c.c...)
	}

	return c.c
}

func (c *Cache[T]) Flush() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.c) == 0 {
		return nil
	}

	date := time.Now().Format("20060102150405")
	filename := date + "__" + c.appName + "__" + c.RunId + "__" + strconv.Itoa(int(c.flushTotal)) + ".log"

	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	if err = json.NewEncoder(file).Encode(c.c); err != nil {
		return err
	}

	c.flushTotal++
	c.Clear()

	return nil
}

func (c *Cache[T]) Clear() {
	c.c = []T{}
	c.size = 0
}

func (c *Cache[T]) FromFiles(key string) []T {
	err := c.readLogFiles()
	if err != nil {
		return nil
	}

	var logs []T
	for _, path := range c.logFilesPaths[key] {
		file, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
		if err != nil {
			return nil
		}

		var fileLogs []T
		err = json.NewDecoder(file).Decode(&fileLogs)
		if err != nil {
			return nil
		}

		logs = append(logs, fileLogs...)

		file.Close()
	}

	return logs
}

// UID generates a unique identifier string.
// The length of the UID is specified by uidLength.
func UID(uidLength int) string {
	const lut = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// Buffer to hold random bytes
	randomBytes := make([]byte, uidLength)
	if _, err := rand.Read(randomBytes); err != nil {
		return ""
	}

	// Create UID using the LUT
	uid := make([]byte, uidLength)
	for i, b := range randomBytes {
		// Using bitwise AND to make sure index is within the range of lut
		index := b & (byte(len(lut) - 1))
		uid[i] = lut[index]
	}

	return string(uid)
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

//func (c *Cache[T]) fetchCurrentBatch(paths ...string) []string {
//	logRunPaths := make(map[string][]string)
//	for _, path := range paths {
//		//parts := regexp.MustCompile(`^(\d{14})__([^_]+)__([^_]+)__([0-9]+)\.log$`).FindStringSubmatch(path)
//		parts := regexp.MustCompile(`__([^_]+)__([^_]+)__`).FindStringSubmatch(path)
//		if len(parts) == 3 {
//			//date := parts[1]     // DateTime Tracking
//			//services := parts[2]  // Constant ID
//			//RunId := parts[3]    // Constant ID
//			//batchIdx := parts[4] // BatchOrder: Should be sequential, 0 Idx
//
//			services := parts[1] // Constant ID
//			RunId := parts[2]   // Constant ID
//
//			key := services + "_" + RunId
//			if _, ok := logRunPaths[key]; !ok {
//				logRunPaths[key] = append(logRunPaths[key], path)
//			}
//		}
//
//	}
//}

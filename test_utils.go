package main

import (
	"fmt"
	"log"
	"sync"
)

// TestLogger implements Logger interface for testing with thread-safe entry tracking
type TestLogger struct {
	entries []string
	mutex   sync.Mutex
}

func NewTestLogger() *TestLogger {
	return &TestLogger{
		entries: make([]string, 0),
	}
}

func (tl *TestLogger) Info(format string, args ...interface{}) {
	tl.mutex.Lock()
	defer tl.mutex.Unlock()
	tl.entries = append(tl.entries, fmt.Sprintf("INFO: "+format, args...))
	log.Printf("[INFO] "+format, args...)
}

func (tl *TestLogger) Debug(format string, args ...interface{}) {
	tl.mutex.Lock()
	defer tl.mutex.Unlock()
	tl.entries = append(tl.entries, fmt.Sprintf("DEBUG: "+format, args...))
	log.Printf("[DEBUG] "+format, args...)
}

func (tl *TestLogger) Warn(format string, args ...interface{}) {
	tl.mutex.Lock()
	defer tl.mutex.Unlock()
	tl.entries = append(tl.entries, fmt.Sprintf("WARN: "+format, args...))
	log.Printf("[WARN] "+format, args...)
}

func (tl *TestLogger) Error(format string, args ...interface{}) {
	tl.mutex.Lock()
	defer tl.mutex.Unlock()
	tl.entries = append(tl.entries, fmt.Sprintf("ERROR: "+format, args...))
	log.Printf("[ERROR] "+format, args...)
}

func (tl *TestLogger) GetEntries() []string {
	tl.mutex.Lock()
	defer tl.mutex.Unlock()
	entries := make([]string, len(tl.entries))
	copy(entries, tl.entries)
	return entries
}

func (tl *TestLogger) Reset() {
	tl.mutex.Lock()
	defer tl.mutex.Unlock()
	tl.entries = []string{}
}

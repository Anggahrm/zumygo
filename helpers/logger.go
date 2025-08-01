package helpers

import (
	"io"
	"log"
	"os"
	"sync"
	"time"
	"bufio"
	"bytes"
)

var (
	WarningLogger *log.Logger
	InfoLogger    *log.Logger
	ErrorLogger   *log.Logger
	
	// Performance optimizations
	logBuffer    *bytes.Buffer
	logMutex     sync.Mutex
	flushTicker  *time.Ticker
	stopFlush    chan bool
)

type Logger struct{}

func init() {
	// Initialize buffer for async logging
	logBuffer = bytes.NewBuffer(make([]byte, 0, 4096))
	
	// Create log file with rotation support
	file, err := os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		// If we can't create log file, use stderr only
		InfoLogger = log.New(os.Stderr, "INFO: ", log.Ldate|log.Ltime)
		WarningLogger = log.New(os.Stderr, "WARNING: ", log.Ldate|log.Ltime)
		ErrorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
		ErrorLogger.Printf("Failed to create log file: %v, using stderr only", err)
		return
	}
	
	// Use buffered writer for better performance
	bufferedFile := bufio.NewWriter(file)
	
	// Create multi-writer for both file and stderr
	multiWriter := io.MultiWriter(bufferedFile, os.Stderr)
	
	InfoLogger = log.New(multiWriter, "INFO: ", log.Ldate|log.Ltime)
	WarningLogger = log.New(multiWriter, "WARNING: ", log.Ldate|log.Ltime)
	ErrorLogger = log.New(multiWriter, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	
	// Start async flush routine
	startAsyncFlush(bufferedFile)
}

// startAsyncFlush starts a goroutine to periodically flush the log buffer
func startAsyncFlush(writer *bufio.Writer) {
	flushTicker = time.NewTicker(5 * time.Second) // Flush every 5 seconds
	stopFlush = make(chan bool)
	
	go func() {
		for {
			select {
			case <-flushTicker.C:
				logMutex.Lock()
				if err := writer.Flush(); err != nil {
					// Log error to stderr since we can't use the logger
					os.Stderr.WriteString("Failed to flush log buffer: " + err.Error() + "\n")
				}
				logMutex.Unlock()
			case <-stopFlush:
				flushTicker.Stop()
				logMutex.Lock()
				writer.Flush()
				logMutex.Unlock()
				return
			}
		}
	}()
}

// StopLogger stops the async flush routine
func StopLogger() {
	if stopFlush != nil {
		close(stopFlush)
	}
}

func (log Logger) Info(v any) {
	if InfoLogger != nil {
		logMutex.Lock()
		InfoLogger.Println(v)
		logMutex.Unlock()
	}
}

func (log Logger) Warn(v any) {
	if WarningLogger != nil {
		logMutex.Lock()
		WarningLogger.Println(v)
		logMutex.Unlock()
	}
}

func (log Logger) Error(v any) {
	if ErrorLogger != nil {
		logMutex.Lock()
		ErrorLogger.Println(v)
		logMutex.Unlock()
	}
}

// AsyncInfo logs info message asynchronously
func (log Logger) AsyncInfo(v any) {
	go func() {
		log.Info(v)
	}()
}

// AsyncWarn logs warning message asynchronously
func (log Logger) AsyncWarn(v any) {
	go func() {
		log.Warn(v)
	}()
}

// AsyncError logs error message asynchronously
func (log Logger) AsyncError(v any) {
	go func() {
		log.Error(v)
	}()
}

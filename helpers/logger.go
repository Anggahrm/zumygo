package helpers

import (
	"io"
	"log"
	"os"
)

var (
	WarningLogger *log.Logger
	InfoLogger    *log.Logger
	ErrorLogger   *log.Logger
)

type Logger struct{}

func init() {
	file, err := os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		// If we can't create log file, use stderr only
		InfoLogger = log.New(os.Stderr, "INFO: ", log.Ldate|log.Ltime)
		WarningLogger = log.New(os.Stderr, "WARNING: ", log.Ldate|log.Ltime)
		ErrorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
		ErrorLogger.Printf("Failed to create log file: %v, using stderr only", err)
		return
	}
	
	InfoLogger = log.New(io.MultiWriter(file, os.Stderr), "INFO: ", log.Ldate|log.Ltime)
	WarningLogger = log.New(io.MultiWriter(file, os.Stderr), "WARNING: ", log.Ldate|log.Ltime)
	ErrorLogger = log.New(io.MultiWriter(file, os.Stderr), "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func (log Logger) Info(v any) {
	if InfoLogger != nil {
		InfoLogger.Println(v)
	}
}

func (log Logger) Warn(v any) {
	if WarningLogger != nil {
		WarningLogger.Println(v)
	}
}

func (log Logger) Error(v any) {
	if ErrorLogger != nil {
		ErrorLogger.Println(v)
	}
}

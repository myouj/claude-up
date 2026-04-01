package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Level represents a log level.
type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
	FATAL
)

func (l Level) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Logger is a structured JSON logger.
type Logger struct {
	mu       sync.Mutex
	output   io.Writer
	level    Level
	minLevel Level
}

// logEntry represents a structured log entry.
type logEntry struct {
	Time    string                 `json:"time"`
	Level   string                 `json:"level"`
	TraceID string                 `json:"trace_id,omitempty"`
	Msg     string                 `json:"msg"`
	Fields  map[string]interface{} `json:"fields,omitempty"`
}

// NewLogger creates a new Logger writing to w.
func NewLogger(w io.Writer, minLevel Level) *Logger {
	return &Logger{
		output:   w,
		level:    INFO,
		minLevel: minLevel,
	}
}

// SetMinLevel sets the minimum log level.
func (l *Logger) SetMinLevel(level Level) {
	l.minLevel = level
}

func (l *Logger) log(level Level, traceID, msg string, fields map[string]interface{}) {
	if level < l.minLevel {
		return
	}

	entry := logEntry{
		Time:    time.Now().UTC().Format(time.RFC3339Nano),
		Level:   level.String(),
		TraceID: traceID,
		Msg:     msg,
		Fields:  fields,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(l.output, `{"time":"%s","level":"ERROR","msg":"logger: failed to marshal log entry: %v"}`+"\n",
			time.Now().UTC().Format(time.RFC3339Nano), err)
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.output.Write(append(data, '\n'))
}

// Debug logs a debug message.
func (l *Logger) Debug(msg string, fields ...map[string]interface{}) {
	l.log(DEBUG, "", msg, mergeFields(fields))
}

// Info logs an info message.
func (l *Logger) Info(msg string, fields ...map[string]interface{}) {
	l.log(INFO, "", msg, mergeFields(fields))
}

// Warn logs a warning message.
func (l *Logger) Warn(msg string, fields ...map[string]interface{}) {
	l.log(WARN, "", msg, mergeFields(fields))
}

// Error logs an error message.
func (l *Logger) Error(msg string, fields ...map[string]interface{}) {
	l.log(ERROR, "", msg, mergeFields(fields))
}

// Fatal logs a fatal message and exits.
func (l *Logger) Fatal(msg string, fields ...map[string]interface{}) {
	l.log(FATAL, "", msg, mergeFields(fields))
	os.Exit(1)
}

// WithTraceID returns a trace-bound logger.
func (l *Logger) WithTraceID(traceID string) *LoggerWithTrace {
	return &LoggerWithTrace{logger: l, traceID: traceID}
}

// LoggerWithTrace is a logger bound to a specific trace ID.
type LoggerWithTrace struct {
	logger  *Logger
	traceID string
}

func (l *LoggerWithTrace) Debug(msg string, fields ...map[string]interface{}) {
	l.logger.log(DEBUG, l.traceID, msg, mergeFields(fields))
}

func (l *LoggerWithTrace) Info(msg string, fields ...map[string]interface{}) {
	l.logger.log(INFO, l.traceID, msg, mergeFields(fields))
}

func (l *LoggerWithTrace) Warn(msg string, fields ...map[string]interface{}) {
	l.logger.log(WARN, l.traceID, msg, mergeFields(fields))
}

func (l *LoggerWithTrace) Error(msg string, fields ...map[string]interface{}) {
	l.logger.log(ERROR, l.traceID, msg, mergeFields(fields))
}

func (l *LoggerWithTrace) Fatal(msg string, fields ...map[string]interface{}) {
	l.logger.log(FATAL, l.traceID, msg, mergeFields(fields))
}

// StackTrace captures a stack trace, skipping skipFrames from the top.
func StackTrace(skip int) string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	lines := splitLines(string(buf[:n]))
	if len(lines) > skip*2 {
		return joinLines(lines[skip*2:])
	}
	return joinLines(lines)
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func joinLines(lines []string) string {
	return strings.Join(lines, "\n") + "\n"
}

func mergeFields(fields []map[string]interface{}) map[string]interface{} {
	if len(fields) == 0 {
		return nil
	}
	return fields[0]
}

// DefaultLogger is the package-level logger (stdout, INFO level).
var DefaultLogger = NewLogger(os.Stdout, INFO)

// SetOutput redirects the default logger output.
func SetOutput(w io.Writer) {
	DefaultLogger = NewLogger(w, INFO)
}

// SetLevel sets the default logger level.
func SetLevel(level Level) {
	DefaultLogger.SetMinLevel(level)
}

// Debug logs via DefaultLogger.
func Debug(msg string, fields ...map[string]interface{}) {
	DefaultLogger.Debug(msg, fields...)
}

// Info logs via DefaultLogger.
func Info(msg string, fields ...map[string]interface{}) {
	DefaultLogger.Info(msg, fields...)
}

// Warn logs via DefaultLogger.
func Warn(msg string, fields ...map[string]interface{}) {
	DefaultLogger.Warn(msg, fields...)
}

// Error logs via DefaultLogger.
func Error(msg string, fields ...map[string]interface{}) {
	DefaultLogger.Error(msg, fields...)
}

// Fatal logs via DefaultLogger and exits.
func Fatal(msg string, fields ...map[string]interface{}) {
	DefaultLogger.Fatal(msg, fields...)
}

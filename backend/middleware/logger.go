package middleware

import "prompt-vault/utils"

// Re-export logger types and functions from utils.
type (
	Level           = utils.Level
	Logger          = utils.Logger
	LoggerWithTrace = utils.LoggerWithTrace
)

// Log level constants.
const (
	DEBUG = utils.DEBUG
	INFO  = utils.INFO
	WARN  = utils.WARN
	ERROR = utils.ERROR
	FATAL = utils.FATAL
)

const (
	HeaderTraceID   = "X-Trace-ID"
	TraceIDKey      = "trace_id"
	TraceLoggerKey  = "trace_logger"
)

var (
	NewLogger    = utils.NewLogger
	SetOutput    = utils.SetOutput
	SetLevel     = utils.SetLevel
	Debug        = utils.Debug
	Info         = utils.Info
	Warn         = utils.Warn
	Error        = utils.Error
	Fatal        = utils.Fatal
	StackTrace   = utils.StackTrace
	DefaultLogger = utils.DefaultLogger
)

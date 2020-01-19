package event

import (
	"runtime/debug"
	"strings"
)

// Level maps debug level
type Level uint32

const (
	// UNSPECIFIED is error level
	UNSPECIFIED Level = iota
	TRACE
	DEBUG
	INFO
	WARNING
	ERROR
	CRITICAL
)

// Error type represents error with stacktrace and error message
type Error struct {
	Stack   string
	Message string
	Level   string
}

// FormatError returns formattet error message
func FormatError(level, msg string) Error {
	stack := string(debug.Stack())

	return Error{
		Stack:   stack,
		Message: msg,
		Level:   parseLevel(level).String(),
	}
}

func parseLevel(level string) Level {
	level = strings.ToUpper(level)
	switch level {
	case "UNSPECIFIED":
		return UNSPECIFIED
	case "TRACE":
		return TRACE
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN", "WARNING":
		return WARNING
	case "ERROR":
		return ERROR
	case "CRITICAL":
		return CRITICAL
	default:
		return UNSPECIFIED
	}
}

// String implements Stringer.
func (level Level) String() string {
	switch level {
	case UNSPECIFIED:
		return "UNSPECIFIED"
	case TRACE:
		return "TRACE"
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARNING:
		return "WARNING"
	case ERROR:
		return "ERROR"
	case CRITICAL:
		return "CRITICAL"
	default:
		return "<unknown>"
	}
}

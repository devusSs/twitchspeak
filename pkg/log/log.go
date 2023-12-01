package log

import (
	"fmt"
	"io"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Option is a function that modifies the logger
type Option func(*Logger)

// WithName sets the logger name
func WithName(name string) Option {
	return func(l *Logger) {
		l.name = name
	}
}

// WithDebug enables debug level logging
func WithDebug(debug bool) Option {
	if debug {
		return func(l *Logger) {
			l.console = true
			l.out.Level(zerolog.DebugLevel)
		}
	}
	return func(l *Logger) {
		l.out.Level(zerolog.InfoLevel)
	}
}

// WithConsole enables (pretty) logging to console
func WithConsole(console bool) Option {
	if console {
		return func(l *Logger) {
			l.console = true
		}
	}
	return func(l *Logger) {
		l.console = false
	}
}

// WithOwnLogFile sets the log file name to the given name
func WithOwnLogFile(name string) Option {
	return func(l *Logger) {
		l.out = zerolog.New(newRotatingLogFile(logsDirectory, name))
	}
}

// Logger is a simple wrapper around zerolog
type Logger struct {
	name    string
	console bool
	out     zerolog.Logger
}

// GetWriter returns the logger's writer
func (l *Logger) GetWriter() io.Writer {
	return l.out
}

// Implementation for the Gorm logger interface,
// does not support colored output
func (l *Logger) Printf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)

	l.out.Info().Any("caller", getCaller()).Msg(msg)
	if l.console {
		fmt.Printf(
			"[%s] [%s] %s\n",
			time.Now().Format("15:04:05"),
			strings.ToUpper(l.name),
			msg,
		)
	}
}

// Debug logs a message to l.out if log level is debug
//
// Also (pretty) prints to os.Stdout if l.console is true
func (l *Logger) Debug(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)

	if e := l.out.Debug(); e.Enabled() {
		e.Any("caller", getCaller()).Msg(msg)
		if l.console {
			fmt.Printf(
				"[%s] [%s] %s\n",
				color.GreenString(time.Now().Format("15:04:05")),
				color.WhiteString(strings.ToUpper(l.name)),
				msg,
			)
		}
	}
}

// Info logs a message to l.out
//
// Also (pretty) prints to os.Stdout if l.console is true
func (l *Logger) Info(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)

	l.out.Info().Any("caller", getCaller()).Msg(msg)
	if l.console {
		fmt.Printf(
			"[%s] [%s] %s\n",
			color.GreenString(time.Now().Format("15:04:05")),
			color.BlueString(strings.ToUpper(l.name)),
			msg,
		)
	}
}

// Warn logs a message to l.out
//
// Also (pretty) prints to os.Stdout if l.console is true
func (l *Logger) Warn(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)

	l.out.Warn().Any("caller", getCaller()).Msg(msg)
	if l.console {
		fmt.Printf(
			"[%s] [%s] %s\n",
			color.GreenString(time.Now().Format("15:04:05")),
			color.YellowString(strings.ToUpper(l.name)),
			msg,
		)
	}
}

// Error logs a message to l.out
//
// Also (pretty) prints to os.Stderr if l.console is true
func (l *Logger) Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)

	l.out.Error().Any("caller", getCaller()).Msg(msg)
	if l.console {
		_, err := fmt.Fprintf(
			color.Error,
			"[%s] [%s] %s\n",
			color.GreenString(time.Now().Format("15:04:05")),
			color.RedString(strings.ToUpper(l.name)),
			msg,
		)
		if err != nil {
			fmt.Println("could not print to stderr:", err)
		}
	}
}

// NewLogger creates a new logger with the given options
//
// # NOTE: Please use WithConsole() before WithDebug() if you desire to use both
func NewLogger(options ...Option) *Logger {
	l := &Logger{
		name:    "app",
		console: false,
		out:     zerolog.New(newRotatingLogFile(logsDirectory, logFileName)),
	}

	for _, option := range options {
		option(l)
	}

	return l
}

// Sets the default logs directory, if empty or not set
// via this function, will use "./logs"
func SetDefaultLogsDirectory(dir string) {
	if dir == "" {
		return
	}
	logsDirectory = dir
}

// Sets the default log file name, if empty or not set
// via this function, will use "app.log"
func SetDefaultLogFileName(name string) {
	if name == "" {
		return
	}
	logFileName = name
}

const (
	defaultLogsDirectory = "./logs"
	defaultLogFilename   = "app.log"
)

var (
	logsDirectory string = defaultLogsDirectory
	logFileName   string = defaultLogFilename
)

func newRotatingLogFile(dir string, name string) io.Writer {
	return &lumberjack.Logger{
		Filename:   fmt.Sprintf("%s/%s", dir, name),
		MaxSize:    25,
		MaxAge:     28,
		MaxBackups: 0,
		LocalTime:  true,
		Compress:   false,
	}
}

func getCaller() string {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		return "unknown"
	}
	return fmt.Sprintf("%s:%d", filepath.Base(file), line)
}

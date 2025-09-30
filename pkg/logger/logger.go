package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"
	"sync"
	"time"
)

// CustomHandler - custom handler for slog
type CustomHandler struct {
	opts   HandlerOptions
	mu     *sync.Mutex
	writer io.Writer
	attrs  []slog.Attr
	groups []string
}

// HandlerOptions - options for custom handler
type HandlerOptions struct {
	Level     slog.Leveler
	AddSource bool
	Formatter Formatter
}

// Formatter - interface for log formatting
type Formatter interface {
	Format(record slog.Record, attrs []slog.Attr, groups []string) ([]byte, error)
}

// JSONFormatter - JSON formatter
type JSONFormatter struct{}

func (f *JSONFormatter) Format(record slog.Record, attrs []slog.Attr, groups []string) ([]byte, error) {
	logEntry := make(map[string]interface{})

	if !record.Time.IsZero() {
		logEntry["time"] = record.Time.Format(time.RFC3339Nano)
	}
	logEntry["level"] = record.Level.String()
	logEntry["msg"] = record.Message

	if record.PC != 0 {
		fs := runtime.CallersFrames([]uintptr{record.PC})
		f, _ := fs.Next()
		logEntry["source"] = map[string]interface{}{
			"file": f.File,
			"line": f.Line,
			"func": f.Function,
		}
	}

	current := logEntry
	for _, group := range groups {
		groupMap := make(map[string]interface{})
		current[group] = groupMap
		current = groupMap
	}

	for _, attr := range attrs {
		current[attr.Key] = attr.Value.Any()
	}

	record.Attrs(func(attr slog.Attr) bool {
		current[attr.Key] = attr.Value.Any()
		return true
	})

	b, err := json.Marshal(logEntry)
	if err != nil {
		return nil, err
	}
	b = append(b, '\n')
	return b, nil
}

// TextFormatter - text formatter
type TextFormatter struct{}

func (f *TextFormatter) Format(record slog.Record, attrs []slog.Attr, groups []string) ([]byte, error) {
	var buf []byte

	// Time
	if !record.Time.IsZero() {
		buf = fmt.Appendf(buf, "time=%s ", record.Time.Format("2006-01-02T15:04:05"))
	}

	// Level
	buf = fmt.Appendf(buf, "level=%s ", record.Level.String())

	// Message
	buf = fmt.Appendf(buf, "msg=%q ", record.Message)

	// Source
	if record.PC != 0 {
		fs := runtime.CallersFrames([]uintptr{record.PC})
		f, _ := fs.Next()
		buf = fmt.Appendf(buf, "source=%s:%d ", f.File, f.Line)
	}

	// Groups as prefix
	groupPrefix := ""
	for _, group := range groups {
		groupPrefix += group + "."
	}

	// Attributes from WithAttrs
	for _, attr := range attrs {
		buf = fmt.Appendf(buf, "%s%s=%v ", groupPrefix, attr.Key, attr.Value.Any())
	}

	// Attributes from Record
	record.Attrs(func(attr slog.Attr) bool {
		buf = fmt.Appendf(buf, "%s%s=%v ", groupPrefix, attr.Key, attr.Value.Any())
		return true
	})

	buf = append(buf, '\n')
	return buf, nil
}

// NewCustomHandler creates a new custom handler
func NewCustomHandler(writer io.Writer, opts *HandlerOptions) *CustomHandler {
	if opts == nil {
		opts = &HandlerOptions{}
	}
	if opts.Level == nil {
		opts.Level = slog.LevelInfo
	}
	if opts.Formatter == nil {
		opts.Formatter = &JSONFormatter{}
	}

	return &CustomHandler{
		opts:   *opts,
		mu:     &sync.Mutex{},
		writer: writer,
		attrs:  make([]slog.Attr, 0),
		groups: make([]string, 0),
	}
}

// Enabled checks if the given log level is enabled
func (h *CustomHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.opts.Level.Level()
}

// Handle processes a log record
func (h *CustomHandler) Handle(ctx context.Context, record slog.Record) error {
	if !h.Enabled(ctx, record.Level) {
		return nil
	}

	if h.opts.AddSource && record.PC == 0 {
		var pcs [1]uintptr
		runtime.Callers(4, pcs[:])
		record.PC = pcs[0]
	}

	data, err := h.opts.Formatter.Format(record, h.attrs, h.groups)
	if err != nil {
		return err
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	_, err = h.writer.Write(data)
	return err
}

// WithAttrs creates a new handler with additional attributes
func (h *CustomHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}

	h2 := *h
	h2.attrs = make([]slog.Attr, 0, len(h.attrs)+len(attrs))
	h2.attrs = append(h2.attrs, h.attrs...)
	h2.attrs = append(h2.attrs, attrs...)

	return &h2
}

// WithGroup creates a new handler with a group
func (h *CustomHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	h2 := *h
	h2.groups = make([]string, 0, len(h.groups)+1)
	h2.groups = append(h2.groups, h.groups...)
	h2.groups = append(h2.groups, name)

	return &h2
}

// NewLogger creates a new logger with custom handler
func NewLogger(writer io.Writer, level slog.Level, formatter Formatter, addSource bool) *slog.Logger {
	opts := &HandlerOptions{
		Level:     level,
		Formatter: formatter,
		AddSource: addSource,
	}

	handler := NewCustomHandler(writer, opts)
	return slog.New(handler)
}

// Helper functions for creating different types of loggers
// NewJSONLogger creates a new JSON logger
func NewJSONLogger(writer io.Writer, level slog.Level) *slog.Logger {
	return NewLogger(writer, level, &JSONFormatter{}, false)
}

// NewTextLogger creates a new text logger
func NewTextLogger(writer io.Writer, level slog.Level) *slog.Logger {
	return NewLogger(writer, level, &TextFormatter{}, false)
}

// NewDevelopmentLogger creates a logger for development environment
func NewDevelopmentLogger() *slog.Logger {
	return NewTextLogger(os.Stdout, slog.LevelDebug)
}

// NewProductionLogger creates a logger for production environment
func NewProductionLogger() *slog.Logger {
	return NewJSONLogger(os.Stdout, slog.LevelInfo)
}

var Logger = NewProductionLogger()

// SetDebug sets the global logger to debug mode
func SetDebug() {
	Logger = NewDevelopmentLogger()
}

// SetProduction sets the global logger to production mode
func SetProduction() {
	Logger = NewProductionLogger()
}

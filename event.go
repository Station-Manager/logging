package logging

import (
	"github.com/rs/zerolog"
	"net"
	"time"
)

// LogContext provides a fluent interface for building a context logger with pre-populated fields.
// Fields added through LogContext will be included in all subsequent log messages.
type LogContext interface {
	Str(key, val string) LogContext
	Strs(key string, vals []string) LogContext
	Int(key string, val int) LogContext
	Int64(key string, val int64) LogContext
	Uint(key string, val uint) LogContext
	Uint64(key string, val uint64) LogContext
	Float64(key string, val float64) LogContext
	Bool(key string, val bool) LogContext
	Time(key string, val time.Time) LogContext
	Err(err error) LogContext
	Interface(key string, val interface{}) LogContext
	// Logger creates and returns the new context logger
	Logger() Logger
}

// LogEvent provides a fluent interface for structured logging with type-safe field methods.
// It wraps zerolog.Event to provide a clean API for adding typed fields to log entries.
type LogEvent interface {
	Str(key, val string) LogEvent
	Strs(key string, vals []string) LogEvent
	Stringer(key string, val interface{ String() string }) LogEvent
	Int(key string, val int) LogEvent
	Int8(key string, val int8) LogEvent
	Int16(key string, val int16) LogEvent
	Int32(key string, val int32) LogEvent
	Int64(key string, val int64) LogEvent
	Uint(key string, val uint) LogEvent
	Uint8(key string, val uint8) LogEvent
	Uint16(key string, val uint16) LogEvent
	Uint32(key string, val uint32) LogEvent
	Uint64(key string, val uint64) LogEvent
	Float32(key string, val float32) LogEvent
	Float64(key string, val float64) LogEvent
	Bool(key string, val bool) LogEvent
	Bools(key string, vals []bool) LogEvent
	Time(key string, val time.Time) LogEvent
	Dur(key string, val time.Duration) LogEvent
	Err(err error) LogEvent
	AnErr(key string, err error) LogEvent
	Bytes(key string, val []byte) LogEvent
	Hex(key string, val []byte) LogEvent
	IPAddr(key string, val net.IP) LogEvent
	MACAddr(key string, val net.HardwareAddr) LogEvent
	Interface(key string, val interface{}) LogEvent
	Dict(key string, dict func(LogEvent)) LogEvent
	Msg(msg string)
	Msgf(format string, v ...interface{})
	Send()
}

// logEvent implements LogEvent by wrapping zerolog.Event
type logEvent struct {
	event *zerolog.Event
}

// newLogEvent creates a new LogEvent wrapper
func newLogEvent(e *zerolog.Event) LogEvent {
	if e == nil {
		return &logEvent{event: nil}
	}
	return &logEvent{event: e}
}

func (e *logEvent) Str(key, val string) LogEvent {
	if e.event != nil {
		e.event.Str(key, val)
	}
	return e
}

func (e *logEvent) Strs(key string, vals []string) LogEvent {
	if e.event != nil {
		e.event.Strs(key, vals)
	}
	return e
}

func (e *logEvent) Stringer(key string, val interface{ String() string }) LogEvent {
	if e.event != nil {
		e.event.Stringer(key, val)
	}
	return e
}

func (e *logEvent) Int(key string, val int) LogEvent {
	if e.event != nil {
		e.event.Int(key, val)
	}
	return e
}

func (e *logEvent) Int8(key string, val int8) LogEvent {
	if e.event != nil {
		e.event.Int8(key, val)
	}
	return e
}

func (e *logEvent) Int16(key string, val int16) LogEvent {
	if e.event != nil {
		e.event.Int16(key, val)
	}
	return e
}

func (e *logEvent) Int32(key string, val int32) LogEvent {
	if e.event != nil {
		e.event.Int32(key, val)
	}
	return e
}

func (e *logEvent) Int64(key string, val int64) LogEvent {
	if e.event != nil {
		e.event.Int64(key, val)
	}
	return e
}

func (e *logEvent) Uint(key string, val uint) LogEvent {
	if e.event != nil {
		e.event.Uint(key, val)
	}
	return e
}

func (e *logEvent) Uint8(key string, val uint8) LogEvent {
	if e.event != nil {
		e.event.Uint8(key, val)
	}
	return e
}

func (e *logEvent) Uint16(key string, val uint16) LogEvent {
	if e.event != nil {
		e.event.Uint16(key, val)
	}
	return e
}

func (e *logEvent) Uint32(key string, val uint32) LogEvent {
	if e.event != nil {
		e.event.Uint32(key, val)
	}
	return e
}

func (e *logEvent) Uint64(key string, val uint64) LogEvent {
	if e.event != nil {
		e.event.Uint64(key, val)
	}
	return e
}

func (e *logEvent) Float32(key string, val float32) LogEvent {
	if e.event != nil {
		e.event.Float32(key, val)
	}
	return e
}

func (e *logEvent) Float64(key string, val float64) LogEvent {
	if e.event != nil {
		e.event.Float64(key, val)
	}
	return e
}

func (e *logEvent) Bool(key string, val bool) LogEvent {
	if e.event != nil {
		e.event.Bool(key, val)
	}
	return e
}

func (e *logEvent) Bools(key string, vals []bool) LogEvent {
	if e.event != nil {
		e.event.Bools(key, vals)
	}
	return e
}

func (e *logEvent) Time(key string, val time.Time) LogEvent {
	if e.event != nil {
		e.event.Time(key, val)
	}
	return e
}

func (e *logEvent) Dur(key string, val time.Duration) LogEvent {
	if e.event != nil {
		e.event.Dur(key, val)
	}
	return e
}

func (e *logEvent) Err(err error) LogEvent {
	if e.event != nil {
		e.event.Err(err)
	}
	return e
}

func (e *logEvent) AnErr(key string, err error) LogEvent {
	if e.event != nil {
		e.event.AnErr(key, err)
	}
	return e
}

func (e *logEvent) Bytes(key string, val []byte) LogEvent {
	if e.event != nil {
		e.event.Bytes(key, val)
	}
	return e
}

func (e *logEvent) Hex(key string, val []byte) LogEvent {
	if e.event != nil {
		e.event.Hex(key, val)
	}
	return e
}

func (e *logEvent) IPAddr(key string, val net.IP) LogEvent {
	if e.event != nil {
		e.event.IPAddr(key, val)
	}
	return e
}

func (e *logEvent) MACAddr(key string, val net.HardwareAddr) LogEvent {
	if e.event != nil {
		e.event.MACAddr(key, val)
	}
	return e
}

func (e *logEvent) Interface(key string, val interface{}) LogEvent {
	if e.event != nil {
		e.event.Interface(key, val)
	}
	return e
}

// Dict for nested objects
func (e *logEvent) Dict(key string, dict func(LogEvent)) LogEvent {
	if e.event != nil {
		dictEvent := zerolog.Dict()
		dict(newLogEvent(dictEvent))
		e.event.Dict(key, dictEvent)
	}
	return e
}

func (e *logEvent) Msg(msg string) {
	if e.event != nil {
		e.event.Msg(msg)
	}
}

func (e *logEvent) Msgf(format string, v ...interface{}) {
	if e.event != nil {
		e.event.Msgf(format, v...)
	}
}

func (e *logEvent) Send() {
	if e.event != nil {
		e.event.Send()
	}
}

// logContext implements LogContext by wrapping zerolog.Context
type logContext struct {
	context zerolog.Context
	service *Service
}

func (c *logContext) Str(key, val string) LogContext {
	c.context = c.context.Str(key, val)
	return c
}

func (c *logContext) Strs(key string, vals []string) LogContext {
	c.context = c.context.Strs(key, vals)
	return c
}

func (c *logContext) Int(key string, val int) LogContext {
	c.context = c.context.Int(key, val)
	return c
}

func (c *logContext) Int64(key string, val int64) LogContext {
	c.context = c.context.Int64(key, val)
	return c
}

func (c *logContext) Uint(key string, val uint) LogContext {
	c.context = c.context.Uint(key, val)
	return c
}

func (c *logContext) Uint64(key string, val uint64) LogContext {
	c.context = c.context.Uint64(key, val)
	return c
}

func (c *logContext) Float64(key string, val float64) LogContext {
	c.context = c.context.Float64(key, val)
	return c
}

func (c *logContext) Bool(key string, val bool) LogContext {
	c.context = c.context.Bool(key, val)
	return c
}

func (c *logContext) Time(key string, val time.Time) LogContext {
	c.context = c.context.Time(key, val)
	return c
}

func (c *logContext) Err(err error) LogContext {
	c.context = c.context.Err(err)
	return c
}

func (c *logContext) Interface(key string, val interface{}) LogContext {
	c.context = c.context.Interface(key, val)
	return c
}

//	func (c *logContext) Logger() Logger {
//		logger := c.context.Logger()
//		newService := &Service{
//			WorkingDir:    c.service.WorkingDir,
//			AppConfig:     c.service.AppConfig,
//			LoggingConfig: c.service.LoggingConfig,
//			isInitialized: c.service.isInitialized,
//		}
//		newService.logger.Store(&logger)
//		return newService
//	}
func (c *logContext) Logger() Logger {
	logger := c.context.Logger()
	c.service.logger.Store(&logger)
	return c.service
}

package util

import (
	"time"

	"github.com/getsentry/sentry-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// 重写zap core，支持将err msg 输出到sentry

type sentryCoreConfig struct {
	level        zapcore.Level
	flushTimeout time.Duration
	tags         map[string]string // 自定义tag
	hub          *sentry.Hub
}

type sentryCore struct {
	zapcore.LevelEnabler

	client       *sentry.Client
	cfg          *sentryCoreConfig
	fields       map[string]interface{}
	flushTimeout time.Duration
}

func sentryLevel(zapLevel zapcore.Level) sentry.Level {
	switch zapLevel {
	case zapcore.DebugLevel:
		return sentry.LevelDebug
	case zapcore.InfoLevel:
		return sentry.LevelInfo
	case zapcore.WarnLevel:
		return sentry.LevelWarning
	case zapcore.ErrorLevel:
		return sentry.LevelError
	case zapcore.DPanicLevel:
		return sentry.LevelFatal
	case zapcore.PanicLevel:
		return sentry.LevelFatal
	case zapcore.FatalLevel:
		return sentry.LevelFatal
	default:
		return sentry.LevelFatal
	}
}

func (c *sentryCore) with(fields []zapcore.Field) *sentryCore {

	m := make(map[string]interface{}, len(c.fields))
	for k, v := range c.fields {
		m[k] = v
	}
	enc := zapcore.NewMapObjectEncoder()
	for i := range fields {
		fields[i].AddTo(enc)
	}

	for k, v := range enc.Fields {
		m[k] = v
	}

	return &sentryCore{client: c.client, cfg: c.cfg, fields: m}
}

func (c *sentryCore) With(fields []zapcore.Field) zapcore.Core {
	return c.with(fields)
}

func (c *sentryCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.cfg.level.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}
	return ce
}

func (c *sentryCore) Sync() error {
	c.client.Flush(c.flushTimeout)
	return nil
}

func (c *sentryCore) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	clone := c.with(fields)

	event := sentry.NewEvent()
	event.Message = ent.Message
	event.Timestamp = ent.Time
	event.Level = sentryLevel(ent.Level)
	event.Platform = "demo"
	event.Extra = clone.fields
	event.Tags = c.cfg.tags

	trace := sentry.NewStacktrace()
	if trace != nil {
		event.Exception = []sentry.Exception{{
			Type:       ent.Message,
			Value:      ent.Caller.TrimmedPath(),
			Stacktrace: trace,
		}}
	}

	_ = c.client.CaptureEvent(event, nil, c.cfg.hub.Scope())
	c.client.Flush(c.flushTimeout)
	return nil
}

func NewSentryCore(cfg *sentryCoreConfig, client *sentry.Client) zapcore.Core {
	core := sentryCore{
		client:       client,
		cfg:          cfg,
		LevelEnabler: cfg.level,
		fields:       make(map[string]interface{}),
		flushTimeout: 3 * time.Second,
	}

	if cfg.flushTimeout > 0 {
		core.flushTimeout = cfg.flushTimeout
	}

	return &core
}

func UseSentryLog(sentryClient *sentry.Client) {

	if sentryClient == nil {
		Logger.Info("没有使用SentryCore，error消息不会发送到sentry")
		return
	}

	cfg := sentryCoreConfig{
		tags:         map[string]string{},
		flushTimeout: 2 * time.Second,
		level:        zap.ErrorLevel,
		hub:          sentry.CurrentHub(),
	}
	sCore := NewSentryCore(&cfg, sentryClient)

	Logger, _ = zap.NewProduction(
		zap.WrapCore(func(c zapcore.Core) zapcore.Core {
			return zapcore.NewTee(c, sCore)
		}),
	)
	Logger.Info("使用SentryCore，error消息会发送到sentry")
}

func GetDefaultLogger() *zap.Logger {
	logger, _ := zap.NewProduction()
	return logger
}

var Logger = GetDefaultLogger()

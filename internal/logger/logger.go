package logger

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gopkg.in/natefinch/lumberjack.v2"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var (
	GlobalLevelVar     = new(slog.LevelVar)
	AppRollingLogger   *lumberjack.Logger
	ErrorRollingLogger *lumberjack.Logger

	permanentMu        sync.Mutex
	permanentLogWriter *os.File
	permanentLevel     = "error"

	ringMu       sync.RWMutex
	recentLogs   []string
	recentErrors []string
	maxRingSize  = 120
)

type Config struct {
	Level          string
	Dir            string
	MaxSizeMB      int
	MaxBackups     int
	MaxAgeDays     int
	Compress       bool
	PermanentLevel string
}

type MultiLevelHandler struct {
	baseHandler slog.Handler
}

func (h *MultiLevelHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.baseHandler.Enabled(ctx, level)
}

func (h *MultiLevelHandler) Handle(ctx context.Context, r slog.Record) error {
	_ = h.baseHandler.Handle(ctx, r)

	formattedTime := r.Time.Format("2006-01-02 15:04:05")
	var attrs strings.Builder
	r.Attrs(func(a slog.Attr) bool {
		attrs.WriteString(fmt.Sprintf(" %s=%v", a.Key, a.Value))
		return true
	})
	line := fmt.Sprintf("[%s] [%s] %s%s", formattedTime, r.Level.String(), r.Message, attrs.String())

	ringMu.Lock()
	recentLogs = append(recentLogs, line)
	if len(recentLogs) > maxRingSize {
		recentLogs = recentLogs[len(recentLogs)-maxRingSize:]
	}
	if r.Level >= slog.LevelError {
		recentErrors = append(recentErrors, line)
		if len(recentErrors) > maxRingSize {
			recentErrors = recentErrors[len(recentErrors)-maxRingSize:]
		}
	}
	ringMu.Unlock()

	if r.Level >= slog.LevelError && ErrorRollingLogger != nil {
		_, _ = fmt.Fprintln(ErrorRollingLogger, line)
	}

	permanentMu.Lock()
	defer permanentMu.Unlock()
	if permanentLogWriter != nil && permanentLevel != "none" {
		targetLvl := slog.LevelError
		if permanentLevel == "warn" {
			targetLvl = slog.LevelWarn
		}
		if r.Level >= targetLvl {
			_, _ = fmt.Fprintln(permanentLogWriter, line)
		}
	}

	return nil
}

func (h *MultiLevelHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &MultiLevelHandler{baseHandler: h.baseHandler.WithAttrs(attrs)}
}

func (h *MultiLevelHandler) WithGroup(name string) slog.Handler {
	return &MultiLevelHandler{baseHandler: h.baseHandler.WithGroup(name)}
}

func InitLogger(cfg Config) error {
	if cfg.Dir == "" {
		cfg.Dir = "logs"
	}
	if err := os.MkdirAll(cfg.Dir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	SetLevel(cfg.Level)
	SetPermanentLevel(cfg.PermanentLevel)

	if cfg.MaxSizeMB <= 0 {
		cfg.MaxSizeMB = 20
	}
	if cfg.MaxBackups <= 0 {
		cfg.MaxBackups = 10
	}
	if cfg.MaxAgeDays <= 0 {
		cfg.MaxAgeDays = 30
	}

	AppRollingLogger = &lumberjack.Logger{
		Filename:   filepath.Join(cfg.Dir, "app.log"),
		MaxSize:    cfg.MaxSizeMB,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAgeDays,
		Compress:   cfg.Compress,
	}

	ErrorRollingLogger = &lumberjack.Logger{
		Filename:   filepath.Join(cfg.Dir, "error.log"),
		MaxSize:    cfg.MaxSizeMB,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAgeDays,
		Compress:   cfg.Compress,
	}

	permanentMu.Lock()
	if permanentLogWriter != nil {
		_ = permanentLogWriter.Close()
	}
	pw, err := os.OpenFile(filepath.Join(cfg.Dir, "permanent_error.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		permanentLogWriter = pw
	}
	permanentMu.Unlock()

	multiWriter := io.MultiWriter(os.Stdout, AppRollingLogger)
	baseHandler := slog.NewTextHandler(multiWriter, &slog.HandlerOptions{
		Level:     GlobalLevelVar,
		AddSource: false,
	})
	slog.SetDefault(slog.New(&MultiLevelHandler{baseHandler: baseHandler}))
	return nil
}

func SetLevel(levelStr string) {
	switch strings.ToLower(strings.TrimSpace(levelStr)) {
	case "debug":
		GlobalLevelVar.Set(slog.LevelDebug)
	case "info":
		GlobalLevelVar.Set(slog.LevelInfo)
	case "warn", "warning":
		GlobalLevelVar.Set(slog.LevelWarn)
	case "error":
		GlobalLevelVar.Set(slog.LevelError)
	default:
		GlobalLevelVar.Set(slog.LevelInfo)
	}
}

func GetLevel() string {
	switch GlobalLevelVar.Level() {
	case slog.LevelDebug:
		return "DEBUG"
	case slog.LevelInfo:
		return "INFO"
	case slog.LevelWarn:
		return "WARN"
	case slog.LevelError:
		return "ERROR"
	default:
		return "INFO"
	}
}

func SetPermanentLevel(lvl string) {
	permanentMu.Lock()
	defer permanentMu.Unlock()
	cleaned := strings.ToLower(strings.TrimSpace(lvl))
	if cleaned == "" {
		cleaned = "error"
	}
	permanentLevel = cleaned
}

func GetPermanentLevel() string {
	permanentMu.Lock()
	defer permanentMu.Unlock()
	return strings.ToUpper(permanentLevel)
}

func GetRecentLogs(count int) []string {
	ringMu.RLock()
	defer ringMu.RUnlock()
	if count <= 0 || count > len(recentLogs) {
		count = len(recentLogs)
	}
	res := make([]string, count)
	copy(res, recentLogs[len(recentLogs)-count:])
	return res
}

func GetRecentErrors(count int) []string {
	ringMu.RLock()
	defer ringMu.RUnlock()
	if count <= 0 || count > len(recentErrors) {
		count = len(recentErrors)
	}
	res := make([]string, count)
	copy(res, recentErrors[len(recentErrors)-count:])
	return res
}

type GormLogger struct {
	SlowThreshold time.Duration
}

func NewGormLogger() *GormLogger {
	return &GormLogger{SlowThreshold: 250 * time.Millisecond}
}

func (g *GormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	return g
}

func (g *GormLogger) Info(ctx context.Context, msg string, data ...any) {
	slog.Info(fmt.Sprintf(msg, data...))
}

func (g *GormLogger) Warn(ctx context.Context, msg string, data ...any) {
	slog.Warn(fmt.Sprintf(msg, data...))
}

func (g *GormLogger) Error(ctx context.Context, msg string, data ...any) {
	slog.Error(fmt.Sprintf(msg, data...))
}

func (g *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		sql, rows := fc()
		slog.Error("Database query failed", "error", err, "elapsed", elapsed, "rows", rows, "sql", sql)
		return
	}
	if g.SlowThreshold > 0 && elapsed > g.SlowThreshold {
		sql, rows := fc()
		slog.Warn("Slow database query", "elapsed", elapsed, "rows", rows, "sql", sql)
		return
	}
	if GlobalLevelVar.Level() <= slog.LevelDebug {
		sql, rows := fc()
		slog.Debug("SQL query", "elapsed", elapsed, "rows", rows, "sql", sql)
	}
}

func GinLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		rawQuery := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		isProbe := strings.HasPrefix(path, "/.well-known") || path == "/favicon.ico"

		switch {
		case statusCode >= http.StatusInternalServerError:
			slog.Error("HTTP request error", "status", statusCode, "method", method, "path", path, "query", rawQuery, "ip", clientIP, "latency", latency)
		case statusCode >= http.StatusBadRequest && !isProbe:
			slog.Warn("HTTP client warning", "status", statusCode, "method", method, "path", path, "ip", clientIP, "latency", latency)
		case isProbe:
			slog.Debug("HTTP probe", "status", statusCode, "method", method, "path", path, "ip", clientIP, "latency", latency)
		default:
			slog.Debug("HTTP request", "status", statusCode, "method", method, "path", path, "ip", clientIP, "latency", latency)
		}
	}
}

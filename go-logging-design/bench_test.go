package benchmarks

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/rs/zerolog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// =============================================================
// E3: Disabled Level Cost — 证明 disabled log 不是零成本
// =============================================================

// slog: disabled level
func BenchmarkSlogDisabledLevel(b *testing.B) {
	h := slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelWarn})
	logger := slog.New(h)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("this will be discarded", "key1", "value1", "key2", 42, "key3", true)
	}
}

// zap: disabled level
func BenchmarkZapDisabledLevel(b *testing.B) {
	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	cfg.OutputPaths = []string{"/dev/null"}
	logger, _ := cfg.Build()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("this will be discarded", zap.String("key1", "value1"), zap.Int("key2", 42), zap.Bool("key3", true))
	}
}

// zap sugar: disabled level (展示 sugar 的额外开销)
func BenchmarkZapSugarDisabledLevel(b *testing.B) {
	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	cfg.OutputPaths = []string{"/dev/null"}
	logger, _ := cfg.Build()
	sugar := logger.Sugar()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sugar.Infow("this will be discarded", "key1", "value1", "key2", 42, "key3", true)
	}
}

// zerolog: disabled level
func BenchmarkZerologDisabledLevel(b *testing.B) {
	logger := zerolog.New(io.Discard).Level(zerolog.WarnLevel)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info().Str("key1", "value1").Int("key2", 42).Bool("key3", true).Msg("this will be discarded")
	}
}

// =============================================================
// E2: Serialization Strategy — JSON vs Text vs 预分配
// =============================================================

// slog JSON handler
func BenchmarkSlogJSON(b *testing.B) {
	h := slog.NewJSONHandler(io.Discard, nil)
	logger := slog.New(h)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled", "method", "GET", "path", "/api/users", "status", 200, "duration_ms", 42)
	}
}

// slog Text handler (logfmt-like)
func BenchmarkSlogText(b *testing.B) {
	h := slog.NewTextHandler(io.Discard, nil)
	logger := slog.New(h)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled", "method", "GET", "path", "/api/users", "status", 200, "duration_ms", 42)
	}
}

// zap JSON
func BenchmarkZapJSON(b *testing.B) {
	enc := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	core := zapcore.NewCore(enc, zapcore.AddSync(io.Discard), zap.InfoLevel)
	logger := zap.New(core)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled", zap.String("method", "GET"), zap.String("path", "/api/users"), zap.Int("status", 200), zap.Int("duration_ms", 42))
	}
}

// zap Console (text)
func BenchmarkZapConsole(b *testing.B) {
	enc := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	core := zapcore.NewCore(enc, zapcore.AddSync(io.Discard), zap.InfoLevel)
	logger := zap.New(core)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled", zap.String("method", "GET"), zap.String("path", "/api/users"), zap.Int("status", 200), zap.Int("duration_ms", 42))
	}
}

// zerolog JSON
func BenchmarkZerologJSON(b *testing.B) {
	logger := zerolog.New(io.Discard)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info().Str("method", "GET").Str("path", "/api/users").Int("status", 200).Int("duration_ms", 42).Msg("request handled")
	}
}

// =============================================================
// E4: Field Binding — 每次传 vs With 预绑定 vs Context
// =============================================================

// slog: 每次传 field
func BenchmarkSlogFieldsEveryCall(b *testing.B) {
	h := slog.NewJSONHandler(io.Discard, nil)
	logger := slog.New(h)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request", "service", "user-api", "instance", "pod-1", "region", "us-east-1", "method", "GET", "status", 200)
	}
}

// slog: With 预绑定公共字段
func BenchmarkSlogFieldsWithPrebound(b *testing.B) {
	h := slog.NewJSONHandler(io.Discard, nil)
	logger := slog.New(h).With("service", "user-api", "instance", "pod-1", "region", "us-east-1")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request", "method", "GET", "status", 200)
	}
}

// slog: context 携带
func BenchmarkSlogFieldsContext(b *testing.B) {
	h := slog.NewJSONHandler(io.Discard, nil)
	logger := slog.New(h)
	ctx := context.Background()
	// 模拟 context 中携带 logger
	loggerWithCtx := logger.With("service", "user-api", "instance", "pod-1", "region", "us-east-1")
	_ = ctx
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		loggerWithCtx.Info("request", "method", "GET", "status", 200)
	}
}

// zap: 每次传 field
func BenchmarkZapFieldsEveryCall(b *testing.B) {
	enc := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	core := zapcore.NewCore(enc, zapcore.AddSync(io.Discard), zap.InfoLevel)
	logger := zap.New(core)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request", zap.String("service", "user-api"), zap.String("instance", "pod-1"), zap.String("region", "us-east-1"), zap.String("method", "GET"), zap.Int("status", 200))
	}
}

// zap: With 预绑定
func BenchmarkZapFieldsWithPrebound(b *testing.B) {
	enc := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	core := zapcore.NewCore(enc, zapcore.AddSync(io.Discard), zap.InfoLevel)
	logger := zap.New(core).With(zap.String("service", "user-api"), zap.String("instance", "pod-1"), zap.String("region", "us-east-1"))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request", zap.String("method", "GET"), zap.Int("status", 200))
	}
}

// =============================================================
// E1: Sync vs Async — 同步 vs 异步写入（用 buffered writer 模拟）
// =============================================================

// 同步写入（直接 io.Discard 代表最快的同步写，实际磁盘更慢）
func BenchmarkZapSyncWrite(b *testing.B) {
	enc := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	core := zapcore.NewCore(enc, zapcore.AddSync(io.Discard), zap.InfoLevel)
	logger := zap.New(core)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled", zap.String("method", "GET"), zap.String("path", "/api/users"), zap.Int("status", 200))
	}
}

// 异步写入（使用 zapcore.NewCore + buffered WriteSyncer）
func BenchmarkZapBufferedWrite(b *testing.B) {
	enc := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	ws := &zapcore.BufferedWriteSyncer{WS: zapcore.AddSync(io.Discard), Size: 4096}
	core := zapcore.NewCore(enc, ws, zap.InfoLevel)
	logger := zap.New(core)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled", zap.String("method", "GET"), zap.String("path", "/api/users"), zap.Int("status", 200))
	}
	ws.Stop()
}

// =============================================================
// E6: 综合对比 — 同库不同配置 vs 不同库默认配置
// =============================================================

// zap 默认配置（未优化）
func BenchmarkZapDefault(b *testing.B) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	// 重定向到 discard (需要 redirect)
	enc := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	core := zapcore.NewCore(enc, zapcore.AddSync(io.Discard), zap.InfoLevel)
	l := zap.New(core, zap.AddCaller())
	b.ResetTimer()
	_ = logger
	for i := 0; i < b.N; i++ {
		l.Info("request handled", zap.String("method", "GET"), zap.String("path", "/api/users"), zap.Int("status", 200))
	}
}

// zap 优化配置（预绑定+buffered+无 caller）
func BenchmarkZapOptimized(b *testing.B) {
	enc := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	ws := &zapcore.BufferedWriteSyncer{WS: zapcore.AddSync(io.Discard), Size: 4096}
	core := zapcore.NewCore(enc, ws, zap.InfoLevel)
	logger := zap.New(core).With(zap.String("service", "user-api"), zap.String("instance", "pod-1"))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled", zap.String("method", "GET"), zap.Int("status", 200))
	}
	ws.Stop()
}

// slog 默认配置
func BenchmarkSlogDefault(b *testing.B) {
	h := slog.NewJSONHandler(io.Discard, nil)
	logger := slog.New(h)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled", "method", "GET", "path", "/api/users", "status", 200)
	}
}

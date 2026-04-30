package benchmarks

import (
	"os"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// E1 补充：真实文件 I/O 场景下的同步 vs buffered
func BenchmarkZapSyncFile(b *testing.B) {
	f, err := os.CreateTemp("", "bench-sync-*.log")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	enc := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	core := zapcore.NewCore(enc, zapcore.AddSync(f), zap.InfoLevel)
	logger := zap.New(core)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled", zap.String("method", "GET"), zap.String("path", "/api/users"), zap.Int("status", 200))
	}
}

func BenchmarkZapBufferedFile(b *testing.B) {
	f, err := os.CreateTemp("", "bench-buffered-*.log")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	enc := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	ws := &zapcore.BufferedWriteSyncer{WS: zapcore.AddSync(f), Size: 256 * 1024}
	core := zapcore.NewCore(enc, ws, zap.InfoLevel)
	logger := zap.New(core)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled", zap.String("method", "GET"), zap.String("path", "/api/users"), zap.Int("status", 200))
	}
	ws.Stop()
}

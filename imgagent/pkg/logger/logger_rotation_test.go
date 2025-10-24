package logger

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestLogRotation 测试日志轮转功能
func TestLogRotation(t *testing.T) {
	// 创建临时测试目录
	tmpDir := filepath.Join(os.TempDir(), "test_logger_rotation")
	err := os.MkdirAll(tmpDir, 0755)
	require.NoError(t, err, "Failed to create temp dir")
	defer os.RemoveAll(tmpDir)

	// 配置日志轮转（使用小的文件大小以便测试）
	logFile := filepath.Join(tmpDir, "test.log")
	conf := Config{
		Level: "debug",
		File:  logFile,
		Rotation: RotationConfig{
			MaxSize:    1, // 1MB
			MaxBackups: 3, // 保留3个备份
			MaxAge:     7, // 保留7天
		},
	}

	// 创建 logger
	logger, err := New(conf)
	require.NoError(t, err, "Failed to create logger")
	defer logger.Sync()

	// 写入日志数据
	for i := 0; i < 1000; i++ {
		logger.Info("This is a test log message",
			zap.Int("iteration", i),
			zap.String("timestamp", time.Now().Format(time.RFC3339)),
			zap.String("data", "Some additional data to make the log entry larger"),
		)
	}

	// 检查日志文件是否存在
	assert.FileExists(t, logFile, "Log file should exist")

	t.Logf("Log file created successfully: %s", logFile)
}

// TestLogRotationDefaults 测试默认配置
func TestLogRotationDefaults(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "test_logger_defaults")
	err := os.MkdirAll(tmpDir, 0755)
	require.NoError(t, err, "Failed to create temp dir")
	defer os.RemoveAll(tmpDir)

	logFile := filepath.Join(tmpDir, "test_defaults.log")
	conf := Config{
		Level: "info",
		File:  logFile,
		// Rotation 使用默认配置（零值）
	}

	logger, err := New(conf)
	require.NoError(t, err, "Failed to create logger")
	require.NotNil(t, logger, "Logger should not be nil")
	defer logger.Sync()

	// 写入测试日志验证默认配置能正常工作
	logger.Info("Test log with default rotation settings")
	logger.Warn("Another test message")

	// 验证日志文件被成功创建
	assert.FileExists(t, logFile, "Log file should exist with default rotation config")

	// 验证日志文件有内容
	fileInfo, err := os.Stat(logFile)
	require.NoError(t, err, "Should be able to stat log file")
	assert.Greater(t, fileInfo.Size(), int64(0), "Log file should not be empty")
}

// TestStdoutLogging 测试标准输出（不使用日志轮转）
func TestStdoutLogging(t *testing.T) {
	conf := Config{
		Level: "info",
		File:  "stdout", // 或者 ""
	}

	logger, err := New(conf)
	require.NoError(t, err, "Failed to create logger")
	require.NotNil(t, logger, "Logger should not be nil")
	defer logger.Sync()

	logger.Info("This message should go to stdout")
	logger.Debug("This debug message should not appear (level is info)")
	logger.Warn("This warning should appear")
}

// BenchmarkLogRotation 性能基准测试
func BenchmarkLogRotation(b *testing.B) {
	tmpDir := filepath.Join(os.TempDir(), "bench_logger_rotation")
	err := os.MkdirAll(tmpDir, 0755)
	require.NoError(b, err, "Failed to create temp dir")
	defer os.RemoveAll(tmpDir)

	logFile := filepath.Join(tmpDir, "bench.log")
	conf := Config{
		Level: "info",
		File:  logFile,
		Rotation: RotationConfig{
			MaxSize:    100,
			MaxBackups: 5,
			MaxAge:     30,
		},
	}

	logger, err := New(conf)
	require.NoError(b, err, "Failed to create logger")
	defer logger.Sync()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("Benchmark log message",
			zap.Int("iteration", i),
			zap.String("timestamp", time.Now().Format(time.RFC3339)),
		)
	}
}

// ExampleNew 演示如何创建带日志轮转的 logger
func ExampleNew() {
	conf := Config{
		Level: "info",
		File:  "logs/app.log",
		Rotation: RotationConfig{
			MaxSize:    100, // 单个文件最大 100MB
			MaxBackups: 10,  // 保留最多 10 个备份文件
			MaxAge:     30,  // 保留最多 30 天的日志
		},
	}

	logger, err := New(conf)
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	logger.Info("Application started",
		zap.String("version", "1.0.0"),
		zap.Int("port", 8080),
	)
}

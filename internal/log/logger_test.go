package log

import (
	"testing"

	"go.uber.org/zap/zapcore"
)

func TestParseLevel(t *testing.T) {
	cases := map[string]zapcore.Level{
		"debug":   zapcore.DebugLevel,
		"DEBUG":   zapcore.DebugLevel,
		"warn":    zapcore.WarnLevel,
		"warning": zapcore.WarnLevel,
		"error":   zapcore.ErrorLevel,
		"info":    zapcore.InfoLevel,
		"":        zapcore.InfoLevel,
		"bogus":   zapcore.InfoLevel,
	}
	for input, want := range cases {
		if got := parseLevel(input); got != want {
			t.Fatalf("parseLevel(%q) = %v, want %v", input, got, want)
		}
	}
}

func TestNewLogger_AppliesConfiguredLevel(t *testing.T) {
	logger, err := NewLogger("error")
	if err != nil {
		t.Fatal(err)
	}
	core := logger.Desugar().Core()
	if core.Enabled(zapcore.InfoLevel) {
		t.Fatal("expected info-level logs to be disabled when configured level is error")
	}
	if !core.Enabled(zapcore.ErrorLevel) {
		t.Fatal("expected error-level logs to be enabled")
	}
}

package log

import (
	"os"
	"os/exec"
	"testing"

	"go.uber.org/zap"
)

func TestInit(t *testing.T) {
	previous := sugarLog
	t.Cleanup(func() {
		sugarLog = previous
	})

	Init()

	if sugarLog == nil {
		t.Fatal("expected logger to be initialized")
	}
}

func TestLogFunctions(t *testing.T) {
	previous := sugarLog
	sugarLog = zap.NewNop().Sugar()
	t.Cleanup(func() {
		sugarLog = previous
	})

	tests := []struct {
		name string
		call func()
	}{
		{name: "info", call: func() { Info("info") }},
		{name: "infof", call: func() { Infof("info %s", "formatted") }},
		{name: "error", call: func() { Error("error") }},
		{name: "errorf", call: func() { Errorf("error %s", "formatted") }},
		{name: "sync", call: func() { Sync() }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.call()
		})
	}
}

func TestLogPanicFunctions(t *testing.T) {
	previous := sugarLog
	sugarLog = zap.NewNop().Sugar()
	t.Cleanup(func() {
		sugarLog = previous
	})

	tests := []struct {
		name string
		call func()
	}{
		{name: "panic", call: func() { Panic("panic") }},
		{name: "panicf", call: func() { Panicf("panic %s", "formatted") }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			func() {
				defer func() {
					if recover() == nil {
						t.Error("expected panic")
					}
				}()

				tt.call()
			}()
		})
	}
}

func TestLogFatalFunctions(t *testing.T) {
	tests := []struct {
		name string
		mode string
	}{
		{name: "fatal", mode: "fatal"},
		{name: "fatalf", mode: "fatalf"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=^TestLogFatalHelper$")
			cmd.Env = append(
				os.Environ(),
				"LOG_FATAL_HELPER=1",
				"LOG_FATAL_MODE="+tt.mode,
			)

			err := cmd.Run()
			exitErr, ok := err.(*exec.ExitError)
			if !ok {
				t.Fatalf("expected subprocess to exit, got error: %v", err)
			}
			if exitErr.ExitCode() != 1 {
				t.Fatalf("exit code = %d, want 1", exitErr.ExitCode())
			}
		})
	}
}

func TestLogFatalHelper(t *testing.T) {
	if os.Getenv("LOG_FATAL_HELPER") != "1" {
		t.Skip("helper process")
	}

	sugarLog = zap.NewNop().Sugar()

	switch os.Getenv("LOG_FATAL_MODE") {
	case "fatalf":
		Fatalf("fatal %s", "formatted")
	default:
		Fatal("fatal")
	}
}

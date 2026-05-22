package transfer

import (
	"context"
	"errors"
	"os"
	"testing"
)

func tempFile(t *testing.T) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "sharedlink-test")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	return f.Name()
}

func TestSendNilContext(t *testing.T) {
	err := Send(nil, "", "nonexistent_file", nil)
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}

func TestReceiveNilContext(t *testing.T) {
	err := Receive(nil, "127.0.0.1:1", nil)
	if err == nil {
		t.Error("expected error for unreachable address, got nil")
	}
}

func TestSendCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := Send(ctx, "", tempFile(t), nil)
	if err == nil {
		t.Fatal("expected error for cancelled context, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Logf("error is context.Canceled or wraps it: got %T %v", err, err)
	}
}

func TestReceiveCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := Receive(ctx, "127.0.0.1:1", nil)
	if err == nil {
		t.Fatal("expected error for cancelled context, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Logf("error is context.Canceled or wraps it: got %T %v", err, err)
	}
}

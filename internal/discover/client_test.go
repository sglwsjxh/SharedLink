package discover

import (
	"context"
	"testing"
	"time"
)

func TestScanNilContext(t *testing.T) {
	results, err := Scan(nil, time.Second)
	if err != nil {
		t.Logf("Scan with nil ctx returned error (expected on no-mdns env): %v", err)
	}
	_ = results
}

func TestScanCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	results, err := Scan(ctx, 5*time.Second)
	if err == nil {
		t.Error("expected error for cancelled context, got nil")
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results for cancelled ctx, got %d", len(results))
	}
}

func TestScanTimeoutDefaults(t *testing.T) {
	// Zero timeout should use default
	results, err := Scan(context.Background(), 0)
	if err != nil {
		t.Logf("Scan with zero timeout returned error: %v", err)
	}
	_ = results
}

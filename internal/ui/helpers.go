package ui

import (
	"fmt"
	"math"
	"time"
)

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func formatSpeed(bytes int64, startTime time.Time) string {
	if bytes == 0 {
		return "0 B/s"
	}
	elapsed := time.Since(startTime).Seconds()
	if elapsed < 0.01 {
		return "0 B/s"
	}
	bps := float64(bytes) / elapsed
	const unit = 1024
	if bps < unit {
		return fmt.Sprintf("%.0f B/s", bps)
	}
	div, exp := unit, 0
	for n := bps / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB/s", bps/float64(div), "KMGTPE"[exp])
}

func formatETA(sent, total int64, startTime time.Time) string {
	if sent == 0 {
		return "calculating..."
	}
	elapsed := time.Since(startTime).Seconds()
	if elapsed < 0.5 {
		return "calculating..."
	}
	bps := float64(sent) / elapsed
	remaining := float64(total-sent) / bps
	if remaining > 3600 {
		return fmt.Sprintf("%dh%dm", int(remaining/3600), int(math.Mod(remaining, 3600))/60)
	}
	if remaining > 60 {
		return fmt.Sprintf("%dm%ds", int(remaining/60), int(math.Mod(remaining, 60)))
	}
	return fmt.Sprintf("%ds", int(remaining))
}

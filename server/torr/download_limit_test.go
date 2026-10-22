package torr

import (
	"testing"

	"github.com/anacrolix/torrent"
	"golang.org/x/time/rate"

	"server/torr/utils"
)

func TestLiveDownloadLimitKeepsClientLimiter(t *testing.T) {
	limiter := utils.Limit(0)
	bt := &BTServer{
		config:          &torrent.ClientConfig{DownloadRateLimiter: limiter},
		downloadLimiter: limiter,
	}

	bt.SetLiveDownloadLimit(5120, false)
	if bt.config.DownloadRateLimiter != limiter {
		t.Fatal("live update replaced the limiter used by the torrent client")
	}
	if got := limiter.Limit(); got != rate.Limit(5120*1024) {
		t.Fatalf("limited rate = %v, want %d bytes/s", got, 5120*1024)
	}
	if got := limiter.Burst(); got < 16*1024 {
		t.Fatalf("burst = %d, cannot fit a torrent chunk", got)
	}

	bt.SetLiveDownloadLimit(0, true)
	if got := limiter.Limit(); got != rate.Inf {
		t.Fatalf("temporary unlimited rate = %v, want unlimited", got)
	}
	if override, ok := bt.DownloadLimitOverride(); !ok || override != 0 {
		t.Fatalf("temporary override = (%d, %t), want (0, true)", override, ok)
	}

	bt.SetLiveDownloadLimit(5120, false)
	if _, ok := bt.DownloadLimitOverride(); ok {
		t.Fatal("restoring the saved limit left the override active")
	}
	if got := limiter.Limit(); got != rate.Limit(5120*1024) {
		t.Fatalf("restored rate = %v, want %d bytes/s", got, 5120*1024)
	}
}

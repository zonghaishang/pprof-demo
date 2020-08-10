package handler

import (
	"net/http"
	"os"
	"path/filepath"
	"pprof-demo/cmd/stats"
	"strings"
	"time"
)

var _hostName = getHost()

// WithStats wraps handlers with stats reporting. It tracks metrics such
// as the number of requests per endpoint, the latency, etc.
func WithStats(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		tags := getStatsTags(r)
		stats.IncCounter("handler.received", tags, 1)

		h(w, r)

		stats.RecordTimer("handler.latency", tags, time.Since(start))
	}
}

func getHost() string {
	host, err := os.Hostname()
	if err != nil {
		return ""
	}

	if idx := strings.IndexByte(host, '.'); idx > 0 {
		host = host[:idx]
	}
	return host
}

func getStatsTags(r *http.Request) map[string]string {
	stats := map[string]string{
		"browser":  "chrome",
		"os":       "mac OS",
		"endpoint": filepath.Base(r.URL.Path),
	}

	// todo case 1:
	// Minimize system calls.
	_hostName := getHost()

	if _hostName != "" {
		stats["host"] = _hostName
	}
	return stats
}

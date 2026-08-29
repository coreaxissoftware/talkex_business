// Package metrics exposes a minimal Prometheus-format /metrics endpoint
// so ops can scrape process + business KPIs without pulling in the full
// prometheus/client_golang dependency graph.
//
// Counters increment via Inc(name); gauges via Set(name, v). The output
// format follows Prometheus's text exposition v0.0.4 spec:
//   # HELP name description
//   # TYPE name counter|gauge
//   name{label="value"} 42
package metrics

import (
	"fmt"
	"io"
	"net/http"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/gin-gonic/gin"
)

type counter struct {
	value uint64
	help  string
}

var (
	countersMu sync.RWMutex
	counters   = map[string]*counter{}
)

// Inc increments the named counter by 1. Auto-registers on first use.
func Inc(name string) { Add(name, 1) }

// Add adds n to the named counter (creating it if needed).
func Add(name string, n uint64) {
	countersMu.RLock()
	c, ok := counters[name]
	countersMu.RUnlock()
	if !ok {
		countersMu.Lock()
		c = counters[name]
		if c == nil {
			c = &counter{help: name}
			counters[name] = c
		}
		countersMu.Unlock()
	}
	atomic.AddUint64(&c.value, n)
}

// Describe sets the HELP text for a metric — call once at startup so
// scraped output includes documentation.
func Describe(name, help string) {
	countersMu.Lock()
	defer countersMu.Unlock()
	c, ok := counters[name]
	if !ok {
		c = &counter{}
		counters[name] = c
	}
	c.help = help
}

// RegisterRoutes wires GET /metrics. Public — no auth so Prometheus
// (which typically runs inside the cluster) can scrape without a JWT.
func RegisterRoutes(r *gin.Engine) {
	r.GET("/metrics", handleMetrics)
}

func handleMetrics(c *gin.Context) {
	c.Header("Content-Type", "text/plain; version=0.0.4")
	writeCounters(c.Writer)
	writeRuntime(c.Writer)
}

func writeCounters(w io.Writer) {
	countersMu.RLock()
	names := make([]string, 0, len(counters))
	for n := range counters {
		names = append(names, n)
	}
	countersMu.RUnlock()
	sort.Strings(names)
	for _, n := range names {
		countersMu.RLock()
		c := counters[n]
		countersMu.RUnlock()
		if c == nil {
			continue
		}
		help := c.help
		if help == "" {
			help = n
		}
		safeName := strings.ReplaceAll(n, "-", "_")
		fmt.Fprintf(w, "# HELP %s %s\n", safeName, help)
		fmt.Fprintf(w, "# TYPE %s counter\n", safeName)
		fmt.Fprintf(w, "%s %d\n", safeName, atomic.LoadUint64(&c.value))
	}
}

func writeRuntime(w io.Writer) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Fprintln(w, "# HELP go_goroutines Current goroutines")
	fmt.Fprintln(w, "# TYPE go_goroutines gauge")
	fmt.Fprintf(w, "go_goroutines %d\n", runtime.NumGoroutine())
	fmt.Fprintln(w, "# HELP go_memstats_alloc_bytes Bytes allocated and in use")
	fmt.Fprintln(w, "# TYPE go_memstats_alloc_bytes gauge")
	fmt.Fprintf(w, "go_memstats_alloc_bytes %d\n", m.Alloc)
	fmt.Fprintln(w, "# HELP go_memstats_sys_bytes Bytes obtained from OS")
	fmt.Fprintln(w, "# TYPE go_memstats_sys_bytes gauge")
	fmt.Fprintf(w, "go_memstats_sys_bytes %d\n", m.Sys)
}

// Middleware records total request count + per-status-class buckets so
// dashboards can graph http_requests_total{class="2xx"} etc.
func Middleware() gin.HandlerFunc {
	Describe("http_requests_total", "Total HTTP requests handled")
	Describe("http_requests_2xx_total", "HTTP 2xx responses")
	Describe("http_requests_4xx_total", "HTTP 4xx responses")
	Describe("http_requests_5xx_total", "HTTP 5xx responses")
	return func(c *gin.Context) {
		c.Next()
		Inc("http_requests_total")
		s := c.Writer.Status()
		switch {
		case s >= 500:
			Inc("http_requests_5xx_total")
		case s >= 400:
			Inc("http_requests_4xx_total")
		case s >= 200:
			Inc("http_requests_2xx_total")
		}
	}
}

// Handler is exported for tests that want to hit metrics without a router.
func Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		writeCounters(w)
		writeRuntime(w)
	}
}

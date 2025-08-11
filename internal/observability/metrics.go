package observability

import (
    "net/http"
    "sync/atomic"
)

type Metrics struct {
    totalRequests  atomic.Int64
    totalFailures  atomic.Int64
}

func NewMetrics() *Metrics { return &Metrics{} }

func (m *Metrics) IncRequests() { m.totalRequests.Add(1) }
func (m *Metrics) IncFailures() { m.totalFailures.Add(1) }

func (m *Metrics) Handler() http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
        _, _ = w.Write([]byte(
            "# HELP go_agw_total_requests Total requests handled\n" +
                "# TYPE go_agw_total_requests counter\n" +
                "go_agw_total_requests " + itoa(m.totalRequests.Load()) + "\n" +
                "# HELP go_agw_total_failures Total failed requests\n" +
                "# TYPE go_agw_total_failures counter\n" +
                "go_agw_total_failures " + itoa(m.totalFailures.Load()) + "\n",
        ))
    })
}

func itoa(v int64) string {
    // minimal alloc integer to string
    if v == 0 { return "0" }
    neg := v < 0
    if neg { v = -v }
    buf := make([]byte, 0, 20)
    for v > 0 {
        d := byte(v % 10)
        buf = append([]byte{'0' + d}, buf...)
        v /= 10
    }
    if neg { buf = append([]byte{'-'}, buf...) }
    return string(buf)
}


package negroniprometheus

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/urfave/negroni"
)

var (
	dflBuckets = []float64{300, 1200, 5000}
)

const (
	latencyName = "negroni_request_duration_microseconds"
)

// Middleware is a handler that exposes prometheus metrics for the number of requests,
// the latency and the response size, partitioned by status code, method and HTTP path.
type Middleware struct {
	latency *prometheus.HistogramVec
}

// NewMiddleware returns a new prometheus Middleware handler.
func NewMiddleware(group, name string, buckets ...float64) *Middleware {
	var m Middleware

	if len(buckets) == 0 {
		buckets = dflBuckets
	}
	m.latency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: group,
		Subsystem: name,
		Name:      latencyName,
		Help:      "How long it took to process the request, partitioned by status code, method and HTTP path.",
		Buckets:   buckets,
	},
		[]string{"code", "method", "path"},
	)
	prometheus.MustRegister(m.latency)
	return &m
}

func (m *Middleware) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	start := time.Now()
	next(rw, r)
	res := negroni.NewResponseWriter(rw)
	m.latency.WithLabelValues(http.StatusText(res.Status()), r.Method, r.URL.Path).Observe(float64(time.Since(start).Nanoseconds()) / 1000)
}

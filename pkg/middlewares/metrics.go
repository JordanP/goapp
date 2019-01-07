package middlewares

import (
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	uuidV4 = regexp.MustCompile(`(?i)[0-9A-F]{8}-[0-9A-F]{4}-[4][0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12}`)
)

func MakeMetrics(promReg prometheus.Registerer, namespace string, buckets []float64) Middleware {
	if len(buckets) == 0 {
		buckets = prometheus.DefBuckets
	}
	if promReg == nil {
		promReg = prometheus.DefaultRegisterer
	}
	httpReqHist := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: "http",
		Name:      "request_duration_seconds",
		Help:      "The latency of the HTTP requests.",
		Buckets:   buckets,
	}, []string{"handler", "method", "code"})
	promReg.MustRegister(httpReqHist)

	return func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			wi := &responseWriterInterceptor{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}
			start := time.Now()
			h.ServeHTTP(wi, r)
			duration := time.Since(start).Seconds()
			// Replace variable UUID by :id to limit cardinality
			url := uuidV4.ReplaceAllString(r.URL.Path, ":id")
			httpReqHist.WithLabelValues(url, r.Method, strconv.Itoa(wi.statusCode)).Observe(duration)
		}
	}
}

type responseWriterInterceptor struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriterInterceptor) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

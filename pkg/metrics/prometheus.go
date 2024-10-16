package metrics

import (
	"time"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Counter for total HTTP requests received, labeled by method and endpoint
	RequestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status_code"},
	)

	// Histogram to track request durations in seconds, labeled by method and endpoint
	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint", "status_code"},
	)
)

// InitMetrics initializes Prometheus metrics
func InitMetrics() {
	// Register the metrics with Prometheus
	prometheus.MustRegister(RequestCounter)
	prometheus.MustRegister(RequestDuration)
}

// MetricsMiddlewareGin is a middleware for Gin to collect metrics for each HTTP request
func MetricsMiddlewareGin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start a timer for request duration
		startTime := time.Now()

		// Process the request
		c.Next()

		// Calculate the duration of the request
		duration := time.Since(startTime).Seconds()

		// Get the response status code
		statusCode := c.Writer.Status()

		// Increment the request counter with labels
		RequestCounter.WithLabelValues(c.Request.Method, c.FullPath(), http.StatusText(statusCode)).Inc()

		// Observe the duration of the request
		RequestDuration.WithLabelValues(c.Request.Method, c.FullPath(), http.StatusText(statusCode)).Observe(duration)
	}
}

// PrometheusHandler returns the HTTP handler for Prometheus metrics scraping
func PrometheusHandler() http.Handler {
	return promhttp.Handler()
}

package monitoring

import (
	"net/http"
	"time"
)

// Middleware records HTTP responses and request duration.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Exclude monitoring requests from traffic metrics.
		if r.URL.Path == "/metrics" {
			next.ServeHTTP(w, r)
			return
		}

		// Record the start time.
		start := time.Now()

		// Process the HTTP request.
		next.ServeHTTP(w, r)

		// Increment the HTTP response counter.
		HTTPResponses.Inc()

		// Record the request duration in seconds.
		RequestDuration.Observe(time.Since(start).Seconds())
	})
}

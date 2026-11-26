package monitoring

import (
	"database/sql"
	"log"
	"net/http"

	"maxi-knows/internal/storage"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// SearchRequests counts search requests by source.
var SearchRequests = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "maxiknows_search_requests_total",
		Help: "Total number of search requests.",
	},
	[]string{"source"},
)

// SearchZeroResults counts searches that return no results.
var SearchZeroResults = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "maxiknows_search_zero_results_total",
		Help: "Total number of searches returning no results.",
	},
	[]string{"source"},
)

// RegisteredUsers measures the current number of registered users.
var RegisteredUsers = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "maxiknows_registered_users",
		Help: "Current number of registered users.",
	},
)

// UsersWithValidSessions measures users with non-expired sessions.
var UsersWithValidSessions = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "maxiknows_users_with_valid_sessions",
		Help: "Current number of users with valid sessions.",
	},
)

// HTTPResponses counts HTTP responses sent by the application.
var HTTPResponses = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "maxiknows_http_responses_total",
		Help: "Total number of HTTP responses sent.",
	},
)

// RequestDuration measures HTTP request duration in seconds.
var RequestDuration = prometheus.NewHistogram(
	prometheus.HistogramOpts{
		Name:    "maxiknows_request_duration_seconds",
		Help:    "Duration of HTTP requests in seconds.",
		Buckets: prometheus.DefBuckets,
	},
)

func init() {
	prometheus.MustRegister(SearchRequests)
	prometheus.MustRegister(SearchZeroResults)
	prometheus.MustRegister(RegisteredUsers)
	prometheus.MustRegister(UsersWithValidSessions)
	prometheus.MustRegister(HTTPResponses)
	prometheus.MustRegister(RequestDuration)
}

// UpdateUserMetrics updates user metrics using the database.
func UpdateUserMetrics(db *sql.DB) error {
	users, err := storage.CountUsers(db)
	if err != nil {
		return err
	}

	usersWithValidSessions, err := storage.CountUsersWithValidSessions(db)
	if err != nil {
		return err
	}

	RegisteredUsers.Set(float64(users))
	UsersWithValidSessions.Set(float64(usersWithValidSessions))

	return nil
}

// MetricsHandler updates user metrics and exposes Prometheus metrics.
func MetricsHandler(db *sql.DB) http.HandlerFunc {
	handler := promhttp.Handler()

	return func(w http.ResponseWriter, r *http.Request) {
		if err := UpdateUserMetrics(db); err != nil {
			log.Printf("Could not update user metrics: %v", err)
			http.Error(w, "Could not collect metrics", http.StatusServiceUnavailable)
			return
		}

		handler.ServeHTTP(w, r)
	}
}

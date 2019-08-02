package app

import (
	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"time"
)

var (
	dbMetrics *prometheus.SummaryVec
)

func init() {
	dbMetrics = promauto.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       "db_access_latency",
			Help:       "db access latency (in seconds)",
			Objectives: map[float64]float64{0.5: 0.05, 0.95: 0.005, 0.99: 0.001},
			MaxAge:     time.Hour,
		},
		[]string{"query"},
	)
}

type queryFunction func(interface{}, string, ...interface{}) error

type cursorFunction func(string, ...interface{}) (*sqlx.Rows, error)

func queryWithMetrics(label string, function queryFunction) queryFunction {
	return func(destination interface{}, query string, args ...interface{}) (err error) {
		start := time.Now()
		err = function(destination, query, args...)
		if err == nil {
			dbMetrics.
				With(prometheus.Labels{"query": label}).
				Observe(time.Since(start).Seconds())
		}
		return
	}
}

func cursorWithMetrics(label string, function cursorFunction) cursorFunction {
	return func(query string, args ...interface{}) (*sqlx.Rows, error) {
		start := time.Now()
		cursor, err := function(query, args...)
		if err == nil {
			dbMetrics.
				With(prometheus.Labels{"query": label}).
				Observe(time.Since(start).Seconds())
		}
		return cursor, err
	}
}

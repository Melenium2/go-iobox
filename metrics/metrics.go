package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	msBuckets = []float64{10, 50, 100, 200, 300, 500, 1000, 1500, 3000, 5000, 7000, 10000, 20000, 60000}

	storageCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "storage_sql_counter_total",
		},
		[]string{"sql_query", "query_status"},
	)

	storageHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "storage_sql_latency_ms",
		},
		[]string{"sql_query", "sql_status"},
	)
)

func init() {
	prometheus.MustRegister(
		storageCounter,
		storageHistogram,
	)
}

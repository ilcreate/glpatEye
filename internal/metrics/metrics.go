package metrics

import (
	"fmt"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func InitAppMetrics(port string) {
	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	log.Println("Metrics exposed to endpoint '/metrics'")
	if err != nil {
		log.Fatalf("failed to start metrics server: %v", err)
	}
}

var (
	metricName           string = "gl_days_until_expire"
	tokenDaysUntilExpire        = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: metricName,
		Help: "Count days until expire Gitlab project access token",
	}, []string{
		"name",
		"project_name",
		"url_to_repo",
		"id",
		"last_used",
		"root_token",
	},
	)
)

type MetricLabels struct {
	Name, ProjectName, UrlToRepo, Id, LastUsed, Root string
	DaysExpire                                       int
}

func ResetMetrics() {
	log.Printf("Reset metrics!")
	tokenDaysUntilExpire.Reset()
}

func (m MetricLabels) UpdateMetric() {
	tokenDaysUntilExpire.WithLabelValues(
		m.Name,
		m.ProjectName,
		m.UrlToRepo,
		m.Id,
		m.LastUsed,
		m.Root,
	).Set(float64(m.DaysExpire))
	log.Printf("Updated metric for ID: %s", m.Id)
}

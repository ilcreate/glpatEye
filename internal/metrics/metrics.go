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
	// activeIDs sync.Map
)

type MetricLabels struct {
	Name, ProjectName, UrlToRepo, Id, LastUsed, Root string
	DaysExpire                                       int
}

func ResetMetrics() {
	log.Printf("Reset metrics!")
	tokenDaysUntilExpire.Reset()
}

// func ClearActiveIDs() {
// 	activeIDs = sync.Map{}
// }

func (m MetricLabels) UpdateMetric() {
	// activeIDs.Store(m.Id, true)
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

// func ResetStaleMetrics() {
// 	log.Println("Starting to reset stale metrics!")
// 	currentIDs := make(map[string]bool)
// 	activeIDs.Range(func(key, value interface{}) bool {
// 		currentIDs[key.(string)] = true
// 		return true
// 	})
// 	log.Printf("Current active IDs: %v", currentIDs)
// 	metricList, err := prometheus.DefaultGatherer.Gather()
// 	if err != nil {
// 		log.Printf("Failed to gather metrics: %v", err)
// 		return
// 	}
// 	log.Printf("METRIC LIST: %v", metricList)
// 	for _, metric := range metricList {
// 		if metric.GetName() == metricName {
// 			for _, m := range metric.GetMetric() {
// 				var id string
// 				labelValues := []string{}
// 				for _, label := range m.GetLabel() {
// 					if label.GetName() == "id" {
// 						id = label.GetValue()
// 					}
// 					labelValues = append(labelValues, label.GetValue())
// 				}
// 				if _, exists := activeIDs.Load(id); !exists {
// 					tokenDaysUntilExpire.DeleteLabelValues(labelValues...)
// 					log.Printf("Deleted stale metric with ID: %s", id)
// 				}
// 			}
// 		}
// 	}
// }

package handler

import (
	"net/http"
	"strings"

	"aliyun-exporter/pkg/collector"
	"aliyun-exporter/pkg/config"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func RegisterHandler(metrics map[string][]*config.Metric, cloudType []string) {
	http.HandleFunc("/probe", func(w http.ResponseWriter, r *http.Request) {
		target := r.URL.Query().Get("target")
		if target == "" {
			http.Error(w, "'target' parameter must be specified once", 400)
			return
		}

		for i := 0; i < len(cloudType); i++ {
			if cloudType[i] == "aliyun" {
				target = strings.Split(target, ".")[0]
			}
		}

		cloud := checkCloudType(target)
		c := &collector.CloudMonitor{
			Cloud:      cloud,
			InstanceID: target,
			Metrics:    metrics,
		}
		registry := prometheus.NewRegistry()
		registry.MustRegister(c)

		h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
		h.ServeHTTP(w, r)
	})
}

func checkCloudType(target string) string {
	var cloud string

	name := strings.Split(target, "-")[0]

	switch name {
	case "alb":
		cloud = "aliyun"
	case "nlb":
		cloud = "aliyun"
	case "lb":
		cloud = "qcloud"
	}
	return cloud
}

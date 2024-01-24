package handler

import (
	"net/http"
	"strings"

	"aliyun-exporter/pkg/collector"
	"aliyun-exporter/pkg/config"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func RegisterHandler(metrics map[string][]*config.Metric) {
	http.HandleFunc("/probe", func(w http.ResponseWriter, r *http.Request) {
		target := r.URL.Query().Get("target")
		if target == "" {
			http.Error(w, "'target' parameter must be specified once", 400)
			return
		}

		target = strings.Split(target, ".")[0]

		c := &collector.CloudMonitor{
			InstanceID: target,
			Metrics:    metrics,
		}
		registry := prometheus.NewRegistry()
		registry.MustRegister(c)

		h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
		h.ServeHTTP(w, r)
	})
}

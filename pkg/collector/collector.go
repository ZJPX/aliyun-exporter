package collector

import (
	"strings"

	"aliyun-exporter/pkg/cache"
	"aliyun-exporter/pkg/config"

	"github.com/prometheus/client_golang/prometheus"
)

type CloudMonitor struct {
	InstanceID string
	Metrics    map[string][]*config.Metric
}

func (m *CloudMonitor) Describe(ch chan<- *prometheus.Desc) {
}

func (m *CloudMonitor) Collect(ch chan<- prometheus.Metric) {
	for namespace, metrics := range m.Metrics {
		if !m.checkNamespace(namespace) {
			continue
		}
		namespace = strings.Split(namespace, "_")[1]
		for _, metric := range metrics {
			if ims, ok := cache.Metrics[m.InstanceID]; ok {
				dp := ims[metric.Name]
				val := dp.Get(metric.Measure)
				ch <- prometheus.MustNewConstMetric(
					metric.Desc(namespace),
					prometheus.GaugeValue,
					val,
				)
			}
		}
	}
}

func (m *CloudMonitor) checkNamespace(namespace string) bool {
	var name string
	switch namespace {
	case "acs_alb":
		name = "alb"
	case "acs_nlb":
		name = "nlb"
	case "acs_slb_dashboard":
		name = "lb"
	}
	if name != strings.Split(m.InstanceID, "-")[0] {
		return false
	}
	return true
}

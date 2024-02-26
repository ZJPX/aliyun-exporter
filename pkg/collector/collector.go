package collector

import (
	"strings"

	"aliyun-exporter/pkg/cache"
	"aliyun-exporter/pkg/config"

	"github.com/prometheus/client_golang/prometheus"
)

type CloudMonitor struct {
	CloudType  []string
	InstanceID string
	Metrics    map[string][]*config.Metric
}

func (m *CloudMonitor) Describe(ch chan<- *prometheus.Desc) {
}

func (m *CloudMonitor) Collect(ch chan<- prometheus.Metric) {
	for i := 0; i < len(m.CloudType); i++ {
		for namespace, metrics := range m.Metrics {
			if !m.checkNamespace(namespace) {
				continue
			}
			namespace = m.checkCloudType(m.CloudType[i], namespace)
			// namespace = strings.Split(namespace, "_")[1]
			for _, metric := range metrics {
				if ims, ok := cache.Metrics[m.CloudType[i]][m.InstanceID]; ok {
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
}

func (m *CloudMonitor) checkNamespace(namespace string) bool {
	var name string
	switch namespace {
	case "acs_alb":
		name = "alb"
	case "acs_nlb":
		name = "nlb"
	case "lb_public":
		name = "lb"
	}
	if name != strings.Split(m.InstanceID, "-")[0] {
		return false
	}
	return true
}

func (m *CloudMonitor) checkCloudType(cloudType, name string) string {
	switch cloudType {
	case "aliyun":
		name = strings.Split(name, "_")[1]
	case "qcloud":
	}

	return name
}

package collector

import (
	"aliyun-exporter/pkg/client"
	"aliyun-exporter/pkg/config"
	"strings"
	"sync"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

const AppName = "cloudmonitor"

// cloudMonitor ..
type cloudMonitor struct {
	namespace  string
	instanceID string
	cfg        *config.Config
	logger     log.Logger
	// sdk client
	client *client.MetricClient
	rate   int
	lock   sync.Mutex
}

// NewCloudMonitorCollector create a new collector for cloud monitor
func NewCloudMonitorCollector(appName string, cfg *config.Config, rate int, logger log.Logger) (map[string]prometheus.Collector, map[string]*client.MetricClient, error) {
	if logger == nil {
		logger = log.NewNopLogger()
	}
	mClient := make(map[string]*client.MetricClient)
	cloudMonitors := make(map[string]prometheus.Collector)
	for cloudID, credential := range cfg.Credentials {
		cli, err := client.NewMetricClient(cloudID, credential.AccessKey, credential.AccessKeySecret, credential.Region, logger)
		if err != nil {
			continue
		}
		cloudMonitors[cloudID] = &cloudMonitor{
			namespace: appName,
			cfg:       cfg,
			logger:    logger,
			client:    cli,
			rate:      rate,
		}
		mClient[cloudID] = cli
	}
	return cloudMonitors, mClient, nil
}

func NewExporterCollector(cli *client.MetricClient, cloudID, target string, cfg *config.Config, rate int, logger log.Logger) map[string]prometheus.Collector {
	if logger == nil {
		logger = log.NewNopLogger()
	}
	collectors := make(map[string]prometheus.Collector)
	collectors[cloudID] = &cloudMonitor{
		namespace:  AppName,
		instanceID: target,
		cfg:        cfg,
		logger:     logger,
		client:     cli,
		rate:       rate,
	}
	return collectors
}

func (m *cloudMonitor) Describe(ch chan<- *prometheus.Desc) {
}

func (m *cloudMonitor) Collect(ch chan<- prometheus.Metric) {
	// do collect
	m.client.SetTransport(m.rate)
	for sub, metrics := range m.cfg.Metrics {
		if ok := m.setNamespace(sub); !ok {
			continue
		}
		for i := range metrics {
			m.client.Collect(m.namespace, sub, m.instanceID, metrics[i], ch)
		}
	}
}

func (m *cloudMonitor) setNamespace(namespace string) bool {
	var name string
	switch namespace {
	case "acs_alb":
		name = "alb"
	case "acs_nlb":
		name = "nlb"
		// case "acs_slb_dashboard":
		// 	name = "lb"
	}
	if name != strings.Split(m.instanceID, "-")[0] {
		return false
	}
	return true
}

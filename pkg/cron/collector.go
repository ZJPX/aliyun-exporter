package cron

import (
	"sync"

	"aliyun-exporter/pkg/cache"
	"aliyun-exporter/pkg/client"
	"aliyun-exporter/pkg/config"
)

type cloudMonitor struct {
	metrics map[string][]*config.Metric
	client  *client.MetricClient
	lock    sync.Mutex
}

func (c *cloudMonitor) collect() {
	c.lock.Lock()
	defer c.lock.Unlock()

	wg := &sync.WaitGroup{}
	cache.MetricsTemp = make(map[string]map[string]cache.Datapoint)
	for namespace, metrics := range c.metrics {
		for _, m := range metrics {
			wg.Add(1)
			go func(namespace string, metric *config.Metric) {
				defer wg.Done()
				c.client.GetMetrics(namespace, metric)
			}(namespace, m)
		}
	}
	wg.Wait()
	cache.Metrics = cache.MetricsTemp
	return
}

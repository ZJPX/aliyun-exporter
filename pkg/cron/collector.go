package cron

import (
	"aliyun-exporter/pkg/cache"
	"aliyun-exporter/pkg/config"
	"sync"
)

func (m cloudMonitor) aliCloudMonitorCollect() {
	c := m.aliCloud
	c.lock.Lock()
	defer c.lock.Unlock()

	wg := &sync.WaitGroup{}
	// do collect
	c.client.SetTransport(c.rate)
	cache.MetricsTemp = make(map[string]map[string]cache.Datapoint)
	for sub, metrics := range c.cfg.Metrics {
		for i := range metrics {
			wg.Add(1)
			go func(namespace string, metric *config.Metric) {
				defer wg.Done()
				c.client.GetMetrics(namespace, metric)
			}(sub, metrics[i])
		}
	}
	wg.Wait()
	cache.Metrics = cache.MetricsTemp
	return
}

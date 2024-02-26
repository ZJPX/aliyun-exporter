package cron

import (
	"fmt"
	"sync"

	"aliyun-exporter/pkg/cache"
	"aliyun-exporter/pkg/client"
	"aliyun-exporter/pkg/config"
)

type cloudMonitor struct {
	cloudType string
	metrics   map[string][]*config.Metric
	client    *client.MetricClient
	lock      sync.Mutex
}

func (c *cloudMonitor) collect() {
	c.lock.Lock()
	defer c.lock.Unlock()

	wg := &sync.WaitGroup{}

	cache.MetricsTemp = make(map[string]map[string]cache.Datapoint)

	for namespace, metrics := range c.metrics {
		// 跳过腾讯云监控指标
		if namespace == "lb_public" {
			continue
		}

		for _, m := range metrics {
			wg.Add(1)
			go func(namespace string, metric *config.Metric) {
				defer wg.Done()
				c.client.GetMetrics(namespace, metric)
			}(namespace, m)
		}
	}
	wg.Wait()

	if _, ok := cache.Metrics[c.cloudType]; !ok {
		cache.Metrics[c.cloudType] = make(map[string]map[string]cache.Datapoint)
	}
	cache.Metrics[c.cloudType] = cache.MetricsTemp
	fmt.Printf("resp :%+v\n", cache.Metrics)
	return
}

package cron

import (
	"aliyun-exporter/pkg/cache"
	"aliyun-exporter/pkg/client"
	"aliyun-exporter/pkg/config"
	"fmt"
	"sync"
	"time"
)

var (
	product2Namespace = map[string]string{
		"lb_public": "QCE/LB_PUBLIC",
	}
)

type qCloudMonitor struct {
	cloudType string
	metrics   map[string][]*config.Metric
	repos     map[string]collectRepo
	lock      sync.Mutex
}

type collectRepo struct {
	instanceRepo client.TcInstanceRepository
	monitorRepo  client.TcmMetricRepository
}

func (c *qCloudMonitor) collect() {
	st := time.Now().Unix() - 120 // 当前时间戳减去120s
	fmt.Printf("st: %v\n", st)

	c.lock.Lock()
	defer c.lock.Unlock()

	wg := &sync.WaitGroup{}

	cache.TcMetricsTemp = make(map[string]map[string]cache.Datapoint)

	for region := range c.repos {
		repo := c.repos[region]
		insList, err := repo.instanceRepo.ListByFilters()
		// insList, err := c.repos[region].instanceRepo.ListByFilters()
		if err != nil {
			panic(err)
		}
		fmt.Printf("insList %d, region: %s, time: %s \n", len(insList), region, time.Now().String())

		for namespace, metrics := range c.metrics {
			tcNamespace, ok := product2Namespace[namespace]
			if !ok {
				continue
			}

			for k := range metrics {
				wg.Add(1)
				go func(tcNamespace string, insList []client.TcInstance, metric *config.Metric) {
					defer wg.Done()
					repo.monitorRepo.GetMetrics(tcNamespace, insList, metric, st, 0)
				}(tcNamespace, insList, metrics[k])
			}
		}
		wg.Wait()
	}

	if _, ok := cache.Metrics[c.cloudType]; !ok {
		cache.Metrics[c.cloudType] = make(map[string]map[string]cache.Datapoint)
	}
	cache.Metrics[c.cloudType] = cache.TcMetricsTemp
	fmt.Printf("resp :%+v\n", cache.Metrics)
	return
}

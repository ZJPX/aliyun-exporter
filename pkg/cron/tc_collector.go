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
	cloudType    string
	metrics      map[string][]*config.Metric
	instanceRepo client.TcInstanceRepository
	monitorRepo  client.TcmMetricRepository
	lock         sync.Mutex
}

func (c *qCloudMonitor) collect() {
	st := time.Now().Unix() - 120 // 当前时间戳减去120s
	fmt.Printf("st: %v\n", st)
	insList, err := c.instanceRepo.ListByFilters()
	if err != nil {
		return
	}
	fmt.Println(len(insList))

	c.lock.Lock()
	defer c.lock.Unlock()

	wg := &sync.WaitGroup{}

	cache.TcMetricsTemp = make(map[string]map[string]cache.Datapoint)

	for namespace, metrics := range c.metrics {
		tcNamespace, ok := product2Namespace[namespace]
		if !ok {
			continue
		}

		for _, m := range metrics {
			metric := m
			wg.Add(1)
			go func() {
				err = func(tcNamespace string, metric *config.Metric) error {
					defer wg.Done()
					return c.monitorRepo.GetMetrics(tcNamespace, insList, metric, st, 0)
				}(tcNamespace, metric)
				if err != nil {
					panic(err)
				}
			}()
		}
	}
	wg.Wait()

	if _, ok := cache.Metrics[c.cloudType]; !ok {
		cache.Metrics[c.cloudType] = make(map[string]map[string]cache.Datapoint)
	}
	cache.Metrics[c.cloudType] = cache.TcMetricsTemp
	fmt.Printf("resp :%+v\n", cache.Metrics)
	return
}

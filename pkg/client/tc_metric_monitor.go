package client

import (
	"aliyun-exporter/pkg/cache"
	"aliyun-exporter/pkg/config"
	"aliyun-exporter/pkg/util"
	"context"
	"strconv"
	"time"

	"golang.org/x/time/rate"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	monitor "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/monitor/v20180724"
)

var (
	timeStampFormat = "2006-01-02 15:04:05"
)

// TcmMetricRepository 腾讯云监控指标Repository
type TcmMetricRepository interface {
	// GetMetrics 按时间范围获取单个时间线的数据点
	GetMetrics(namespace string, instances []TcInstance, m *config.Metric, startTime int64, endTime int64) (err error)
}

// TcmMetricClient wrap monitor client
type TcmMetricClient struct {
	cloudID string
	client  *monitor.Client
	limiter *rate.Limiter // 限速
	ctx     context.Context
	logger  log.Logger
}

func (repo *TcmMetricClient) GetMetrics(namespace string, instances []TcInstance, m *config.Metric, st int64, et int64) (err error) {
	// 限速
	ctx, cancel := context.WithCancel(repo.ctx)
	defer cancel()
	err = repo.limiter.Wait(ctx)
	if err != nil {
		level.Error(repo.logger).Log("limiter err ", err.Error())
		return
	}

	request := repo.buildGetMonitorDataRequest(namespace, m, st, et)

	// response := &monitor.GetMonitorDataResponse{}
	batchSize := 10
	batchData := make([][]TcInstance, 0)
	for i := 0; i < len(instances); i += batchSize {
		j := i + batchSize
		if j > len(instances) {
			j = len(instances)
		}

		batchData = append(batchData, instances[i:j])
	}

	lock.Lock()
	defer lock.Unlock()
	for _, tcInstances := range batchData {
		request.Instances = []*monitor.Instance{}
		response := &monitor.GetMonitorDataResponse{}
		for _, instance := range tcInstances {
			instanceFilters := &monitor.Instance{
				Dimensions: []*monitor.Dimension{},
			}
			name := m.Dimensions[0]
			value := instance.GetInstanceId()
			instanceFilters.Dimensions = append(instanceFilters.Dimensions, &monitor.Dimension{Name: &name, Value: &value})
			request.Instances = append(request.Instances, instanceFilters)
		}

		response, err = repo.getMonitorDataWithRetry(request)
		if err != nil {
			level.Error(repo.logger).Log("GetMonitorDataErr ", err.Error())
			return
		}

		for _, points := range response.Response.DataPoints {
			for _, dim := range points.Dimensions {
				instanceID := *dim.Value
				if _, ok := cache.TcMetricsTemp[instanceID]; !ok {
					cache.TcMetricsTemp[instanceID] = make(map[string]cache.Datapoint)
				}
				if _, ok := cache.TcMetricsTemp[instanceID][m.Name]; !ok {
					cache.TcMetricsTemp[instanceID][m.Name] = make(map[string]interface{})
				}

				for i := 0; i < len(points.Timestamps); i++ {
					cache.TcMetricsTemp[instanceID][m.Name]["timestamp"] = *points.Timestamps[i]
					cache.TcMetricsTemp[instanceID][m.Name][m.Measure] = *points.Values[i]
				}
			}
		}
	}

	// fmt.Printf("len: %v\n", len(response.Response.DataPoints))
	// fmt.Printf("resp1: %s\n", response.ToJsonString())

	// lock.Lock()
	// defer lock.Unlock()
	// for _, points := range response.Response.DataPoints {
	// 	for _, dim := range points.Dimensions {
	// 		instanceID := *dim.Value
	// 		if _, ok := cache.TcMetricsTemp[instanceID]; !ok {
	// 			cache.TcMetricsTemp[instanceID] = make(map[string]cache.Datapoint)
	// 		}
	// 		if _, ok := cache.TcMetricsTemp[instanceID][m.Name]; !ok {
	// 			cache.TcMetricsTemp[instanceID][m.Name] = make(map[string]interface{})
	// 		}
	//
	// 		for i := 0; i < len(points.Timestamps); i++ {
	// 			cache.TcMetricsTemp[instanceID][m.Name]["timestamp"] = *points.Timestamps[i]
	// 			cache.TcMetricsTemp[instanceID][m.Name][m.Measure] = *points.Values[i]
	// 		}
	// 	}
	// }

	return err
}

func (repo *TcmMetricClient) buildGetMonitorDataRequest(namespace string, m *config.Metric, st, et int64) *monitor.GetMonitorDataRequest {
	request := monitor.NewGetMonitorDataRequest()
	request.Namespace = &namespace
	request.MetricName = &m.Name

	period, _ := strconv.ParseUint(m.Period, 10, 64)
	request.Period = &period

	stStr := util.FormatTime(time.Unix(st, 0), timeStampFormat)
	request.StartTime = &stStr
	if et != 0 {
		etStr := util.FormatTime(time.Unix(et, 0), timeStampFormat)
		request.EndTime = &etStr
	}

	return request
}

func (repo *TcmMetricClient) getMonitorDataWithRetry(request *monitor.GetMonitorDataRequest) (*monitor.GetMonitorDataResponse, error) {
	// var lastErr error
	monitorClient := repo.client
	// for i := 0; i < 3; i++ {
	// 	resp, err := monitorClient.GetMonitorData(request)
	// 	if err != nil {
	// 		if strings.Contains(err.Error(), context.DeadlineExceeded.Error()) {
	// 			lastErr = err
	// 			continue
	// 		}
	// 		return nil, err
	// 	}
	// 	return resp, nil
	// }
	// return nil, lastErr

	return monitorClient.GetMonitorData(request)
}

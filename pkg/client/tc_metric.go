package client

import "sync"

// TcmMetric 代表一个指标, 包含多个时间线
type TcmMetric struct {
	Id          string
	Namespace   string
	QueryLabels map[string]string
	seriesLock  sync.Mutex
}

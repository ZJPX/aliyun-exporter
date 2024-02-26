package client

import "reflect"

type TcInstance interface {
	// GetMonitorQueryKey 用于查询云监控数据的主键字段, 一般是实例id
	GetMonitorQueryKey() string

	// GetInstanceId 获取实例的id
	GetInstanceId() string
}

type baseTcInstance struct {
	instanceId string
	value      reflect.Value
}

func (ins *baseTcInstance) GetMonitorQueryKey() string {
	return ins.instanceId
}

func (ins *baseTcInstance) GetInstanceId() string {
	return ins.instanceId
}

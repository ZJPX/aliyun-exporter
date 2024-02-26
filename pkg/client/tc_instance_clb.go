package client

import (
	"fmt"
	"reflect"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	clb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
)

// TcInstanceRepository 每个产品的实例对象的Repository
type TcInstanceRepository interface {
	ListByFilters() ([]TcInstance, error)
}

// ClbTcInstanceRepository wrap clb client
type ClbTcInstanceRepository struct {
	cloudID string
	client  *clb.Client
	logger  log.Logger
}

type ClbInstance struct {
	baseTcInstance
	meta *clb.LoadBalancer
}

func (repo *ClbTcInstanceRepository) ListByFilters() (instances []TcInstance, err error) {
	request := clb.NewDescribeLoadBalancersRequest()
	var offset int64 = 0
	var limit int64 = 100
	var total int64 = -1
	var open = "OPEN"

	request.Offset = &offset
	request.Limit = &limit
	request.LoadBalancerType = &open

getMoreInstances:
	resp, err := repo.client.DescribeLoadBalancers(request)
	if err != nil {
		level.Error(repo.logger).Log("msg", fmt.Sprintf("An API error has returned: %s", err))
		return
	}

	if total == -1 {
		total = int64(*resp.Response.TotalCount)
	}
	for _, meta := range resp.Response.LoadBalancerSet {
		ins, e := NewClbTcInstance(*meta.LoadBalancerId, meta)
		if e != nil {
			level.Error(repo.logger).Log("msg", "Create clb instance fail", "id", *meta.LoadBalancerId)
			continue
		}
		if (meta.LoadBalancerVips == nil || len(meta.LoadBalancerVips) == 0) && meta.AddressIPv6 == nil {
			level.Warn(repo.logger).Log("msg", "clb instance no include vip", "id", *meta.LoadBalancerId)
			continue
		}
		instances = append(instances, ins)
	}
	offset += limit
	if offset < total {
		request.Offset = &offset
		goto getMoreInstances
	}

	return
}

func NewClbTcInstance(instanceId string, meta *clb.LoadBalancer) (ins *ClbInstance, err error) {
	if instanceId == "" {
		return nil, fmt.Errorf("instanceId is empty ")
	}
	if meta == nil {
		return nil, fmt.Errorf("meta is empty ")
	}
	ins = &ClbInstance{
		baseTcInstance: baseTcInstance{
			instanceId: instanceId,
			value:      reflect.ValueOf(*meta),
		},
		meta: meta,
	}
	return
}

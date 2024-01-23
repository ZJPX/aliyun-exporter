package client

import (
	"aliyun-exporter/pkg/cache"
	"aliyun-exporter/pkg/config"
	"aliyun-exporter/pkg/ratelimit"
	"encoding/json"
	"sync"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/cms"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

var lock sync.RWMutex

// map for all avaliable namespaces
// todo: is there a way to add desc into yaml file?
var allNamespaces = map[string]string{
	"acs_ecs_dashboard":              "Elastic Compute Service (ECS)",
	"acs_containerservice_dashboard": "Container Service for Swarm",
	"acs_kubernetes":                 "Container Service for Kubernetes (ACK)",
	"acs_oss_dashboard":              "Object Storage Service (OSS)",
	"acs_slb_dashboard":              "Server Load Balancer (SLB)",
	"acs_vpc_eip":                    "Elastic IP addresses (EIPs)",
	"acs_nat_gateway":                "NAT Gateway",
	"acs_anycast_eip":                "Anycast Elastic IP address (EIP)",
	"acs_rds_dashboard":              "ApsaraDB RDS",
	"acs_mongodb":                    "ApsaraDB for MongoDB",
	"acs_memcache":                   "ApsaraDB for Memcache",
	"acs_kvstore":                    "ApsaraDB for Redis",
	"acs_hitsdb":                     "Time Series Database (TSDB)",
	"acs_clickhouse":                 "ClickHouse",
	"acs_cds":                        "ApsaraDB for Cassandra",
	"waf":                            "Web Application Firewall (WAF)",
	"acs_elasticsearch":              "Elasticsearch",
	"acs_mns_new":                    "queues of Message Service (MNS)",
	"acs_kafka":                      "Message Queue for Apache Kafka",
	"acs_amqp":                       "Alibaba Cloud Message Queue for AMQP instances",
	"acs_alb":                        "Server Load Balancer (SLB)",
	"acs_nlb":                        "Server Load Balancer (SLB)",
}

// AllNamespaces return allNamespaces
func AllNamespaces() map[string]string {
	return allNamespaces
}

// allNamesOfNamespaces return all avaliable namespaces
func allNamesOfNamespaces() []string {
	ss := make([]string, 0, len(allNamespaces))
	for name := range allNamespaces {
		ss = append(ss, name)
	}
	return ss
}

// MetricClient wrap cms.client
type MetricClient struct {
	cloudID string
	cms     *cms.Client
	logger  log.Logger
}

// NewMetricClient create metric Client
func NewMetricClient(cloudID, ak, secret, region string, logger log.Logger) (*MetricClient, error) {
	cmsClient, err := cms.NewClientWithAccessKey(region, ak, secret)
	if err != nil {
		return nil, err
	}
	if logger == nil {
		logger = log.NewNopLogger()
	}
	return &MetricClient{cloudID, cmsClient, logger}, nil
}

func (c *MetricClient) SetTransport(rate int) {
	rt := ratelimit.New(rate)
	c.cms.SetTransport(rt)
}

func (c *MetricClient) createDescribeMetricLastReq(sub, name, period, nextToken string) (*cms.DescribeMetricLastResponse, error) {
	req := cms.CreateDescribeMetricLastRequest()
	req.ReadTimeout = 50 * time.Second
	req.Namespace = sub
	req.MetricName = name
	req.Period = period
	req.NextToken = nextToken
	return c.cms.DescribeMetricLast(req)
}

// retrive get datapoints for metric
func (c *MetricClient) retrive(sub, name, period string) ([]cache.Datapoint, error) {
	resp, err := c.createDescribeMetricLastReq(sub, name, period, "")
	if err != nil {
		level.Error(c.logger).Log("DescribeMetricLastReqErr", err)
		return nil, err
	}

	for resp.NextToken != "" {
		resp, err = c.createDescribeMetricLastReq(sub, name, period, resp.NextToken)
		if err != nil {
			level.Error(c.logger).Log("DescribeMetricLastReqErr", err)
			return nil, err
		}
	}

	var datapoints []cache.Datapoint
	if err = json.Unmarshal([]byte(resp.Datapoints), &datapoints); err != nil {
		// some execpected error
		level.Debug(c.logger).Log("content", resp.GetHttpContentString(), "error", err)
		return nil, err
	}
	return datapoints, nil
}

// GetMetrics get metrics into map
func (c *MetricClient) GetMetrics(sub string, m *config.Metric) {
	if m.Name == "" {
		level.Warn(c.logger).Log("msg", "metric name must been set")
		return
	}

	datapoints, err := c.retrive(sub, m.Name, m.Period)
	if err != nil {
		level.Error(c.logger).Log("msg", "failed to retrive datapoints", "cloudID", c.cloudID, "namespace", sub, "name", m.String(), "error", err)
		return
	}

	lock.Lock()
	defer lock.Unlock()
	instanceKey := m.Dimensions[0]
	for _, datapoint := range datapoints {
		instanceID := datapoint[instanceKey].(string)
		if _, ok := cache.MetricsTemp[instanceID]; !ok {
			cache.MetricsTemp[instanceID] = make(map[string]cache.Datapoint)
		}
		if _, ok := cache.MetricsTemp[instanceID][m.Name]; !ok {
			cache.MetricsTemp[instanceID][m.Name] = make(map[string]interface{})
		}
		cache.MetricsTemp[instanceID][m.Name]["timestamp"] = datapoint["timestamp"]
		cache.MetricsTemp[instanceID][m.Name][m.Measure] = datapoint[m.Measure]
	}
}

// DescribeMetricMetaList return metrics meta list
func (c *MetricClient) DescribeMetricMetaList(namespaces ...string) (map[string][]cms.Resource, error) {
	namespaces = filterNamespaces(namespaces...)
	m := make(map[string][]cms.Resource)
	for _, ns := range namespaces {
		req := cms.CreateDescribeMetricMetaListRequest()
		req.Namespace = ns
		req.PageSize = requests.NewInteger(100)
		resp, err := c.cms.DescribeMetricMetaList(req)
		if err != nil {
			return nil, err
		}
		level.Debug(c.logger).Log("content", resp.GetHttpContentString())
		m[ns] = resp.Resources.Resource
	}
	return m, nil
}

func filterNamespaces(namespaces ...string) []string {
	if len(namespaces) == 0 {
		return allNamesOfNamespaces()
	}
	filters := make([]string, 0)
	for _, ns := range namespaces {
		if ns == "all" {
			return allNamesOfNamespaces()
		}
		if _, ok := allNamespaces[ns]; ok {
			filters = append(filters, ns)
		}
	}
	return filters
}

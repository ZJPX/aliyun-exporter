package client

import (
	"context"
	"net"
	"net/http"
	"time"

	"golang.org/x/time/rate"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/cms"
	"github.com/go-kit/log"

	clb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	monitor "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/monitor/v20180724"
)

// NewTcClbClient 初始化腾讯云CLB客户端
func NewTcClbClient(cred common.CredentialIface, cloudID, region string, logger log.Logger) (repo TcInstanceRepository, err error) {
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "clb.tencentcloudapi.com"

	client, err := clb.NewClient(cred, region, cpf)
	if err != nil {
		return
	}
	if logger == nil {
		logger = log.NewNopLogger()
	}

	repo = &ClbTcInstanceRepository{
		cloudID: cloudID,
		client:  client,
		logger:  logger,
	}
	return
}

// NewTcMonitorClient 初始化腾讯云Monitor客户端
func NewTcMonitorClient(cred common.CredentialIface, cloudID, region string, rateLimit int, logger log.Logger) (repo TcmMetricRepository, err error) {
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "monitor.tencentcloudapi.com"

	client, err := newMonitorClient(cred, region, cpf)
	if err != nil {
		return
	}
	if logger == nil {
		logger = log.NewNopLogger()
	}

	limiter := rate.NewLimiter(rate.Limit(rateLimit), 1)
	repo = &TcmMetricClient{
		cloudID: cloudID,
		client:  client,
		limiter: limiter,
		ctx:     context.Background(),
		logger:  logger,
	}
	return
}

func newMonitorClient(credential common.CredentialIface,
	region string, clientProfile *profile.ClientProfile) (client *monitor.Client, err error) {
	client = &monitor.Client{}
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 5 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          0,
		IdleConnTimeout:       30 * time.Second,
		TLSHandshakeTimeout:   30 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	clientProfile.HttpProfile.ReqTimeout = 30
	client.Init(region).
		WithCredential(credential).
		WithProfile(clientProfile).WithHttpTransport(transport)
	return
}

// NewMetricClient 初始化阿里云metric客户端
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

package cron

import (
	"aliyun-exporter/pkg/cache"
	"aliyun-exporter/pkg/client"
	"aliyun-exporter/pkg/config"
	"fmt"

	"github.com/robfig/cron/v3"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"

	"github.com/go-kit/log"
)

func New(logger log.Logger, cfg *config.Config, cloudType []string) (err error) {
	c := cron.New(cron.WithSeconds())

	cache.Metrics = make(map[string]map[string]map[string]cache.Datapoint)

	for i := 0; i < len(cloudType); i++ {
		switch cloudType[i] {
		case "aliyun":
			err = aliYunMetric(logger, cfg, c, cloudType[i])
			fmt.Println("阿里云执行成功...")
		case "qcloud":
			err = qCloudMetric(logger, cfg, c, cloudType[i])
			fmt.Println("腾讯云执行成功...")
		}
	}

	// m, err = aliYunMetricClient(logger, cfg)
	// if err != nil {
	// 	level.Error(logger).Log("AliYunMetricClientErr", err)
	// 	return
	// }

	// _, err = c.AddFunc(cfg.Cron.Spec, m.collect)
	// if err != nil {
	// 	return
	// }

	c.Start()
	return
}

// aliYunMetric 初始化阿里云sdk
func aliYunMetric(logger log.Logger, cfg *config.Config, c *cron.Cron, cloudType string) (err error) {
	credential := cfg.Credentials["tenantId1"]
	cli, err := client.NewMetricClient(
		"tenantId1",
		credential.AccessKey,
		credential.AccessKeySecret,
		credential.Region,
		logger,
	)
	if err != nil {
		return
	}

	m := &cloudMonitor{
		cloudType: cloudType,
		metrics:   cfg.Metrics,
		client:    cli,
	}

	_, err = c.AddFunc(cfg.Cron.Spec, m.collect)
	if err != nil {
		return
	}
	return
}

func qCloudMetric(logger log.Logger, cfg *config.Config, c *cron.Cron, cloudType string) (err error) {
	credential := cfg.Credentials["tenantId2"]
	cred := common.NewCredential(
		credential.AccessKey,
		credential.AccessKeySecret,
	)
	instanceRepo, err := client.NewTcClbClient(cred, "tenantId2", credential.Region, logger)
	if err != nil {
		return
	}
	monitorRepo, err := client.NewTcMonitorClient(cred, "tenantId2", credential.Region, cfg.RateLimit, logger)
	if err != nil {
		return
	}

	tcm := qCloudMonitor{
		cloudType:    cloudType,
		metrics:      cfg.Metrics,
		instanceRepo: instanceRepo,
		monitorRepo:  monitorRepo,
	}

	_, err = c.AddFunc(cfg.Cron.Spec, tcm.collect)
	if err != nil {
		return
	}
	return
}

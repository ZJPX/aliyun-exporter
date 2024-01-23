package cron

import (
	"aliyun-exporter/pkg/client"
	"aliyun-exporter/pkg/config"

	"github.com/go-kit/log"
	"github.com/robfig/cron/v3"
)

func New(logger log.Logger, cfg *config.Config) (err error) {
	credential := cfg.Credentials["tenantId1"]
	cli, err := client.NewMetricClient(
		"tenantId1",
		credential.AccessKey,
		credential.AccessKeySecret,
		credential.Region,
		logger,
	)
	if err != nil {
		return err
	}

	m := &cloudMonitor{
		metrics: cfg.Metrics,
		client:  cli,
	}

	c := cron.New(cron.WithSeconds())

	_, err = c.AddFunc(cfg.Cron.Spec, m.collect)
	if err != nil {
		return
	}

	c.Start()
	return
}

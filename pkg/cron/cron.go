package cron

import (
	"aliyun-exporter/pkg/client"
	"aliyun-exporter/pkg/config"
	"sync"

	"github.com/go-kit/kit/log"
	"github.com/robfig/cron/v3"
)

type cloudMonitor struct {
	aliCloud *aliCloudMonitor
}

type aliCloudMonitor struct {
	cfg    *config.Config
	logger log.Logger
	// sdk client
	client *client.MetricClient
	rate   int
	lock   sync.Mutex
}

var m cloudMonitor

func New(logger log.Logger, rate int, cfg *config.Config, client map[string]*client.MetricClient) (err error) {
	m.aliCloud = &aliCloudMonitor{
		cfg:    cfg,
		logger: logger,
		client: client["tenantId1"],
		rate:   rate,
	}

	c := cron.New(cron.WithSeconds())
	err = initCron(c, cfg.Cron.Spec)
	if err != nil {
		return
	}
	c.Start()
	return
}

func initCron(c *cron.Cron, spec string) (err error) {
	_, err = c.AddFunc(spec, collect)
	return
}

func collect() {
	m.aliCloudMonitorCollect()
	return
}

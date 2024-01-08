package handler

import (
	"aliyun-exporter/pkg/client"
	"aliyun-exporter/pkg/collector"
	"aliyun-exporter/pkg/config"
	"strings"

	"fmt"
	"net"
	"net/http"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Handler http metrics handler
type Handler struct {
	logger log.Logger
	server *http.Server
}

// New create http handler
func New(addr string, logger log.Logger, rate int, cfg *config.Config, c map[string]prometheus.Collector, mClient map[string]*client.MetricClient) (*Handler, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	h := &Handler{
		logger: logger,
		server: &http.Server{
			Addr: net.JoinHostPort(host, port),
		},
	}
	http.HandleFunc("/probe", func(w http.ResponseWriter, r *http.Request) {
		handlerExporter(w, r, logger, rate, cfg, mClient)
	})
	return h, nil
}

func handlerExporter(w http.ResponseWriter, r *http.Request, logger log.Logger, rate int, cfg *config.Config, mClient map[string]*client.MetricClient) {
	query := r.URL.Query()
	target := query.Get("target")
	if len(query["target"]) != 1 || target == "" {
		http.Error(w, "'target' parameter must be specified once", 400)
		level.Error(logger).Log("'target' parameter must be specified once")
		return
	}

	str := strings.Split(target, ".")[0]
	c := collector.NewExporterCollector(mClient["tenantId1"], "tenantId1", str, cfg, rate, logger)
	registry := prometheus.NewRegistry()
	for i, _ := range c {
		registry.MustRegister(c[i])
	}
	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
	return
}

// Run start server
func (h *Handler) Run() error {
	level.Info(h.logger).Log("msg", "Starting metric handler", "port", h.server.Addr)
	fmt.Println("msg", "Starting metric handler", "port", h.server.Addr)
	return h.server.ListenAndServe()
}

package handler

import (
	"aliyun-exporter/pkg/client"
	"aliyun-exporter/pkg/collector"
	"aliyun-exporter/pkg/config"

	"fmt"
	"net"
	"net/http"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"sigs.k8s.io/yaml"
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
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
            <head>
            <title>Aliyun Exporter</title>
            <style>
            label{
            display:inline-block;
            width:160px;
            }
            form label {
            margin: 10px;
            }
            form input {
            margin: 10px;
            }
            </style>
            </head>
            <body>
            <h1>Aliyun Exporter</h1>
            <form action="/monitors">
            <label>tenantId:</label> <input type="text" name="tenantId" placeholder="" value="tenant001" style="width:210px" required><br>
            <label>accessKey:</label> <input type="text" name="accessKey" placeholder="" value="" style="width:210px" required><br>
            <label>accessKeySecret:</label> <input type="text" name="accessKeySecret" placeholder="" value="" style="width:210px" required><br>
            <label>regionId:</label> <input type="text" name="regionId" placeholder="" value="cn-hangzhou" style="width:210px"><br>
            <input type="submit" value="Submit">
            </form>
						<p><a href="/config">Config</a></p>
            </body>
            </html>`))
	})
	http.HandleFunc("/metrics/exporter", func(w http.ResponseWriter, r *http.Request) {
		handlerExporter(w, r, logger, rate, cfg, mClient)
	})
	http.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		c, err := yaml.Marshal(cfg)
		if err != nil {
			level.Error(logger).Log("msg", "Error marshaling configuration", "err", err)
			http.Error(w, err.Error(), 500)
			return
		}
		w.Write(c)
	})
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Service is UP"))
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

	c := collector.NewExporterCollector(mClient["tenantId1"], "tenantId1", target, cfg, rate, logger)
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

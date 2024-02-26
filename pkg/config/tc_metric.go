package config

import (
	"github.com/prometheus/client_golang/prometheus"
)

// TcMetric meta
type TcMetric struct {
	Name        string   `json:"name"`
	Alias       string   `json:"alias,omitempty"`
	Period      string   `json:"period,omitempty"`
	Description string   `json:"desc,omitempty"`
	Dimensions  []string `json:"dimensions,omitempty"`
	Unit        string   `json:"unit,omitempty"`
	Measure     string   `json:"measure,omitempty"`
	Format      bool     `json:"format,omitempty"`
	desc        *prometheus.Desc
}

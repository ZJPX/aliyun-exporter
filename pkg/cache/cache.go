package cache

import (
	"fmt"
	"sort"
)

// Datapoint datapoint
type Datapoint map[string]interface{}

var Metrics map[string]map[string]map[string]Datapoint
var MetricsTemp map[string]map[string]Datapoint
var TcMetricsTemp map[string]map[string]Datapoint

var ignores = map[string]struct{}{
	"timestamp": {},
	"Maximum":   {},
	"Minimum":   {},
	"Average":   {},
}

// Get return value for measure
func (d Datapoint) Get(measure string) float64 {
	if v, ok := d[measure]; ok {
		if v == nil {
			return 0
		}
		return v.(float64)
	}
	return 0
}

// Labels return labels that not in ignores
func (d Datapoint) Labels() []string {
	labels := make([]string, 0)
	for k := range d {
		if _, ok := ignores[k]; !ok {
			labels = append(labels, k)
		}
	}
	sort.Strings(labels)
	return labels
}

// Values return values for lables
func (d Datapoint) Values(labels ...string) []string {
	values := make([]string, 0, len(labels))
	for i := range labels {
		values = append(values, fmt.Sprintf("%s", d[labels[i]]))
	}
	return values
}

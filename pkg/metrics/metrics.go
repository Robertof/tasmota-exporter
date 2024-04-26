package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	pm     *PlainMetrics
}

func NewMetrics(pm *PlainMetrics) *Metrics {
	return &Metrics{pm}
}

func (m *Metrics) Describe(ch chan<- *prometheus.Desc) {
	// create an unchecked collector, as our metrics could change at any time.
}

func (m *Metrics) Collect(ch chan<- prometheus.Metric) {
	m.pm.lock.Lock()
	defer m.pm.lock.Unlock()

	for source, sourceMetrics := range m.pm.metrics {
		statusTopic, topicExists := sourceMetrics.Metrics["status_topic"].(string)
		statusNetHostname, hostnameExists := sourceMetrics.Metrics["status_net_hostname"].(string)
		statusDeviceName, deviceNameExists := sourceMetrics.Metrics["status_device_name"].(string)

		if !topicExists || !hostnameExists || !deviceNameExists {
			continue
		}

		for pmk, pmv := range sourceMetrics.Metrics {
			send := func(pmk string, val float64, extraLabels prometheus.Labels) {
				labels := prometheus.Labels{
					"source": source,
					"status_topic": statusTopic,
					"status_net_hostname": statusNetHostname,
					"status_device_name": statusDeviceName,
				}

				if extraLabels != nil {
					for k, v := range extraLabels {
						labels[k] = v
					}
				}

				metric := prometheus.MustNewConstMetric(
					prometheus.NewDesc(pmk, "", nil, labels),
					prometheus.GaugeValue,
					val,
				)

				ch <- prometheus.NewMetricWithTimestamp(
					sourceMetrics.lastAccessed,
					metric,
				)
			}

			if float, ok := pmv.(float64); ok {
				send(pmk, float, nil)
			} else if multi, ok := pmv.(MultiDimMetric); ok {
				for labelVal, metricVal := range multi.values {
					send(pmk, metricVal, prometheus.Labels{
						multi.label: labelVal,
					})
				}
			}
		}
	}
}

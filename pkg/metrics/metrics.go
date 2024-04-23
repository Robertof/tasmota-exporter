package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	pm     *PlainMetrics
	gauges map[string]*prometheus.GaugeVec
}

func NewMetrics(pm *PlainMetrics) *Metrics {
	return &Metrics{pm, map[string]*prometheus.GaugeVec{}}
}

func (m *Metrics) Refresh() {
	m.pm.lock.Lock()
	defer m.pm.lock.Unlock()
	for source, sourceMetrics := range m.pm.metrics {
		statusTopic, topicExists := sourceMetrics.Metrics["status_topic"].(string)
		statusNetHostname, hostnameExists := sourceMetrics.Metrics["status_net_hostname"].(string)
		statusNetIpAddress, ipAddressExists := sourceMetrics.Metrics["status_net_ip_address"].(string)
		statusDeviceName, deviceNameExists := sourceMetrics.Metrics["status_device_name"].(string)

		if !topicExists || !hostnameExists || !ipAddressExists || !deviceNameExists {
			continue
		}

		for pmk, pmv := range sourceMetrics.Metrics {
			if float, ok := pmv.(float64); ok {
				m.registerGauge(pmk, []string{
					"source", "status_topic", "status_net_hostname",
					"status_net_ip_address", "status_device_name",
				})
				m.gauges[pmk].WithLabelValues(source, statusTopic, statusNetHostname, statusNetIpAddress, statusDeviceName).Set(float)
			} else if multi, ok := pmv.(MultiDimMetric); ok {
				m.registerGauge(pmk, []string{
					"source", "status_topic", "status_net_hostname",
					"status_net_ip_address", "status_device_name",
					multi.label,
				})

				for labelVal, metricVal := range multi.values {
					m.gauges[pmk].WithLabelValues(
						source, statusTopic, statusNetHostname,
						statusNetIpAddress, statusDeviceName,
						labelVal,
					).Set(metricVal)
				}
			}
		}
	}
}

func (m *Metrics) registerGauge(pmk string, labelNames []string) {
	if _, ok := m.gauges[pmk]; !ok {
		m.gauges[pmk] = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name:      pmk,
				Namespace: "tasmota",
			},
			labelNames,
		)
		prometheus.MustRegister(m.gauges[pmk])
	}
}

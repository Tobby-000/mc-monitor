package collector

import (
	"github.com/prometheus/client_golang/prometheus"

	"mc-monitor/internal/probe"
)

type MCCollector struct {
	cache        *probe.Cache
	onlineDesc   *prometheus.Desc
	playerDesc   *prometheus.Desc
	protocolDesc *prometheus.Desc
	totalDesc    *prometheus.Desc
}

func NewMCCollector(cache *probe.Cache) (*MCCollector, error) {
	return &MCCollector{
		cache: cache,
		onlineDesc: prometheus.NewDesc(
			"minecraft_server_online",
			"1 if the server is online, 0 otherwise.",
			[]string{"server"},
			nil,
		),
		playerDesc: prometheus.NewDesc(
			"minecraft_players_online",
			"Number of players currently online.",
			[]string{"server"},
			nil,
		),
		protocolDesc: prometheus.NewDesc(
			"minecraft_probe_protocol_duration_seconds",
			"Duration of the Minecraft Server List Ping protocol exchange (handshake, status request, status response).",
			[]string{"server"},
			nil,
		),
		totalDesc: prometheus.NewDesc(
			"minecraft_probe_total_duration_seconds",
			"Total duration of the Minecraft server probe, including DNS resolution, TCP handshake, and protocol exchange.",
			[]string{"server"},
			nil,
		),
	}, nil
}

func (m *MCCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- m.onlineDesc
	ch <- m.playerDesc
	ch <- m.protocolDesc
}

func (m *MCCollector) Collect(ch chan<- prometheus.Metric) {
	snap := m.cache.Snapshot()
	for name, res := range snap {
		online := 0.0
		if res.Online {
			online = 1.0
		}
		ch <- prometheus.MustNewConstMetric(
			m.onlineDesc, prometheus.GaugeValue, online, name)
		if res.Err == nil {
			ch <- prometheus.MustNewConstMetric(
				m.playerDesc, prometheus.GaugeValue, float64(res.Players), name)
			ch <- prometheus.MustNewConstMetric(
				m.protocolDesc, prometheus.GaugeValue, res.ProtocolDuration.Seconds(), name)
			ch <- prometheus.MustNewConstMetric(
				m.totalDesc, prometheus.GaugeValue, res.TotalDuration.Seconds(), name)
		}
	}
}

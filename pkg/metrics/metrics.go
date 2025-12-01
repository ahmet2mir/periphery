package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	PrefixUp = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "herald_prefix_up",
			Help: "Prefix announcement status (1=announced, 0=withdrawn)",
		},
		[]string{"prefix", "name"},
	)

	ProbeSuccess = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "herald_probe_success_total",
			Help: "Total number of successful probe executions",
		},
		[]string{"prefix", "probe_type", "name"},
	)

	ProbeFailure = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "herald_probe_failure_total",
			Help: "Total number of failed probe executions",
		},
		[]string{"prefix", "probe_type", "name"},
	)

	ProbeDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "herald_probe_duration_seconds",
			Help:    "Duration of probe execution in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"prefix", "probe_type", "name"},
	)

	BGPPeerUp = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "herald_bgp_peer_up",
			Help: "BGP peer status (1=established, 0=down)",
		},
		[]string{"peer_address", "peer_asn"},
	)

	BGPPeerState = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "herald_bgp_peer_state",
			Help: "BGP peer session state (0=unknown, 1=idle, 2=connect, 3=active, 4=opensent, 5=openconfirm, 6=established)",
		},
		[]string{"peer_address", "peer_asn"},
	)

	BGPPeerMessagesSent = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "herald_bgp_peer_messages_sent_total",
			Help: "Total number of BGP messages sent to peer",
		},
		[]string{"peer_address", "peer_asn", "message_type"},
	)

	BGPPeerMessagesReceived = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "herald_bgp_peer_messages_received_total",
			Help: "Total number of BGP messages received from peer",
		},
		[]string{"peer_address", "peer_asn", "message_type"},
	)

	BGPRouteCount = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "herald_bgp_route_count",
			Help: "Number of BGP routes",
		},
		[]string{"route_table"},
	)

	ServiceRestarts = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "herald_service_restarts_total",
			Help: "Total number of service restarts triggered",
		},
		[]string{"name"},
	)
)

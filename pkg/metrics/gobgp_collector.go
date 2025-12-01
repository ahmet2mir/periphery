package metrics

import (
	"context"
	"strconv"
	"time"

	api "github.com/osrg/gobgp/v3/api"
	"github.com/osrg/gobgp/v3/pkg/server"
	"go.uber.org/zap"
)

type GoBGPCollector struct {
	server   *server.BgpServer
	ctx      context.Context
	interval time.Duration
	stopCh   chan struct{}
}

func NewGoBGPCollector(bgpServer *server.BgpServer, ctx context.Context, interval time.Duration) *GoBGPCollector {
	return &GoBGPCollector{
		server:   bgpServer,
		ctx:      ctx,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

func (c *GoBGPCollector) Start() {
	go func() {
		ticker := time.NewTicker(c.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				c.collect()
			case <-c.stopCh:
				return
			case <-c.ctx.Done():
				return
			}
		}
	}()
}

func (c *GoBGPCollector) Stop() {
	close(c.stopCh)
}

func (c *GoBGPCollector) collect() {
	c.collectPeerMetrics()
	c.collectRouteMetrics()
}

func (c *GoBGPCollector) collectPeerMetrics() {
	err := c.server.ListPeer(c.ctx, &api.ListPeerRequest{}, func(peer *api.Peer) {
		if peer == nil {
			return
		}

		peerAddr := peer.Conf.NeighborAddress
		peerASN := strconv.FormatUint(uint64(peer.Conf.PeerAsn), 10)

		if peer.State != nil {
			state := float64(peer.State.SessionState)
			BGPPeerState.WithLabelValues(peerAddr, peerASN).Set(state)

			if peer.State.SessionState == api.PeerState_ESTABLISHED {
				BGPPeerUp.WithLabelValues(peerAddr, peerASN).Set(1)
			} else {
				BGPPeerUp.WithLabelValues(peerAddr, peerASN).Set(0)
			}

			if peer.State.Messages != nil {
				if peer.State.Messages.Sent != nil {
					BGPPeerMessagesSent.WithLabelValues(peerAddr, peerASN, "update").Add(float64(peer.State.Messages.Sent.Update))
					BGPPeerMessagesSent.WithLabelValues(peerAddr, peerASN, "notification").Add(float64(peer.State.Messages.Sent.Notification))
					BGPPeerMessagesSent.WithLabelValues(peerAddr, peerASN, "open").Add(float64(peer.State.Messages.Sent.Open))
					BGPPeerMessagesSent.WithLabelValues(peerAddr, peerASN, "keepalive").Add(float64(peer.State.Messages.Sent.Keepalive))
				}

				if peer.State.Messages.Received != nil {
					BGPPeerMessagesReceived.WithLabelValues(peerAddr, peerASN, "update").Add(float64(peer.State.Messages.Received.Update))
					BGPPeerMessagesReceived.WithLabelValues(peerAddr, peerASN, "notification").Add(float64(peer.State.Messages.Received.Notification))
					BGPPeerMessagesReceived.WithLabelValues(peerAddr, peerASN, "open").Add(float64(peer.State.Messages.Received.Open))
					BGPPeerMessagesReceived.WithLabelValues(peerAddr, peerASN, "keepalive").Add(float64(peer.State.Messages.Received.Keepalive))
				}
			}
		}
	})

	if err != nil {
		zap.S().Debug("Failed to collect BGP peer metrics", err)
	}
}

func (c *GoBGPCollector) collectRouteMetrics() {
	family := &api.Family{Afi: api.Family_AFI_IP, Safi: api.Family_SAFI_UNICAST}

	count := 0
	err := c.server.ListPath(c.ctx, &api.ListPathRequest{
		Family: family,
	}, func(destination *api.Destination) {
		count++
	})

	if err != nil {
		zap.S().Debug("Failed to collect BGP route metrics", err)
		return
	}

	BGPRouteCount.WithLabelValues("global").Set(float64(count))
}

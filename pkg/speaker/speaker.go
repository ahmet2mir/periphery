package speaker

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"strconv"

	api "github.com/osrg/gobgp/v3/api"
	"github.com/osrg/gobgp/v3/pkg/packet/bgp"
	"github.com/osrg/gobgp/v3/pkg/server"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/ahmet2mir/periphery/pkg/config"
	"github.com/ahmet2mir/periphery/pkg/logger"
)

type Speaker struct {
	Config  *config.Config
	Server  *server.BgpServer
	Context context.Context
}

func New(c *config.Config, ctx context.Context) (*Speaker, error) {
	s := server.NewBgpServer(
		server.GrpcListenAddress(c.API.GetURI()),
		server.LoggerOption(logger.NewGoBGPLogger()),
	)
	return &Speaker{Config: c, Server: s, Context: ctx}, nil
}

func (s *Speaker) Stop() {
	if err := s.Server.StopBgp(s.Context, &api.StopBgpRequest{}); err != nil {
		zap.S().Warn("Unable to top bgp %w", err)
	}
	s.Server.Stop()
}

func (s *Speaker) Serve() {
	s.Server.Serve()
}

func (s *Speaker) Start() error {
	if err := s.startBgp(); err != nil {
		return fmt.Errorf("setup error starting bgp: %w", err)
	}
	if err := s.addNeighbors(); err != nil {
		return fmt.Errorf("setup error adding neighbors: %w", err)
	}
	return nil
}

func (s *Speaker) startBgp() error {
	g := &api.Global{
		Asn:        s.Config.Speaker.ASN,
		RouterId:   s.Config.Speaker.RouterID,
		ListenPort: -1,
	}
	if s.Config.Speaker.GracefulRestartEnabled {
		g.GracefulRestart = &api.GracefulRestart{
			Enabled:     true,
			RestartTime: s.Config.Speaker.GracefulRestartRestartTime,
		}
	} else {
		g.GracefulRestart = &api.GracefulRestart{
			Enabled: false,
		}
	}
	return s.Server.StartBgp(s.Context, &api.StartBgpRequest{Global: g})
}

func (s *Speaker) addNeighbors() error {
	for _, neighbor := range s.Config.Neighbors {
		peer := &api.Peer{
			Conf: &api.PeerConf{
				NeighborAddress: neighbor.Address,
				PeerAsn:         neighbor.ASN,
			},
			EbgpMultihop: &api.EbgpMultihop{
				Enabled: neighbor.EbgpMultihopEnabled,
			},
		}

		zap.S().Info("NeighborAddress", neighbor.Address, "PeerAsn", neighbor.ASN, "Enabled", neighbor.EbgpMultihopEnabled)

		if err := s.Server.AddPeer(s.Context, &api.AddPeerRequest{Peer: peer}); err != nil {
			return err
		}
	}
	return nil
}

func (s *Speaker) anycastPath(p config.Prefix) (*api.Path, error) {
	ip, nw, err := net.ParseCIDR(p.IPAddress)
	if err != nil {
		return nil, fmt.Errorf("failed ParseCIDR %w", err)
	}
	ones, _ := nw.Mask.Size()
	nlri, err := anypb.New(&api.IPAddressPrefix{
		Prefix:    ip.String(),
		PrefixLen: uint32(ones),
	})
	if err != nil {
		return nil, fmt.Errorf("error creating network layer reachability information: %w", err)
	}
	attrs := make([]*anypb.Any, 0, 1)

	origin := &api.OriginAttribute{
		Origin: uint32(bgp.BGP_ORIGIN_ATTR_TYPE_IGP),
	}
	if attr, err := anypb.New(origin); err != nil {
		return nil, fmt.Errorf("error origin %w", err)
	} else {
		attrs = append(attrs, attr)
	}

	nextHop := &api.NextHopAttribute{
		NextHop: p.NextHop,
	}
	if attr, err := anypb.New(nextHop); err != nil {
		return nil, fmt.Errorf("error nextHop %w", err)
	} else {
		attrs = append(attrs, attr)
	}

	var ucom []uint32
	var _regexpCommunity = regexp.MustCompile(`(\d+):(\d+)`)
	zap.S().Info("p.Communities", "p.Communities", p.Communities)

	for _, c := range p.Communities {
		i, err := strconv.ParseUint(c, 10, 32)
		if err == nil {
			ucom = append(ucom, uint32(i))
		} else {
			elems := _regexpCommunity.FindStringSubmatch(c)
			if len(elems) == 3 {
				fst, _ := strconv.ParseUint(elems[1], 10, 16)
				snd, _ := strconv.ParseUint(elems[2], 10, 16)
				ucom = append(ucom, uint32(fst<<16|snd))
			}
		}
	}
	communities := &api.CommunitiesAttribute{
		Communities: ucom,
	}
	if attr, err := anypb.New(communities); err != nil {
		return nil, fmt.Errorf("error communities %w", err)
	} else {
		attrs = append(attrs, attr)
	}

	zap.S().Info("Attributes", "communities", communities)
	zap.S().Info("Attributes", "nextHop", nextHop)
	zap.S().Info("Attributes", "origin", origin)

	return &api.Path{
		Family: &api.Family{Afi: api.Family_AFI_IP, Safi: api.Family_SAFI_UNICAST},
		Nlri:   nlri,
		Pattrs: attrs,
	}, nil
}

func (s *Speaker) AddPath(p config.Prefix) error {
	path, err := s.anycastPath(p)
	if err != nil {
		return err
	}
	zap.S().Info("addPath", "anycast_ip", p.IPAddress)
	_, err = s.Server.AddPath(s.Context, &api.AddPathRequest{Path: path})
	return err
}

func (s *Speaker) DeletePath(p config.Prefix) error {
	bgpPath, err := s.anycastPath(p)
	if err != nil {
		return err
	}
	zap.S().Warn("deletePath", "anycast_ip", p.IPAddress)
	return s.Server.DeletePath(s.Context, &api.DeletePathRequest{Path: bgpPath})
}

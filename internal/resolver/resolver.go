package resolver

import (
	"fmt"
	"time"

	"github.com/miekg/dns"
)

type Resolver struct {
	udpClient   *dns.Client
	tcpClient   *dns.Client
	upstreams   []string
	tcpFallback bool
}

func New(upstreams []string, timeout time.Duration, tcpFallback bool) *Resolver {
	return &Resolver{
		udpClient:   &dns.Client{Net: "udp", Timeout: timeout},
		tcpClient:   &dns.Client{Net: "tcp", Timeout: timeout},
		upstreams:   append([]string(nil), upstreams...),
		tcpFallback: tcpFallback,
	}
}

func (r *Resolver) Resolve(request *dns.Msg) (*dns.Msg, error) {
	var lastError error
	for _, upstream := range r.upstreams {
		response, _, err := r.udpClient.Exchange(request, upstream)
		if err != nil {
			lastError = fmt.Errorf("query upstream %s over UDP: %w", upstream, err)
			continue
		}
		if !r.tcpFallback || !response.Truncated {
			return response, nil
		}

		response, _, err = r.tcpClient.Exchange(request, upstream)
		if err == nil {
			return response, nil
		}
		lastError = fmt.Errorf("query truncated response from %s over TCP: %w", upstream, err)
	}
	return nil, fmt.Errorf("all upstream resolvers failed: %w", lastError)
}

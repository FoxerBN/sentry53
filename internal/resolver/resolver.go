package resolver

import (
	"fmt"
	"time"

	"github.com/miekg/dns"
)

type Resolver struct {
	client   *dns.Client
	upstream string
}

// * New creates a new Resolver instance with the provided upstream DNS server address.
// * It initializes a dns.Client with a 3-second timeout and returns a pointer to the Resolver.
func New(upstream string) *Resolver {
	return &Resolver{
		client: &dns.Client{
			Net:     "udp",
			Timeout: 3 * time.Second},
		upstream: upstream,
	}
}

func (r *Resolver) Resolve(request *dns.Msg) (*dns.Msg, error) {
	response, _, err := r.client.Exchange(request, r.upstream)
	if err != nil {
		return nil, fmt.Errorf("query upstream DNS: %w", err)
	}
	return response, nil
}

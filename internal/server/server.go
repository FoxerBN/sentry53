package server

import (
	"fmt"

	"github.com/miekg/dns"
	"sentry53/internal/blocklist"
	"sentry53/internal/resolver"
)

// * Server represents a DNS server that uses a blocklist to filter requests. It contains a pointer to a Blocklist instance.
type Server struct {
	blocklist *blocklist.Blocklist
	resolver  *resolver.Resolver
}

// * New creates a new Server instance with the provided blocklist. It returns a pointer to the Server.
func New(blocked *blocklist.Blocklist, upstream *resolver.Resolver) *Server {
	return &Server{
		blocklist: blocked,
		resolver:  upstream,
	}
}

// * ServeDNS implements the dns.Handler interface. It checks if the requested domain is blocked and responds accordingly.
func (s *Server) ServeDNS(w dns.ResponseWriter, request *dns.Msg) {
	// * Reject malformed requests that carry no question.
	if len(request.Question) == 0 {
		response := new(dns.Msg)
		response.SetRcode(request, dns.RcodeFormatError)
		_ = w.WriteMsg(response)
		return
	}

	// * Blocked domains are answered locally with NXDOMAIN.
	if s.blocklist.IsBlocked(request.Question[0].Name) {
		response := new(dns.Msg)
		response.SetRcode(request, dns.RcodeNameError)
		_ = w.WriteMsg(response)
		return
	}

	// * Allowed domains are forwarded upstream; a failure becomes SERVFAIL.
	response, err := s.resolver.Resolve(request)
	if err != nil {
		response = new(dns.Msg)
		response.SetRcode(request, dns.RcodeServerFailure)
	}
	_ = w.WriteMsg(response)
}

// * ListenAndServe starts the DNS server on the specified address.
// * It listens for both UDP and TCP connections and returns an error if the server stops.
func (s *Server) ListenAndServe(address string) error {
	errors := make(chan error, 2)
	go func() {
		errors <- (&dns.Server{Addr: address, Net: "udp", Handler: s}).ListenAndServe()
	}()
	go func() {
		errors <- (&dns.Server{Addr: address, Net: "tcp", Handler: s}).ListenAndServe()
	}()
	return fmt.Errorf("DNS server stopped: %w", <-errors)
}

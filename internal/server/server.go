package server

import (
	"fmt"

	"github.com/miekg/dns"
	"sentry53/internal/blocklist"
)

// * Server represents a DNS server that uses a blocklist to filter requests. It contains a pointer to a Blocklist instance.
type Server struct {
	blocklist *blocklist.Blocklist
}

// * New creates a new Server instance with the provided blocklist. It returns a pointer to the Server.
func New(blocked *blocklist.Blocklist) *Server {
	return &Server{blocklist: blocked}
}

// * ServeDNS implements the dns.Handler interface. It checks if the requested domain is blocked and responds accordingly.
func (s *Server) ServeDNS(w dns.ResponseWriter, request *dns.Msg) {
	response := new(dns.Msg)
	response.SetReply(request)

	// * Check if the request has any questions
	if len(request.Question) == 0 {
		response.SetRcode(request, dns.RcodeFormatError)
	} else if s.blocklist.IsBlocked(request.Question[0].Name) {
		response.SetRcode(request, dns.RcodeNameError)
	} else {
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

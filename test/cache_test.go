package test

import (
	"testing"
	"time"

	"github.com/miekg/dns"
	"sentry53/internal/cache"
)

func TestCacheReturnsCopyWithDecrementedTTL(t *testing.T) {
	responses := cache.New(cache.Config{MaxEntries: 1, MaxTTL: time.Minute, CleanupInterval: time.Hour})
	defer responses.Close()

	question := dns.Question{Name: "Example.COM.", Qtype: dns.TypeA, Qclass: dns.ClassINET}
	message := new(dns.Msg)
	message.SetReply(&dns.Msg{Question: []dns.Question{question}})
	message.Answer = []dns.RR{&dns.A{Hdr: dns.RR_Header{Name: question.Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60}}}
	responses.Set(question, message)

	message.Answer[0].Header().Ttl = 1
	cached := responses.Get(dns.Question{Name: "example.com.", Qtype: dns.TypeA, Qclass: dns.ClassINET})
	if cached == nil {
		t.Fatal("expected cache hit")
	}
	if ttl := cached.Answer[0].Header().Ttl; ttl < 59 || ttl > 60 {
		t.Fatalf("unexpected cached TTL: %d", ttl)
	}
	cached.Answer[0].Header().Ttl = 0
	if ttl := responses.Get(question).Answer[0].Header().Ttl; ttl == 0 {
		t.Fatal("cache returned a shared message")
	}
}

func TestCacheDoesNotStoreTTLsBelowMinimum(t *testing.T) {
	responses := cache.New(cache.Config{MaxEntries: 1, MinTTL: time.Minute, MaxTTL: time.Hour, CleanupInterval: time.Hour})
	defer responses.Close()

	question := dns.Question{Name: "example.com.", Qtype: dns.TypeA, Qclass: dns.ClassINET}
	message := new(dns.Msg)
	message.Answer = []dns.RR{&dns.A{Hdr: dns.RR_Header{Ttl: 30}}}
	responses.Set(question, message)
	if cached := responses.Get(question); cached != nil {
		t.Fatal("response below minimum TTL was cached")
	}
}

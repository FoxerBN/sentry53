package cache

import (
	"strings"
	"sync"
	"time"

	"github.com/miekg/dns"
)

type Config struct {
	MaxEntries      int
	MinTTL          time.Duration
	MaxTTL          time.Duration
	CleanupInterval time.Duration
}

type CacheKey struct {
	Name  string
	Type  uint16
	Class uint16
}

type CacheEntry struct {
	Message   *dns.Msg
	StoredAt  time.Time
	ExpiresAt time.Time
}

type Cache struct {
	mu      sync.RWMutex
	entries map[CacheKey]CacheEntry
	maxSize int
	minTTL  time.Duration
	maxTTL  time.Duration
	done    chan struct{}
	once    sync.Once
}

func New(config Config) *Cache {
	cache := &Cache{
		entries: make(map[CacheKey]CacheEntry),
		maxSize: config.MaxEntries,
		minTTL:  config.MinTTL,
		maxTTL:  config.MaxTTL,
		done:    make(chan struct{}),
	}
	go cache.cleanupLoop(config.CleanupInterval)
	return cache
}

func (c *Cache) Close() {
	c.once.Do(func() { close(c.done) })
}

func (c *Cache) Get(q dns.Question) *dns.Msg {
	key := makeKey(q)
	now := time.Now()

	c.mu.RLock()
	entry, found := c.entries[key]
	c.mu.RUnlock()
	if !found {
		return nil
	}
	if !now.Before(entry.ExpiresAt) {
		c.mu.Lock()
		if current, exists := c.entries[key]; exists && !now.Before(current.ExpiresAt) {
			delete(c.entries, key)
		}
		c.mu.Unlock()
		return nil
	}

	message := entry.Message.Copy()
	decrementTTLs(message, uint32(now.Sub(entry.StoredAt)/time.Second))
	return message
}

func (c *Cache) Set(q dns.Question, msg *dns.Msg) {
	ttl, found := smallestAnswerTTL(msg)
	if !found || ttl == 0 {
		return
	}
	duration := time.Duration(ttl) * time.Second
	if duration < c.minTTL {
		return
	}
	if duration > c.maxTTL {
		duration = c.maxTTL
	}

	now := time.Now()
	key := makeKey(q)
	c.mu.Lock()
	c.removeExpiredLocked(now)
	if _, exists := c.entries[key]; !exists && len(c.entries) >= c.maxSize {
		c.mu.Unlock()
		return
	}
	c.entries[key] = CacheEntry{
		Message:   msg.Copy(),
		StoredAt:  now,
		ExpiresAt: now.Add(duration),
	}
	c.mu.Unlock()
}

func (c *Cache) cleanupLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.mu.Lock()
			c.removeExpiredLocked(time.Now())
			c.mu.Unlock()
		case <-c.done:
			return
		}
	}
}

func (c *Cache) removeExpiredLocked(now time.Time) {
	for key, entry := range c.entries {
		if !now.Before(entry.ExpiresAt) {
			delete(c.entries, key)
		}
	}
}

func makeKey(q dns.Question) CacheKey {
	return CacheKey{Name: dns.Fqdn(strings.ToLower(q.Name)), Type: q.Qtype, Class: q.Qclass}
}

func smallestAnswerTTL(message *dns.Msg) (uint32, bool) {
	if len(message.Answer) == 0 {
		return 0, false
	}
	ttl := message.Answer[0].Header().Ttl
	for _, record := range message.Answer[1:] {
		if record.Header().Ttl < ttl {
			ttl = record.Header().Ttl
		}
	}
	return ttl, true
}

func decrementTTLs(message *dns.Msg, elapsed uint32) {
	for _, section := range [][]dns.RR{message.Answer, message.Ns, message.Extra} {
		for _, record := range section {
			header := record.Header()
			if header.Ttl > elapsed {
				header.Ttl -= elapsed
			} else {
				header.Ttl = 0
			}
		}
	}
}

package blocklist

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Blocklist struct {
	domains map[string]struct{}
}

func Load(path string) (*Blocklist, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open blocklist: %w", err)
	}
	defer file.Close()

	domains := make(map[string]struct{}, 170_000)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		domain := normalize(scanner.Text())
		if domain == "" || strings.HasPrefix(domain, "#") {
			continue
		}
		domains[domain] = struct{}{}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan blocklist: %w", err)
	}
	return &Blocklist{domains: domains}, nil
}

func (b *Blocklist) IsBlocked(domain string) bool {
	_, found := b.domains[normalize(domain)]
	return found
}

func (b *Blocklist) Len() int {
	return len(b.domains)
}

func normalize(domain string) string {
	domain = strings.ToLower(strings.TrimSpace(domain))
	return strings.TrimSuffix(domain, ".")
}

package blacklist

import (
	"bufio"
	"hash/fnv"
	"io"
	"net/http"
	"os"
	"strings"
)

type Blacklist struct {
	entries map[string]bool
}

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func FromURL(url string) (*Blacklist, error) {

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return fromReader(resp.Body)
}

func FromList(entries []string) *Blacklist {
	domains := map[string]bool{}

	for _, domain := range entries {
		domains[domain] = true
	}

	return &Blacklist{entries: domains}
}

func FromFile(filepath string) (*Blacklist, error) {

	h, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}

	return fromReader(h)
}

func fromReader(r io.Reader) (*Blacklist, error) {

	domains := map[string]bool{}

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		domain := strings.TrimSpace(scanner.Text())
		if domain == "" || strings.HasPrefix(domain, "#") {
			continue
		}
		domains[domain] = true
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return &Blacklist{entries: domains}, nil

}

func (b *Blacklist) Includes(domain string) bool {
	domain = strings.TrimSuffix(domain, ".")
	for {
		_, ok := b.entries[domain]
		if ok {
			return true
		}
		if !strings.Contains(domain, ".") {
			break
		}
		domain = domain[strings.Index(domain, ".")+1:]
	}
	return false
}

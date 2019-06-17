package blacklist

import (
	"bufio"
	"io"
	"net/http"
	"os"
	"strings"
)

type Blacklist struct {
	entries []string
}

func FromURL(url string) (*Blacklist, error) {

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return fromReader(resp.Body)
}

func FromFile(filepath string) (*Blacklist, error) {

	h, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}

	return fromReader(h)
}

func fromReader(r io.Reader) (*Blacklist, error) {

	domains := []string{}

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		domain := strings.TrimSpace(scanner.Text())
		if domain == "" || strings.HasPrefix(domain, "#") {
			continue
		}
		domains = append(domains, domain)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return &Blacklist{entries: domains}, nil

}

func (b *Blacklist) Includes(domain string) bool {
	domain = "." + strings.TrimSuffix(domain, ".")
	for _, e := range b.entries {
		if strings.HasSuffix(domain, "."+e) {
			return true
		}
	}
	return false
}

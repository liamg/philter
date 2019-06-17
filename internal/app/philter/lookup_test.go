package philter

import (
	"fmt"
	"log"
	"net"
	"testing"

	"github.com/liamg/philter/internal/app/philter/server"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/require"
)

func lookup(target string, port int) (net.IP, error) {

	c := dns.Client{}
	m := dns.Msg{}
	m.SetQuestion(target+".", dns.TypeA)
	r, t, err := c.Exchange(&m, fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return nil, err
	}
	log.Printf("Took %v", t)
	if len(r.Answer) == 0 {
		return nil, fmt.Errorf("lookup failed: no results")
	}
	for _, ans := range r.Answer {
		record := ans.(*dns.A)
		return record.A, nil
	}

	return nil, fmt.Errorf("lookup failed: unknown error")

}

func TestLookup(t *testing.T) {

	s := server.New(nil)
	defer s.Stop()

	port := 35353

	s.Start(port)

	_, err := lookup("google.com", port)
	require.NoError(t, err)

}

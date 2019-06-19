package server

import (
	"fmt"
	"net"

	"github.com/liamg/philter/internal/app/philter/blacklist"
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	blacklist *blacklist.Blacklist
	cache     *Cache
	server    *dns.Server
}

func New(blacklist *blacklist.Blacklist) *Server {
	return &Server{
		blacklist: blacklist,
		cache:     newCache(),
	}
}

func (s *Server) Stop() {
	s.server.Shutdown()
	s.cache.Close()
}

func (s *Server) Start(port int) error {
	s.cache.Start()
	dns.HandleFunc(".", s.handleRequest)
	started := make(chan error)
	var ok bool
	s.server = &dns.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Net:               "udp",
		NotifyStartedFunc: func() { ok = true; started <- nil },
	}
	var err error
	go func() {
		err = s.server.ListenAndServe()
		if !ok {
			started <- err
		}
	}()
	return <-started
}

func (s *Server) handleRequest(w dns.ResponseWriter, r *dns.Msg) {

	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false

	if r.Opcode == dns.OpcodeQuery {
		s.handleQuery(m)
	}

	w.WriteMsg(m)
}

func (s *Server) handleQuery(m *dns.Msg) {

	for _, q := range m.Question {

		answer, found := s.cache.Read(q)
		if found {
			log.Printf("Query for %s returned from cache", q.Name)
			m.Answer = answer
			continue
		}

		if q.Qtype == dns.TypeA {
			if s.blacklist != nil && s.blacklist.Includes(q.Name) {
				rr, err := dns.NewRR(fmt.Sprintf("%s A 127.0.0.1", q.Name))
				if err == nil {
					log.Printf("Query for %s blocked", q.Name)
					m.Answer = append(m.Answer, rr)
					continue
				}
			}
		} else if q.Qtype == dns.TypeAAAA {
			if s.blacklist != nil && s.blacklist.Includes(q.Name) {
				rr, err := dns.NewRR(fmt.Sprintf("%s AAAA ::1", q.Name))
				if err == nil {
					log.Printf("Query for %s blocked", q.Name)
					m.Answer = append(m.Answer, rr)
					continue
				}
			}
		}

		c := new(dns.Client)
		msg := new(dns.Msg)
		msg.SetQuestion(dns.Fqdn(q.Name), q.Qtype)
		msg.RecursionDesired = true

		r, _, err := c.Exchange(msg, net.JoinHostPort("8.8.8.8", "53"))
		if r == nil {
			log.Errorf("*** error: %s\n", err.Error())
			return
		}

		if r.Rcode != dns.RcodeSuccess {
			log.Errorf(" *** invalid answer name %s", q.Name)
			return
		}

		log.Printf("Query for %s allowed", q.Name)
		m.Answer = r.Answer
		go s.cache.Write(q, r.Answer)
	}
}

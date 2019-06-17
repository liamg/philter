package server

import (
	"sync"
	"time"

	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
)

type Cache struct {
	cache     map[string]cacheItem
	lock      sync.RWMutex
	closeChan chan struct{}
}

type cacheItem struct {
	answer  []dns.RR
	expires time.Time
}

func newCache() *Cache {
	return &Cache{
		cache:     map[string]cacheItem{},
		closeChan: make(chan struct{}),
	}
}

func (c *Cache) Start() {
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		select {
		case <-c.closeChan:
			return
		case <-ticker.C:
			c.cleanse()
		}
	}()
}

func (c *Cache) cleanse() {
	now := time.Now()
	for key, val := range c.cache {
		if now.After(val.expires) {
			c.lock.Lock()
			log.Infof("Removing expired item from cache: %s", key)
			delete(c.cache, key)
			c.lock.Unlock()
		}
	}
}

func (c *Cache) Close() {
	close(c.closeChan)
}

func (c *Cache) Write(q dns.Question, answer []dns.RR) {
	if len(answer) == 0 || answer[0].Header().Ttl == 0 {
		return
	}
	c.lock.Lock()
	defer c.lock.Unlock()
	c.cache[q.String()] = cacheItem{
		answer:  answer,
		expires: time.Now().Add(time.Duration(answer[0].Header().Ttl) * time.Second),
	}
}

func (c *Cache) Read(q dns.Question) ([]dns.RR, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	item, ok := c.cache[q.String()]
	if !ok {
		return nil, false
	}
	return item.answer, true
}

package cache

import (
	"log"
	"sync"
	"time"

	"github.com/miekg/dns"
)

// DNSCacheEntry ..
type DNSCacheEntry struct {
	msg *dns.Msg
	exp time.Time
}

// DNSCache ..
type DNSCache struct {
	servers []string
	mutex   sync.Mutex
	storage map[string]*DNSCacheEntry
}

func packUint16(i uint16) []byte { return []byte{byte(i >> 8), byte(i)} }

func cacheKey(q dns.Question) string {
	return string(append([]byte(q.Name), packUint16(q.Qtype)...))
}

// Query ..
func (c *DNSCache) Query(payload []byte) *dns.Msg {
	request := new(dns.Msg)
	e := request.Unpack(payload)
	if e != nil {
		return nil
	}
	if len(request.Question) == 0 {
		return nil
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()
	key := cacheKey(request.Question[0])
	entry := c.storage[key]
	if entry == nil {
		return nil
	}
	if time.Now().After(entry.exp) {
		delete(c.storage, key)
		return nil
	}
	entry.msg.Id = request.Id
	return entry.msg
}

// Store ...
func (c *DNSCache) Store(payload []byte) {
	resp := new(dns.Msg)
	e := resp.Unpack(payload)
	if e != nil {
		return
	}
	if resp.Rcode != dns.RcodeSuccess {
		return
	}
	if len(resp.Question) == 0 || len(resp.Answer) == 0 {
		return
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()
	key := cacheKey(resp.Question[0])
	log.Printf("cache DNS response for %s", key)
	c.storage[key] = &DNSCacheEntry{
		msg: resp,
		exp: time.Now().Add(time.Duration(resp.Answer[0].Header().Ttl) * time.Second),
	}
}

// NewDNSCache ...
func NewDNSCache() *DNSCache {

	return &DNSCache{
		storage: make(map[string]*DNSCacheEntry),
	}
}

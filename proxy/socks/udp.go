package socks

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"lwip2socks/common/dns"
	"lwip2socks/common/dns/cache"
	"lwip2socks/core"
	"net"
	"strconv"
	"sync"
	"time"
)

type udpHandler struct {
	sync.Mutex

	proxyHost   string
	proxyPort   uint16
	remoteAddrs map[core.UDPConn]*net.UDPAddr // UDP relay server addresses
	udpSocks    map[core.UDPConn]net.Conn
	timeout     time.Duration

	dnsCache *cache.DNSCache
}

// NewUDPHandler ...
func NewUDPHandler(proxyHost string, proxyPort uint16, timeout time.Duration, dnsCache *cache.DNSCache) core.UDPConnHandler {
	return &udpHandler{
		proxyHost:   proxyHost,
		proxyPort:   proxyPort,
		dnsCache:    dnsCache,
		timeout:     timeout,
		remoteAddrs: make(map[core.UDPConn]*net.UDPAddr, 8),
		udpSocks:    make(map[core.UDPConn]net.Conn, 8),
	}
}

func settimeout(con net.Conn, second int) {
	readTimeout := time.Duration(second) * time.Second
	v := time.Now().Add(readTimeout)
	con.SetReadDeadline(v)
	con.SetWriteDeadline(v)
	con.SetDeadline(v)
}

//Connect ...
func (h *udpHandler) Connect(conn core.UDPConn, target *net.UDPAddr) error {
	dest := net.JoinHostPort(h.proxyHost, strconv.Itoa(int(h.proxyPort)))
	netcc, err := net.Dial("udp", dest)
	if err != nil {
		conn.Close()
		log.Println("socks connect failed:", err, dest)
		return err
	}

	if target != nil {
		h.Lock()
		h.remoteAddrs[conn] = target
		h.Unlock()
	}

	h.Lock()
	v, ok := h.udpSocks[conn]
	if ok {
		v.Close()
		delete(h.udpSocks, conn)
	}
	h.udpSocks[conn] = netcc
	h.Unlock()

	settimeout(netcc, 120) // set timeout

	go h.fetchSocksData(conn)

	return nil
}

func (h *udpHandler) fetchSocksData(conn core.UDPConn) {
	buf := core.NewBytes(core.BufSize)
	defer func() {
		core.FreeBytes(buf)
		h.Close(conn)
	}()

	h.Lock()
	netcc, ok := h.udpSocks[conn]
	newTarget, ok0 := h.remoteAddrs[conn]
	h.Unlock()
	if !ok {
		log.Println("can not find socks <-->", conn.LocalAddr().String())
		return
	}
	if !ok0 {
		log.Println("can not remote address <-->", conn.LocalAddr().String())
		return
	}

	n, err := netcc.Read(buf)
	if err != nil {
		log.Println(err, "read from socks failed")
		return
	}

	raw := buf[:n]
	conn.WriteFrom(raw, newTarget)
	if newTarget.Port == dns.COMMON_DNS_PORT {
		h.dnsCache.Store(raw)
	}

}

func packUDPHeader(b []byte, addr net.Addr) []byte {

	var out bytes.Buffer
	n := len([]byte(addr.String()))
	sz := make([]byte, 4)
	binary.BigEndian.PutUint32(sz, uint32(n))

	out.Write(sz)
	out.Write([]byte(addr.String()))
	out.Write(b)
	return out.Bytes()
}

// ReceiveTo will be called when data arrives from TUN.
func (h *udpHandler) ReceiveTo(conn core.UDPConn, data []byte, addr *net.UDPAddr) error {
	h.Lock()
	udpsocks, ok1 := h.udpSocks[conn]
	remoteAddr, ok2 := h.remoteAddrs[conn]
	h.Unlock()

	if !ok1 {
		dest := net.JoinHostPort(h.proxyHost, strconv.Itoa(int(h.proxyPort)))
		udpsocks, err := net.Dial("udp", dest)
		if err != nil {
			h.Close(conn)
			return err
		}
		h.Lock()
		h.udpSocks[conn] = udpsocks
		h.remoteAddrs[conn] = addr
		h.Unlock()
	}
	if !ok2 {
		remoteAddr = addr
		h.Lock()
		h.remoteAddrs[conn] = addr
		h.Unlock()
	}

	if !ok1 {
		go h.fetchSocksData(conn)
	}

	if answer := h.dnsCache.Query(data); answer != nil {
		var buf [1024]byte
		resp, _ := answer.PackBuffer(buf[:])
		_, err := conn.WriteFrom(resp, addr)
		h.Close(conn)
		if err != nil {
			estring := fmt.Sprintf("write dns answer failed: %v", err)
			log.Println(estring)
			return errors.New(estring)
		}
		return nil
	}

	n, err := udpsocks.Write(packUDPHeader(data, remoteAddr))
	if err != nil {
		h.Close(conn)
		log.Println("write to proxy failed", err)
		return errors.New("write to proxy failed")
	}
	log.Println("write bytes n", n)

	return nil
}

func (h *udpHandler) Close(conn core.UDPConn) {
	conn.Close()

	h.Lock()
	defer h.Unlock()

	if c, ok := h.udpSocks[conn]; ok {
		c.Close()
		delete(h.udpSocks, conn)
	}
	if _, ok := h.remoteAddrs[conn]; ok {
		delete(h.remoteAddrs, conn)
	}
}

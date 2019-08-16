package socks

import (
	"io"
	"net"
	"strconv"
	"sync"

	"golang.org/x/net/proxy"

	"lwip2socks/core"
)

type tcpHandler struct {
	sync.Mutex

	proxyHost string
	proxyPort uint16
}

//NewTCPHandler ...
func NewTCPHandler(proxyHost string, proxyPort uint16) core.TCPConnHandler {
	return &tcpHandler{
		proxyHost: proxyHost,
		proxyPort: proxyPort,
	}
}

type direction byte

const (
	dirUplink direction = iota
	dirDownlink
)

func (h *tcpHandler) Handle(conn net.Conn, target *net.TCPAddr) error {
	dialer, err := proxy.SOCKS5("tcp", core.ParseTCPAddr(h.proxyHost, h.proxyPort).String(), nil, nil)
	if err != nil {
		conn.Close()
		return err
	}

	// Replace with a domain name if target address IP is a fake IP.
	var targetHost string
	targetHost = target.IP.String()

	dest := net.JoinHostPort(targetHost, strconv.Itoa(target.Port))

	c, err := dialer.Dial(target.Network(), dest)
	if err != nil {
		conn.Close()
		return err
	}

	go func() {
		defer func() {
			conn.Close()
			c.Close()
		}()
		upCh := make(chan int, 1)
		go func() {
			_, err = io.Copy(conn, c)
			upCh <- 1
		}()

		_, err = io.Copy(c, conn)
		<-upCh
		close(upCh)
	}()
	return nil
}

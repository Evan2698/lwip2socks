module lwip2socks

go 1.12

replace golang.org/x/net => github.com/golang/net latest

replace golang.org/x/crypto => github.com/golang/crypto latest

replace golang.org/x/text => github.com/golang/text latest

replace golang.org/x/sys => github.com/golang/sys latest

require (
	github.com/eycorsican/go-tun2socks latest
	github.com/miekg/dns latest
	golang.org/x/net latest
	golang.org/x/text latest
)

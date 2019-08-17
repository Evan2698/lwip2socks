module lwip2socks

go 1.12

replace golang.org/x/net => github.com/golang/net v0.0.0-20190813141303-74dc4d7220e7

replace golang.org/x/crypto => github.com/golang/crypto v0.0.0-20190701094942-4def268fd1a4

replace golang.org/x/text => github.com/golang/text v0.3.2

replace golang.org/x/sys => github.com/golang/sys v0.0.0-20190813064441-fde4db37ae7a

require (
	github.com/eycorsican/go-tun2socks v1.16.2
	github.com/miekg/dns v1.1.15
	golang.org/x/net v0.0.0-20190404232315-eb5bcb51f2a3
	golang.org/x/text v0.3.0
)

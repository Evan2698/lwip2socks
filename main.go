package main

import (
	"lwip2socks/common/dns/cache"
	"lwip2socks/core"
	"lwip2socks/proxy/socks"
)

func main() {

	dnsCache := cache.NewDNSCache()
	core.RegisterOutputFn(func(data []byte) (int, error) {
		return 1, nil

		//return tunDev.Write(data)
	})

	core.RegisterTCPConnHandler(socks.NewTCPHandler("127.0.0.1", 1080))
	core.RegisterUDPConnHandler(socks.NewUDPHandler("127.0.0.1", 1080, 120, dnsCache))

	/*
		go func() {
			_, err := io.CopyBuffer(lwipWriter, tunDev, make([]byte, MTU))
			if err != nil {
				log.Fatalf("copying data failed: %v", err)
			}
		}()
	*/
}

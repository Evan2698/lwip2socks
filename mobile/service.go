package mobile

import (
	"log"
	"lwip2socks/common/dns/cache"
	"lwip2socks/core"
	"lwip2socks/proxy/socks"
	"os"
	"time"
)

var ginterrupt bool

const (
	mtu = 1500
)

var dnsCache = cache.NewDNSCache()

var lwipWriter = core.NewLWIPStack()

// StartService ...
func StartService(fd int, proxy string, dns string) bool {
	f := os.NewFile(uintptr(fd), "")

	core.RegisterTCPConnHandler(socks.NewTCPHandler("127.0.0.1", 1080))
	core.RegisterUDPConnHandler(socks.NewUDPHandler("127.0.0.1", 1080, 120, dnsCache))

	core.RegisterOutputFn(func(data []byte) (int, error) {
		return f.Write(data)
	})

	ginterrupt = false

	go func() {
		buf := core.NewBytes(mtu)
		defer func() {
			core.FreeBytes(buf)
			lwipWriter.Close()
			f.Close()
		}()
		for {
			if ginterrupt {
				break
			}

			n, err := f.Read(buf)
			if err != nil {
				log.Println("read from tun failed,", err)
				break
			}

			n, err = lwipWriter.Write(buf[:n])
			if err != nil {
				log.Println("write to stack failed,", err)
				continue
			}
		}

	}()

	return true

}

// StopService ...
func StopService() {
	ginterrupt = true
	time.Sleep(4 * time.Second)
}

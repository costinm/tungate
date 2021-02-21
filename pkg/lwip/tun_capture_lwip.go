package lwip

import (
	"io"
	"log"
	"net"

	"github.com/costinm/tungate"
	"github.com/eycorsican/go-tun2socks/core"
)

// Temp: should move to android or higher level package.

const (
	MTU = 1500
)

// LWIPTun adapts the LWIP interfaces - in particular UDPConn
type LWIPTun struct {
	lwip       core.LWIPStack
	tcpHandler tungate.TUNHandler
	udpHandler tungate.UDPHandler
}

func (t *LWIPTun) Connect(conn core.UDPConn, target *net.UDPAddr) error {
	return nil
}

func (t *LWIPTun) ReceiveTo(conn core.UDPConn, data []byte, addr *net.UDPAddr) error {
	//t.udpHandler.HandleUdp()
	return nil
}

func (t *LWIPTun) Handle(conn net.Conn, target *net.TCPAddr) error {
	// Must return - TCP con will be moved to connected after return.
	// err will abort. While this is executing, will stay in connected
	// TODO: extra param to do all processing and do the proxy in background.
	go t.tcpHandler.HandleTUN(conn, target)
	return nil
}

func (nt *LWIPTun) WriteTo(data []byte, dst *net.UDPAddr, src *net.UDPAddr) (int, error) {
	return 0, nil
}

func NewTUNFD(tunDev io.ReadWriteCloser, handler tungate.TUNHandler, udpNat tungate.UDPHandler) *LWIPTun {

	lwip := core.NewLWIPStack()

	t := &LWIPTun{
		lwip: lwip,
		tcpHandler: handler,
		udpHandler: udpNat,
	}

	core.RegisterTCPConnHandler(t)
	//core.RegisterTCPConnHandler(redirect.NewTCPHandler("127.0.0.1:5201"))

	core.RegisterUDPConnHandler(t)
	
	core.RegisterOutputFn(func(data []byte) (int, error) {
		//log.Println("ip2tunW: ", len(data))
		return tunDev.Write(data)
	})

	// Copy packets from tun device to lwip stack, it's the main loop.
	go func() {
		ba := make([]byte, 10 *MTU)
		for  {
			n, err := tunDev.Read(ba)
			if err != nil {
				log.Println("Err tun", err)
				return
			}
			//log.Println("tun2ipR: ", n)
			_, err = lwip.Write(ba[0:n])
			if err != nil {
				log.Println("Err lwip", err)
				return
			}
		}
	}()

	return t
}

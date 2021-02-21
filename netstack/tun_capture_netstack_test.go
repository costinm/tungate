package netstack

import (
	"bytes"

	"github.com/costinm/tungate"
	"github.com/google/netstack/tcpip"
	"github.com/google/netstack/tcpip/adapters/gonet"
	"github.com/google/netstack/tcpip/buffer"
	"github.com/google/netstack/tcpip/header"
	"github.com/google/netstack/tcpip/network/ipv4"
	"github.com/google/netstack/tcpip/network/ipv6"
	"github.com/google/netstack/tcpip/transport/udp"

	"log"
	"net"
	"os"
	"testing"
)

type EchoHandler struct {
	udpW tungate.UdpWriter
}

func (e *EchoHandler) HandleTUN(conn net.Conn, target *net.TCPAddr) error {
	b := make([]byte, 2048)
	for {
		n, err := conn.Read(b)
		if err != nil {
			conn.Close()
		}
		conn.Write(b[0:n])
	}
}

func (e *EchoHandler) HandleUdp(dstAddr net.IP, dstPort uint16,
	localAddr net.IP, localPort uint16, data []byte) {
	e.udpW.WriteTo(data,
		&net.UDPAddr{IP: localAddr, Port: int(localPort)},
		&net.UDPAddr{IP: dstAddr, Port: int(dstPort)})
}

// Checks we can open the tun directly.
func TestTcpCaptureReal(t *testing.T) {
	handlers := &EchoHandler{}

	fd, err := tungate.OpenTun("dmesh1")
	if err != nil {
		log.Fatal("Failed to open tubn", err)
	}

	tun1 := NewTUNFD(fd, handlers, handlers)
	tun := tun1.(*NetstackTun)

	handlers.udpW = tun
	if err != nil {
		t.Skip("TUN can't be opened, make sure it is setup", err)
	}

	t.Run("external", func(t *testing.T) {
		testTcpEchoLocal(t, tun, "10.10.0.2", 5227)
	})
	t.Run("external6", func(t *testing.T) {
		testTcpEchoLocal(t, tun, "fd00::2", 5227)
	})

	t.Run("local", func(t *testing.T) {
		testTcpEcho(t, tun, "127.0.0.1", 2000)
		testTcpEcho(t, tun, "127.0.0.1", 3000)
	})
}

// Using the net stack for testing.
func testTcpEchoLocal(t *testing.T, tn *NetstackTun, addr string, port uint16) {
	ip4, err := net.ResolveIPAddr("ip", addr)
	if err != nil {
		t.Fatal("Can't resolve ", err)
	}

	c1, err := net.DialTCP("tcp", nil, &net.TCPAddr{IP: ip4.IP, Port: int(port)})
	if err != nil {
		t.Fatal("Failed to dial ", err)
	}
	defer c1.Close()

	c1.Write([]byte("GET / HTTP/1.1\nHost:www.webinf.info\n\n"))

	data := make([]byte, 1024)
	n, err := c1.Read(data[0:])
	if err != nil {
		t.Fatal(err)
	}
	log.Println("Recv: ", c1.RemoteAddr(), string(data[:n]))
}


func TestTcpCapture(t *testing.T) {
	link, _, _ := initPipeLink()

	Dump = true

	handlers := &EchoHandler{}

	tun := NewTunCapture(&link, handlers, handlers, true)
	handlers.udpW = tun

	t.Run("v4", func(t *testing.T) {
		testTcpEcho(t, tun, "127.0.0.1", 2000)
	})

	t.Run("v6", func(t *testing.T) {
		testTcpEcho6(t, tun, 3000)
	})
}
func TestUDP4(t *testing.T) {
	link, linkw, linkr := initPipeLink()
	Dump = true

	handlers := &EchoHandler{}

	tun := NewTunCapture(&link, handlers, handlers, true)
	handlers.udpW = tun
	srcAddr := []byte{10, 0, 0, 1}
		dstAddr := []byte{127, 0, 0, 1}

		ip62 := makeV4UDP([]byte("Hello"),
			tcpip.Address(srcAddr),
			tcpip.Address(dstAddr), 1000, 2000)

		go linkw.Write(ip62)

		data := make([]byte, 2048)
		n, err1 := linkr.Read(data)
		log.Println("received 3", n, err1)
		}

func TestUDP6(t *testing.T) {
	link, linkw, linkr := initPipeLink()
	Dump = true

	handlers := &EchoHandler{}

	tun := NewTunCapture(&link, handlers, handlers, true)
	handlers.udpW = tun

	srcAddr := net.IPv6loopback
	//dstAddr := net.IPv6loopback

	dst, _ := net.ResolveIPAddr("ip", "2001:470:1f04:429::2")

	ip62 := makeV6UDP([]byte("Hello"),
		tcpip.Address(dst.IP),
		tcpip.Address(srcAddr), 1000, 2000)

	go linkw.Write(ip62)

	data := make([]byte, 2048)
	n, err1 := linkr.Read(data)
	log.Println("received 3", n, err1)
}


// init a network interface backed by 2 os pipes, linkr and linkw.
// 'link' is the netstack link
func initPipeLink() (tcpip.LinkEndpointID, *os.File, *os.File) {
	lr, stw, _ := os.Pipe()
	pr, lw, _ := os.Pipe()

	link := NewReaderWriterLink(stw, pr, &Options{MTU: 1600})
	return link, lw, lr
}

// Format a UDP packet, with IP6 header.
// This is an example of how to create UDP and IP packets using the stack.
// In practical use, the gonet interface is much easier.
func makeV4UDP(payload []byte, src, dst tcpip.Address, srcport, dstport uint16) []byte {
	// Allocate a buffer for data and headers.
	buf := buffer.NewView(header.UDPMinimumSize + header.IPv4MinimumSize + len(payload))

	// payload at the end
	copy(buf[len(buf)-len(payload):], payload)

	// Initialize the IP header.
	ip := header.IPv4(buf)
	ip.Encode(&header.IPv4Fields{
		TotalLength: uint16(len(buf)),
		Protocol:    uint8(udp.ProtocolNumber),
		TTL:         65,
		SrcAddr:     src,
		DstAddr:     dst,
		IHL:         header.IPv4MinimumSize,
	})

	// Initialize the UDP header.
	u := header.UDP(buf[header.IPv4MinimumSize:])
	u.Encode(&header.UDPFields{
		SrcPort: srcport,
		DstPort: dstport,
		Length:  uint16(header.UDPMinimumSize + len(payload)),
	})

	// Calculate the UDP pseudo-header checksum.
	xsum := header.Checksum([]byte(src), 0)
	xsum = header.Checksum([]byte(dst), xsum)
	xsum = header.Checksum([]byte{0, uint8(udp.ProtocolNumber)}, xsum)

	// Calculate the UDP checksum and set it.
	length := uint16(header.UDPMinimumSize + len(payload))
	xsum = header.Checksum(payload, xsum)
	u.SetChecksum(^u.CalculateChecksum(xsum, length))

	return buf
}

func makeV6UDP(payload []byte, src, dst tcpip.Address, srcport, dstport uint16) []byte {
	// Allocate a buffer for data and headers.
	buf := buffer.NewView(header.UDPMinimumSize + header.IPv6MinimumSize + len(payload))

	// payload at the end
	copy(buf[len(buf)-len(payload):], payload)

	// Initialize the IP header.
	ip := header.IPv6(buf)
	ip.Encode(&header.IPv6Fields{
		PayloadLength: uint16(header.UDPMinimumSize + len(payload)),
		NextHeader:    uint8(udp.ProtocolNumber),
		HopLimit:      65,
		SrcAddr:       src,
		DstAddr:       dst,
	})

	// Initialize the UDP header.
	u := header.UDP(buf[header.IPv6MinimumSize:])
	u.Encode(&header.UDPFields{
		SrcPort: srcport,
		DstPort: dstport,
		Length:  uint16(header.UDPMinimumSize + len(payload)),
	})

	// Calculate the UDP pseudo-header checksum.
	xsum := header.Checksum([]byte(src), 0)
	xsum = header.Checksum([]byte(dst), xsum)
	xsum = header.Checksum([]byte{0, uint8(udp.ProtocolNumber)}, xsum)

	// Calculate the UDP checksum and set it.
	length := uint16(header.UDPMinimumSize + len(payload))
	xsum = header.Checksum(payload, xsum)
	u.SetChecksum(^u.CalculateChecksum(xsum, length))

	return buf
}

func testTcpEcho6(t *testing.T, tn *NetstackTun, port uint16) {
	ip6 := net.IPv6loopback

	c1, err := gonet.DialTCP(tn.IPStack, tcpip.FullAddress{
		// Doesn't seem to work - regardless of routes.
		//Addr: tcpip.Address(net.IPv4(10, 12, 0, 2).To4()),
		Addr: tcpip.Address(ip6),
		Port: port,
		// If NIC is missing, it hangs.
		//NIC:  1,
	}, ipv6.ProtocolNumber)

	if err != nil {
		t.Fatal("Failed to dial ", err)
	}

	var cc net.Conn
	cc = c1
	c1.Write([]byte(echoMsg))
	if cw, ok := cc.(tungate.CloseWriter); ok {
		cw.CloseWrite()
	}
	data := make([]byte, 1024)
	n, err := c1.Read(data[0:])
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data[0:n], []byte(echoMsg)) {
		t.Fatal("Failed to receive ", string(data[0:n]))
	}
	c1.Close()
}

const echoMsg = "GET / HTTP/1.1\n\n"

// This is going trough the stack, but not using the link
//
// Using the gonet stack for testing. This doesn't require a TUN device.
// gonet implements the normal go.net interfaces, but using the soft stack.
func testTcpEcho(t *testing.T, tn *NetstackTun, addr string, port uint16) {
	ip4, err := net.ResolveIPAddr("ip", addr)
	if err != nil {
		t.Fatal("Can't resolve ", err)
	}

	c1, err := gonet.DialTCP(tn.IPStack, tcpip.FullAddress{
		// Doesn't seem to work - regardless of routes.
		//Addr: tcpip.Address(net.IPv4(10, 12, 0, 2).To4()),
		Addr: tcpip.Address(ip4.IP.To4()),
		Port: port,
	}, ipv4.ProtocolNumber)
	if err != nil {
		t.Fatal("Failed to dial ", err)
	}

	TcpEchoTest(c1)
}

// tests on the 'echo' server. c1 is an established connection to the echo server, possibly
// using intermediaries.
func TcpEchoTest(c1 net.Conn) {

	c1.Write([]byte("GET / HTTP/1.1\n\n"))

	data := make([]byte, 1024)
	n, _ := c1.Read(data[0:])
	log.Println("Recv: ", string(data[:n]))

	c1.Close()
}


//func TestUdpEcho(t *testing.T) {
//	l, err := net.ListenUDP("udp4", &net.UDPAddr{Port: 1999})
//	if err != nil {
//		t.Fatal(err)
//	}
//	ip, err := net.ResolveIPAddr("ip", "h.webinf.info")
//	if err != nil {
//		t.Fatal(err)
//	}
//	go func() {
//		for {
//			l.WriteToUDP([]byte("Hi1"), &net.UDPAddr{Port: 5228, IP: ip.IP})
//			time.Sleep(4 * time.Second)
//		}
//	}()
//	for {
//		b := make([]byte, 1600)
//		n, addr, _ := l.ReadFromUDP(b)
//		log.Println("RCV: ", addr, string(b[0:n]))
//	}
//
//}

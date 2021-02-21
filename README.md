# TUN-based gateway.

This package is an extension to the costinm/ugate package, providing 
capture using a TUN device, primarily used for Android. 

In regular Linux it is also possible to use a TUN and routes, but 
iptables is significantly faster.

To use the TUN we need a TCP/IP stack in user space - the 
gateway works at L3/L7, multiplexing all TCP/HTTP connections. 

The default implementation uses LWIP. 

Originally we used a fork of google netstack, with few extensions to 
capture all ports. Netstack is deprecated now, replaced by gvisor. 

This repo includes all 3 variants - current testing shows 
gvisor to be faster but it has far more dependencies and is 
more complex/harder to integrate.

`make all` will build all 3, with a minimal proxy config, and 
show the size and stripped size. Current numbers on stripped:

- No TUN: 5.7 M
- LWIP: 6 M
- Netstack: 6.7 M
- GVisor: 18 M


``` 
# setup a tun device ('dmesh1' in this example), see the script for example

tun, err := tuntransport.NewTUN(&tuntransport.TunConfig{
		Name: "dmesh1",
		UDPHandler: ug,
		TCPHandler: ug,
	})

 	

```

The tun implements the UDPWriter interface to send packets via TUN.
The UDPHandler and TCPHandler are called for each UDP packet or connection. 

Note that currently the Conn is from the perspective of the stack accepting a connection:
LocalAddr() returns the destination address of the connection that originated on the 
local machine and is routed via TUN. The connection appears to be an 'accepted' connection.
RemoteAddr() is hardcoded to 10.12.0.1, which is the internal address assigned by the stack 
to the tun.

# Performance

Right now iperf3 shows ~450Mbps even on localhost, compared with 23Gbps with iptables or 
socks. It is good enough for Android NAT - but so far not best option on server. 

lwip: 
- 238Mbps - basic tunredir app, LWIP
- 40Mbps - gotun2socks with a minimal go tcp stack
- 480Mbps - old netstack, golang - 'direct fd' not working.

# Debugging 

```

ip route show table all

ip roule show

ip route get 1.2.3.4 mark 1338
```

GVisor stack, emulating iptables: 
- rawfile_unsafe.go BlockingReadv or the channel
- nic.go DeliverNetworkPacket
- ipv4.go HandlePacket - may call IPTables.Check
- back to nic.go DeliverTransportPacket
- transport_demuxer.go deliverRawPacket - handle "Raw" endpoints
- transport_demuxer.go deliverPacket -> endpointsByNIC.handlePacket
- tcp.go QueuePacket to the tcp endpoint
- Background loops:
    - accept.go protocolListenLoop -> handleSynSegment
    
Ports:
- tcpip/ports.go allocatedPorts - bind reserves the port
- Listen -> transport_demuxer.singleRegisterEndpoint by NIC, epsByNIC.endpoints, multiPortEndpoints


UDP:
```
gvisor.dev/gvisor/pkg/tcpip/stack.(*endpointsByNIC).handlePacket at transport_demuxer.go:189
gvisor.dev/gvisor/pkg/tcpip/stack.(*transportDemuxer).deliverPacket at transport_demuxer.go:578
gvisor.dev/gvisor/pkg/tcpip/stack.(*NIC).DeliverTransportPacket at nic.go:799
gvisor.dev/gvisor/pkg/tcpip/network/ipv4.(*endpoint).handlePacket at ipv4.go:754
gvisor.dev/gvisor/pkg/tcpip/network/ipv4.(*endpoint).HandlePacket at ipv4.go:575
gvisor.dev/gvisor/pkg/tcpip/stack.(*NIC).DeliverNetworkPacket at nic.go:722
gvisor.dev/gvisor/pkg/tcpip/link/channel.(*Endpoint).InjectLinkAddr at channel.go:190
gvisor.dev/gvisor/pkg/tcpip/link/channel.(*Endpoint).InjectInbound at channel.go:185
github.com/costinm/tungate.NewGvisorTUN.func1 at tun_capture_gvisor.go:220
runtime.goexit at asm_amd64.s:1374
 - Async stack trace
github.com/costinm/tungate.NewGvisorTUN at tun_capture_gvisor.go:207

```

TODO:
- how are packets recycled ?
- set iptable rule
  - stack/iptables_types

# Issues

On Android everything seems to work great.

On Linux, capturing outbound and inbound works. I still can't figure out how to capture
localhost traffic - the local route table takes priority.

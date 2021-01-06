package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/costinm/tungate"
	"github.com/costinm/tungate/pkg/gvisor"
	"github.com/costinm/ugate"
)

// Similar with the sample micro-gate, but adding a TUN capture.
// Used to experiment with TUN instead of iptables capture.

func main() {
	config := ugate.NewConf(".")

	auth := ugate.NewAuth(config, "", "h.webinf.info")

	cfg := &ugate.GateCfg{
		BasePort: 15000,
	}

	data, err := ioutil.ReadFile("h2gate.json")
	if err != nil {
		json.Unmarshal(data, cfg)
	}

	// By default, pass through using net.Dialer
	ug := ugate.NewGate(&net.Dialer{}, auth)
	ugate.Reversed = true

	fd, err := tungate.OpenTun("dmesh1")
	if err != nil {
		log.Fatal("Failed to open tubn", err)
	}

	tun := gvisor.NewTUNFD(fd,ug, ug)
	ug.TUNUDPWriter = tun

	log.Println("TUN started ", tun)

		// direct TCP connect to local iperf3 and fortio (or HTTP on default port)
	ug.Add(&ugate.ListenerConf{
		Port: 15101,
		Remote: "localhost:5201",
	})

	log.Println("Started debug on 15020, UID/GID", os.Getuid(), os.Getegid())
	err = http.ListenAndServe(":15020", nil)
	if err != nil {
		log.Fatal(err)
	}

}

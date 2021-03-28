package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/costinm/ugate"
)

// Similar with the sample micro-gate, but adding a TUN capture.
// Used to experiment with TUN instead of iptables capture.

func main() {
	config := ugate.NewConf(".")

	auth := ugate.NewAuth(config, "", "h.webinf.info")

	cfg := &ugate.GateCfg{
		BasePort: 12000,
	}

	data, err := ioutil.ReadFile("h2gate.json")
	if err != nil {
		json.Unmarshal(data, cfg)
	}

	// By default, pass through using net.Dialer
	ug := ugate.NewGate(&net.Dialer{}, auth)

	// direct TCP connect to local iperf3 and fortio (or HTTP on default port)
	ug.Add(&ugate.ListenerConf{
		Port: 12011,
		Remote: "localhost:5201",
	})

	log.Println("Started debug on 12019, UID/GID", os.Getuid(), os.Getegid())
	err = http.ListenAndServe(":12019", nil)
	if err != nil {
		log.Fatal(err)
	}
}

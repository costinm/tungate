#!/bin/bash

test_setup() {
  pkill iperf3
  iperf3 -s &
  echo $! >build/iperf.pid

  export TUNUSER=${USER}
  sudo TUNDEV=dmesh1 TUNNET=10.12 TUNNETFW=10.13.0.0/16 TUNFW=134 ./setup.sh setup
  sudo TUNDEV=dmesh2 TUNNET=10.14 TUNNETFW=10.15.0.0/16 TUNFW=135 ./setup.sh setup
  sudo TUNDEV=dmesh3 TUNNET=10.16 TUNNETFW=10.17.0.0/16 TUNFW=136 ./setup.sh setup
}

test_app() {
  pkill tun_lwip
  pkill tun_netstack
  pkill tun_gvisor
  ./build/tun_lwip &
  echo $! >build/lwip.pid

  ./gvisor/build/tun_gvisor &
  echo $! >build/gvisor.pid

  ./netstack/build/tun_netstack &
  echo $! >build/netstack.pid
}

test_run() {
  # Direct access
  iperf3 -c localhost:5201
  # Via ugate, whitebox TCP capture
  iperf3 -c localhost:12111

  # Via routes
  iperf3 -c 10.13.0.1:12111
  iperf3 -c 10.15.0.1:12211
  iperf3 -c 10.17.0.1:15311
}

test_cleanup() {
  kill $(cat build/iperf.pid)

  TUNDEV=dmesh1 TUNNET=10.12 TUNNETFW=10.13.0.0/16 TUNFW=134 sudo setup.sh clean
  TUNDEV=dmesh2 TUNNET=10.14 TUNNETFW=10.15.0.0/16 TUNFW=135 sudo setup.sh clean
  TUNDEV=dmesh3 TUNNET=10.16 TUNNETFW=10.17.0.0/16 TUNFW=136 sudo setup.sh clean
}

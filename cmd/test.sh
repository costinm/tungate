#!/bin/bash

function start() {
  pkill iperf3
  iperf3 -s &
  echo $! > build/iperf.pid
}

function stop() {
 kill $(cat build/iperf.pid)
}

start
./build/tun_lwip &
echo $! > build/lwip.pid

iperf3 -c localhost:15101

#!/bin/sh

# Setup the TUN device for capture.
#
# Must be run as root.

export TUNUSER=${TUNUSER:-istio-proxy}
export TUNDEV=${TUNDEV:-dmesh}

# Net for the tun device. Localhost is 0.1
export TUNNET=${TUNNET:-10.11}
# Base for the tags, 7 and 8 added ( 1337, 1338 )
export TUNFW=${TUNFW:-133}
# Net to route via device
export TUNNETFW=${TUNNETFW:-10.10.0.0/16}

set -x
env
# Create a TUN device.
# Default 'dmesh', user 'istio-proxy', IP 10.12.0.1
setupTUN() {
  ip tuntap add dev ${TUNDEV} mode tun user ${TUNUSER} group ${TUNUSER}
  ip addr add ${TUNNET}.0.1/16 dev ${TUNDEV}
  # No IP6 address - confuses linux
  ip link set ${TUNDEV} up

  # Don't remember why this was required
  echo 2 > /proc/sys/net/ipv4/conf/${TUNDEV}/rp_filter
  sysctl -w net.ipv4.ip_forward=1
}

# Setup routes:
# - add a routing table (1338) to dmesh
# - all packets with mark 1338 will use the new routing table
# - route 10.10.0.0/16 via the tun
#
#
setup() {

  # For iptables-based capture to TUN

  # For iptables capture/marks:
  ip route add default dev ${TUNDEV} table ${TUNFW}8
  ip rule add  fwmark ${TUNFW}8 priority 10  lookup ${TUNFW}8

  # Route various ranges to dmesh1 - the gate can't initiate its own connections
  # to those ranges. Service VIPs can also use this simpler model.
  #ip route add fd::/8 dev ${TUNDEV}
  ip route add ${TUNNETFW} dev ${TUNDEV}

  # 1337 means deliver to local host
  ip route add local 0.0.0.0/0 dev lo table ${TUNFW}7
  ip rule add fwmark ${TUNFW}7 lookup ${TUNFW}7
  # Anything from the TUN will be sent to localhost
  # That means packets injected into TUN.
  ip rule add iif ${TUNDEV} lookup ${TUNFW}7
  #ip route add local ::/0 dev lo table ${TUNFW}7
}

cleanup() {
  # App must be stopped
  ip tuntap del dev ${TUNDEV} mode tun

  ip rule delete  fwmark ${TUNFW}8 priority 10  lookup ${TUNFW}8
  ip route del default dev ${TUNDEV} table ${TUNFW}8

  ip rule del fwmark ${TUNFW}7 lookup ${TUNFW}7
  ip rule del iif ${TUNDEV} lookup ${TUNFW}7
  ip route del local 0.0.0.0/0 dev lo table ${TUNFW}7
}


stop() {
  iptables -t mangle -D OUTPUT -j DMESH_MANGLE_OUT
  iptables -t mangle -D PREROUTING -j DMESH_MANGLE_PRE

  iptables -t mangle -F DMESH_MANGLE_OUT 2>/dev/null
  iptables -t mangle -X DMESH_MANGLE_OUT 2>/dev/null

  iptables -t mangle -F DMESH_MANGLE_PRE 2>/dev/null
  iptables -t mangle -X DMESH_MANGLE_PRE 2>/dev/null
}

# Setup will create route-based rules for the NAT.
# This function intercepts additional packets, using
# Istio-style rules.
start() {
  GID=$(id -g ${TUNUSER})

  # -j MARK only works in mangle !
  iptables -t mangle -N DMESH_MANGLE_PRE
  iptables -t mangle -A PREROUTING -j DMESH_MANGLE_PRE

  # Mark packets injected into dmesh1 so they get injected into localhost
  #iptables -t mangle -A DMESH_MANGLE_PRE -j MARK -p tcp --dport 5201 --set-mark 1338
  iptables -t mangle -A DMESH_MANGLE_PRE -i ${TUNDEV} -j MARK --set-mark ${TUNFW}7

  # Capture outbound packets
  iptables -t mangle -N DMESH_MANGLE_OUT
  iptables -t mangle -F DMESH_MANGLE_OUT
  iptables -t mangle -A DMESH_MANGLE_OUT -m owner --gid-owner "${GID}" -j RETURN

  # Capture everything else
  #iptables -t mangle -A DMESH_MANGLE_OUT -j MARK --set-mark 1338

  # Explicit
  #iptables -t mangle -A DMESH_MANGLE_OUT -p tcp -d 169.254.169.254 -j DROP

  # Explicit by-port capture, for testing
  iptables -t mangle -A DMESH_MANGLE_OUT -j MARK -p tcp --dport 5201 --set-mark ${TUNFW}8
  iptables -t mangle -A DMESH_MANGLE_OUT -j MARK -p udp --dport 5201 --set-mark ${TUNFW}8
  #iptables -t mangle -A DMESH_MANGLE_OUT -j MARK -p tcp --dport 80 --set-mark 1338

  # Jump to the ISTIO_OUTPUT chain from OUTPUT chain for all tcp traffic.
  iptables -t mangle -A OUTPUT -j DMESH_MANGLE_OUT
}

if [ "$1" = "setup" ] ; then
  setupTUN
  setup
elif [ "$1" = "start" ] ; then
  start
elif [  "$1" = "stop" ] ; then
  stop
elif [ "$1" = "clean" ] ; then
  cleanup
  stop
fi

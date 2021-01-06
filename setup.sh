#!/bin/sh

# Setup the TUN device for capture.
#
# Must run as root.
TUNUSER=costin

setupTUN() {
  ip tuntap add dev dmesh1 mode tun user ${TUNUSER:-istio-proxy} group ${TUNUSER:-istio-proxy}
  ip addr add ${IP4:-10.12.0.1/16} dev dmesh1
  # No IP6 address - confuses linux
  ip link set dmesh1 up


  # Don't remember why this was required
  echo 2 > /proc/sys/net/ipv4/conf/dmesh1/rp_filter
  sysctl -w net.ipv4.ip_forward=1
}

cleanup() {
  # App must be stopped
  ip tuntap del dev dmesh1 mode tun

  ip rule delete  fwmark 1338 priority 10  lookup 1338
  ip route del default dev dmesh1 table 1338

  ip rule del fwmark 1337 lookup 1337
  ip rule del iif dmesh1 lookup 1337
  ip route del local 0.0.0.0/0 dev lo table 1337
}

setup() {
  setupTUN
  ip route add default dev dmesh1 table 1338
  ip rule add  fwmark 1338 priority 10  lookup 1338

  # Route various ranges to dmesh1 - the gate can't initiate its own connections
  # to those ranges. Service VIPs can also use this simpler model.
  ip route add fd::/8 dev dmesh1
  ip route add 10.10.0.0/16 dev dmesh1

  # 1337 means deliver to local host
  ip rule add fwmark 1337 lookup 1337
  ip rule add iif dmesh1 lookup 1337
  ip route add local 0.0.0.0/0 dev lo table 1337
  #ip route add local ::/0 dev lo table 1337
}

stop() {
  iptables -t mangle -D OUTPUT -j DMESH_MANGLE_OUT
  iptables -t mangle -D PREROUTING -j DMESH_MANGLE_PRE

  iptables -t mangle -F DMESH_MANGLE_OUT 2>/dev/null
  iptables -t mangle -X DMESH_MANGLE_OUT 2>/dev/null

  iptables -t mangle -F DMESH_MANGLE_PRE 2>/dev/null
  iptables -t mangle -X DMESH_MANGLE_PRE 2>/dev/null
}

# -j MARK only works in mangle !!!
start() {
    GID=$(id -g ${TUNUSER})

  # Mark packets from dmesh1
  iptables -t mangle -N DMESH_MANGLE_PRE
  iptables -t mangle -A PREROUTING -j DMESH_MANGLE_PRE
  # Incoming packages
  #iptables -t mangle -A DMESH_MANGLE_PRE -j MARK -p tcp --dport 5201 --set-mark 1338
  iptables -t mangle -A DMESH_MANGLE_PRE -i dmesh1 -j MARK --set-mark 1337

  iptables -t mangle -N DMESH_MANGLE_OUT
  iptables -t mangle -F DMESH_MANGLE_OUT
  iptables -t mangle -A DMESH_MANGLE_OUT -m owner --gid-owner "${GID}" -j RETURN
  # Capture everything else
  #iptables -t mangle -A DMESH_MANGLE_OUT -j MARK --set-mark 1338
  #iptables -t mangle -A DMESH_MANGLE_OUT -p tcp -d 169.254.169.254 -j DROP
  #iptables -t mangle -A DMESH_MANGLE_OUT -j MARK -p tcp --dport 80 --set-mark 1338
  iptables -t mangle -A DMESH_MANGLE_OUT -j MARK -p tcp --dport 5201 --set-mark 1338
  # Jump to the ISTIO_OUTPUT chain from OUTPUT chain for all tcp traffic.
  iptables -t mangle -A OUTPUT -j DMESH_MANGLE_OUT
}

if [ "$1" = "setup" ] ; then
  setup
elif [ "$1" = "start" ] ; then
  start
elif [  "$1" = "stop" ] ; then
  stop
elif [ "$1" = "clean" ] ; then
  cleanup
  stop
fi

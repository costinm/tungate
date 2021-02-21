OUT ?= build

all:  lwip ugate
	$(MAKE) TAG=netstack gotag
	$(MAKE) TAG=gvisor gotag


stop:

lwip:
	go build  -o ${OUT}/tun_lwip ./cmd/tungate_lwip
	ls -l ${OUT}/tun_lwip
	strip ${OUT}/tun_lwip
	ls -l ${OUT}/tun_lwip

ugate:
	CGO_ENABLED=0 go build -o ${OUT}/ugate github.com/costinm/ugate/cmd/ugate
	ls -l ${OUT}/ugate
	strip ${OUT}/ugate
	ls -l ${OUT}/ugate


gotag:
	cd ${TAG} && go build  -o ${OUT}/tun_${TAG} ./cmd
	ls -l ${OUT}/tun_${TAG}
	strip ${OUT}/tun_${TAG}
	ls -l ${OUT}/tun_${TAG}

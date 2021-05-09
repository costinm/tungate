OUT ?= $(shell pwd)/build

all:  ugate
	$(MAKE) TAG=lwip gotag
	$(MAKE) TAG=netstack gotag
	$(MAKE) TAG=gvisor gotag
	ls -l ${OUT}


ugate:
	CGO_ENABLED=0 go build -o ${OUT}/ugate github.com/costinm/ugate/cmd/ugatemin
	ls -l ${OUT}/ugate
	strip ${OUT}/ugate
	ls -l ${OUT}/ugate


gotag:
	cd ${TAG} && go build  -o ${OUT}/tun_${TAG} ./cmd
	ls -l ${OUT}/tun_${TAG}
	strip ${OUT}/tun_${TAG}
	ls -l ${OUT}/tun_${TAG}

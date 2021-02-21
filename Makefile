OUT ?= build

all: netstack gvisor lwip ugate


stop:


netstack:
	$(MAKE) TAG=netstack gotag


gvisor:
	$(MAKE) TAG=gvisor gotag

lwip:
	$(MAKE) TAG=lwip gotag

ugate:
	CGO_ENABLED=0 go build -o ${OUT}/ugate github.com/costinm/ugate/cmd/ugate
	ls -l ${OUT}/ugate
	strip ${OUT}/ugate
	ls -l ${OUT}/ugate


gotag:
	go build  -o ${OUT}/tun_${TAG} -tags ${TAG} ./cmd/tungate_${TAG}
	ls -l ${OUT}/tun_${TAG}
	strip ${OUT}/tun_${TAG}
	ls -l ${OUT}/tun_${TAG}

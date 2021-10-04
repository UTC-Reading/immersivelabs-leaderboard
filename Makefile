.DEFAULT_GOAL := build

BINARY=your-package-name

LD_FLAGS += -s -w

compress:
	(which upx > /dev/null && upx -9 -q ${BINARY} > /dev/null) || echo "UPX not installed"

build:
	rm -vf ${BINARY}
	go build -ldflags "${LD_FLAGS}" -v -o ${BINARY}

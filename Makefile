.DEFAULT_GOAL := build

BINARY=immersivelabs-leaderboard

LD_FLAGS += -s -w

compress:
	(which upx > /dev/null && upx -9 -q ${BINARY} > /dev/null) || echo "UPX not installed"

build:
	rm -vf ${BINARY}
	go build -ldflags "${LD_FLAGS}" -v -o ${BINARY}

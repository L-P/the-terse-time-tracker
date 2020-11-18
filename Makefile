VERSION=$(shell git describe --tags)
BUILDFLAGS=-ldflags '-X main.Version=${VERSION}'
EXEC=tt

all: $(EXEC)

$(EXEC):
	go build ${BUILDFLAGS} tt/cmd/tt

lint:
	golangci-lint run

.PHONY: $(EXEC) lint

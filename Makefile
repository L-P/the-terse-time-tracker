VERSION=$(shell git describe --tags)
BUILDFLAGS=-ldflags '-X main.Version=${VERSION}'
EXEC=tt

all: $(EXEC) tags

$(EXEC):
	go build ${BUILDFLAGS} tt/cmd/tt

lint:
	golangci-lint run

tags:
	ctags-universal -R internal cmd

.PHONY: $(EXEC) lint tags

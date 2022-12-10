VERSION=$(shell git describe --tags)
BUILDFLAGS=-ldflags '-X main.Version=${VERSION}'
EXEC=tt

all: $(EXEC)

tt.en.man: man.en.md
	VERSION="$(VERSION)" \
	DATE="$(shell date '+%B %d, %Y')" \
			envsubst '$$VERSION:$$DATE' < "$<" > "man.subst.md"
	pandoc -s -t man "man.subst.md" -o "$@"
	rm "man.subst.md"

$(EXEC):
	go build ${BUILDFLAGS} tt/cmd/tt

debug:
	go build ${BUILDFLAGS} -tags fixture tt/cmd/tt

lint:
	golangci-lint run

tags:
	ctags-universal -R internal cmd

test:
	go test ./...

vendor:
	go get -v
	go mod vendor
	go mod tidy

upgrade:
	go get -u -v
	go mod vendor
	go mod tidy

.PHONY: $(EXEC) lint tags test vendor upgrade

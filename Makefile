VERSION=$(shell git describe --tags)
BUILDFLAGS=-ldflags '-X main.Version=${VERSION}'
EXEC=tt

all: $(EXEC) tt.man

tt.man: man.md
	VERSION="$(VERSION)" \
	DATE="$(shell date '+%B %d, %Y')" \
			envsubst '$$VERSION:$$DATE' < "$<" > "man.subst.md"
	pandoc -s -t man "man.subst.md" -o "$@"
	rm "man.subst.md"

$(EXEC):
	go build ${BUILDFLAGS} tt/cmd/tt

lint:
	golangci-lint run

tags:
	ctags-universal -R internal cmd

test:
	go test ./...

.PHONY: $(EXEC) lint tags test

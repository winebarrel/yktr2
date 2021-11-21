VERSION := v0.1.0
.PHONY: all
all: vet build

.PHONY: build
build:
	go build -ldflags "-X main.version=$(VERSION)" ./cmd/yktr2

.PHONY: vet
vet:
	go vet ./...

.PHONY: clean
clean:
	rm -rf yktr

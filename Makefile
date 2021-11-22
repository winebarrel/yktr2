VERSION := v0.1.0
DOCKER_REPO := public.ecr.aws/winebarrel/yktr2

.PHONY: all
all: vet build

.PHONY: build
build:
	go build -ldflags "-X main.version=$(VERSION)" ./cmd/yktr2

.PHONY: vet
vet:
	go vet ./...

.PHONY: image
image:
	docker build . -t $(DOCKER_REPO)

.PHONY: push
push: image
	docker push $(DOCKER_REPO)

.PHONY: clean
clean:
	rm -rf yktr

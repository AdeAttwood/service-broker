# If the USE_SUDO_FOR_DOCKER env var is set, prefix docker commands with 'sudo'
ifdef USE_SUDO_FOR_DOCKER
	SUDO_CMD = sudo
endif

IMAGE ?= adeattwood/service-broker
TAG ?= $(shell git describe --tags --always)
PULL ?= Always

build: ## Builds the starter pack
	go build -i github.com/AdeAttwood/service-broker/cmd/service-broker

test: ## Runs the tests
	go test -v $(shell go list ./... | grep -v /vendor/ | grep -v /test/)

linux: ## Builds a Linux executable
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
	go build -o service-broker-linux --ldflags="-s" github.com/AdeAttwood/service-broker/cmd/service-broker

image: linux ## Builds a Linux based image
	cp service-broker-linux image/service-broker
	$(SUDO_CMD) docker build image/ -t "$(IMAGE):$(TAG)"

clean: ## Cleans up build artifacts
	rm -f service-broker
	rm -f service-broker-linux
	rm -f image/service-broker

push: image ## Pushes the image to dockerhub, REQUIRES SPECIAL PERMISSION
	$(SUDO_CMD) docker push "$(IMAGE):$(TAG)"

help: ## Shows the help
	@echo 'Usage: make <OPTIONS> ... <TARGETS>'
	@echo ''
	@echo 'Available targets are:'
	@echo ''
	@grep -E '^[ a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
        awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ''

.PHONY: build test linux image clean push help

SHELL=/bin/bash -o pipefail

PROJECT_NAME=kube-job-cleaner
CODE_REPO=veezhang/$(PROJECT_NAME)
DOCKER_REPO=veezhang/$(PROJECT_NAME)
VERSION=$(shell git describe --always --tags --dirty | sed "s/\(.*\)-g`git rev-parse --short HEAD`/\1/")
GIT_SHA=$(shell git rev-parse --short HEAD)

.PHONY: all build check clean test

all: dep check build

dep:
	go mod tidy
	go mod vendor

build: test build-go build-image

build-go:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
	-ldflags "-X github.com/$(CODE_REPO)/pkg/version.Version=$(VERSION) -X github.com/$(CODE_REPO)/pkg/version.GitSHA=$(GIT_SHA)" \
	-o bin/$(PROJECT_NAME)-linux-amd64 cmd/main.go

build-image:
	docker build --build-arg VERSION=$(VERSION) --build-arg GIT_SHA=$(GIT_SHA) -t $(DOCKER_REPO):$(VERSION) .
	docker tag $(DOCKER_REPO):$(VERSION) $(DOCKER_REPO):latest

test:
	# go test $$(go list ./... | grep -v /vendor/) -race -coverprofile=coverage.txt -covermode=atomic

login:
	@docker login -u "$(DOCKER_USER)" -p "$(DOCKER_PASS)"

push: build-image login
	docker push $(DOCKER_REPO):$(VERSION)
	docker push $(DOCKER_REPO):latest

clean:
	rm -f bin/$(PROJECT_NAME)

check: check-format

check-format:
	./scripts/check_format.sh
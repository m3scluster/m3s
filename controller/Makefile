#Dockerfile vars

.PHONY: help build all

TAG=$(shell git describe)
BUILDDATE=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)
BRANCH=$(shell git symbolic-ref --short HEAD)
VERSION_URL=https://raw.githubusercontent.com/AVENTER-UG/mesos-m3s/${BRANCH}/bootstrap/.version.json

.DEFAULT_GOAL := all

build-arm:
	@CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -a -installsuffix cgo -ldflags "-X main.BuildVersion=${BUILDDATE}/ARM -X main.GitVersion=${TAG} -X main.VersionURL=${VERSION_URL} -extldflags \"-static\"" -o controller.arm64 .

build: 
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags "-X main.BuildVersion=${BUILDDATE} -X main.GitVersion=${TAG} -X main.VersionURL=${VERSION_URL} -extldflags \"-static\"" -o controller.amd64 .

update-gomod:
	go get -u
	go mod tidy

seccheck:
	grype --add-cpes-if-none .

sboom:
	syft dir:. > sbom.txt
	syft dir:. -o json > sbom.json

go-fmt:
	@gofmt -w .

all: go-fmt seccheck sboom build

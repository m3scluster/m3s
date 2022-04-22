#Dockerfile vars

.PHONY: help build all

TAG=$(shell git describe)
BUILDDATE=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)
BRANCH=$(shell git symbolic-ref --short HEAD)
VERSION_URL=https://raw.githubusercontent.com/AVENTER-UG/mesos-m3s/${BRANCH}/bootstrap/.version.json


help:
	    @echo "Makefile arguments:"
	    @echo ""
	    @echo "Makefile commands:"
	    @echo "build"
	    @echo "all"

.DEFAULT_GOAL := all

build:
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags "-X main.BuildVersion=${BUILDDATE} -X main.GitVersion=${TAG} -X main.VersionURL=${VERSION_URL} -extldflags \"-static\"" -o server server.go

version:
	@echo ">>>> Generate version file"
	@echo "{\"bootstrapVersion\": {	\"gitVersion\": \"${TAG}\",	\"buildDate\": \"${BUILDDATE}\"}}" > .version.json
	@cat .version.json
	@echo "Saved under .version.json"

gen-files:
	@echo ">>>> Generate Bootstrap and Update Script"
	@sed -si 's/^export BRANCH=.*/export BRANCH=${BRANCH}/' bootstrap.sh
	@sed -si 's/^export BRANCH=.*/export BRANCH=${BRANCH}/' update.sh

all: version gen-files build

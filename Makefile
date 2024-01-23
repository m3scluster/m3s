#Dockerfile vars

#vars
IMAGENAME=mesos-m3s
TAG=`git describe`
BUILDDATE=`date -u +%Y-%m-%dT%H:%M:%SZ`
IMAGEFULLNAME=avhost/${IMAGENAME}
BRANCH=`git symbolic-ref --short HEAD`
VERSION_URL=https://raw.githubusercontent.com/AVENTER-UG/mesos-m3s/${BRANCH}/.version.json
LASTCOMMIT=$(shell git log -1 --pretty=short | tail -n 1 | tr -d " " | tr -d "UPDATE:")

.PHONY: help build bootstrap all docs publish push version

help:
	    @echo "Makefile arguments:"
	    @echo ""
	    @echo "Makefile commands:"
		@echo "push"
	    @echo "build"
		@echo "build-bin"
	    @echo "all"
		@echo "docs"
		@echo "publish"
		@echo "version"
		@echo ${TAG}

.DEFAULT_GOAL := all

ifeq (${BRANCH}, master) 
	BRANCH=latest
endif

ifneq ($(shell echo $(LASTCOMMIT) | grep -E '^v([0-9]+\.){0,2}(\*|[0-9]+)'),)
	BRANCH=${LASTCOMMIT}
else
	BRANCH=latest
endif

build:
	@echo ">>>> Build docker image branch:" ${BRANCH}
	@docker buildx build --build-arg TAG=${TAG} --build-arg BUILDDATE=${BUILDDATE} --build-arg VERSION_URL=${VERSION_URL} -t ${IMAGEFULLNAME}:${BRANCH} .

build-bin:
	@echo ">>>> Build binary"
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags "-X main.BuildVersion=${BUILDDATE} -X main.GitVersion=${TAG} -X main.VersionURL=${VERSION_URL} -extldflags \"-static\"" .

cont:
	@echo ">>>> Build controller"
	$(MAKE) -C controller/
	@cp controller/controller.amd64 bootstrap/

push:
	@echo ">>>> Publish docker image" ${BRANCH}
	@docker buildx create --use --name buildkit
	@docker buildx build --platform linux/arm64,linux/amd64 --push --build-arg TAG=${TAG} --build-arg BUILDDATE=${BUILDDATE} --build-arg VERSION_URL=${VERSION_URL} -t ${IMAGEFULLNAME}:${BRANCH} .
	@docker buildx rm buildkit

docs:
	@echo ">>>> Build docs"
	$(MAKE) -C $@

plugin: 
	@echo ">>> Build plugins"
	cd plugins; $(MAKE)	

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

version:
	@echo ">>>> Generate version file"
	@echo "{\"m3sVersion\": {	\"gitVersion\": \"${TAG}\",	\"buildDate\": \"${BUILDDATE}\"}}" > .version.json
	@cat .version.json
	@echo "Saved under .version.json"

check: go-fmt sboom seccheck
all: check version cont build

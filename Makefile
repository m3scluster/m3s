#Dockerfile vars

#vars
IMAGENAME=mesos-m3s
REPO=localhost:5000
TAG=`git describe`
BUILDDATE=`date -u +%Y-%m-%dT%H:%M:%SZ`
IMAGEFULLNAME=${REPO}/${IMAGENAME}
IMAGEFULLNAMEPUB=avhost/${IMAGENAME}
BRANCH=`git symbolic-ref --short HEAD`
VERSION_URL=https://raw.githubusercontent.com/AVENTER-UG/mesos-m3s/${BRANCH}/.version.json

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

build:
	@echo ">>>> Build docker image"
	@docker buildx build --build-arg TAG=${TAG} --build-arg BUILDDATE=${BUILDDATE} --build-arg VERSION_URL=${VERSION_URL} -t ${IMAGEFULLNAME}:${BRANCH} .

push:
	@echo ">>>> Push into private repo"
	@docker push localhost:5000/mesos-m3s:dev

build-bin:
	@echo ">>>> Build binary"
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags "-X main.BuildVersion=${BUILDDATE} -X main.GitVersion=${TAG} -X main.VersionURL=${VERSION_URL} -extldflags \"-static\"" .

bootstrap:
	@echo ">>>> Build bootstrap"
	$(MAKE) -C $@

publish:
	@echo ">>>> Publish docker image"
	@docker tag ${IMAGEFULLNAME}:${BRANCH} ${IMAGEFULLNAMEPUB}:${BRANCH}
	@docker push ${IMAGEFULLNAMEPUB}:${BRANCH}

docs:
	@echo ">>>> Build docs"
	$(MAKE) -C $@

update-gomod:
	go get -u
	go mod tidy

seccheck:
	gosec --exclude G104 --exclude-dir ./vendor ./... 

sboom:
	syft dir:. > sbom.txt
	syft dir:. -o json > sbom.json

version:
	@echo ">>>> Generate version file"
	@echo "{\"m3sVersion\": {	\"gitVersion\": \"${TAG}\",	\"buildDate\": \"${BUILDDATE}\"}}" > .version.json
	@cat .version.json
	@echo "Saved under .version.json"

all: bootstrap build version

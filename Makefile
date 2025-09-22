#Dockerfile vars

#vars
IMAGENAME=mesos-m3s
TAG=v0.5.3
BUILDDATE=$(shell date -u +%Y%m%d)
IMAGEFULLNAME=avhost/${IMAGENAME}
BRANCH=$(shell git symbolic-ref --short HEAD | xargs basename)
BRANCHSHORT=$(shell echo ${BRANCH} | awk -F. '{ print $$1"."$$2 }')
VERSION_URL=https://raw.githubusercontent.com/AVENTER-UG/mesos-m3s/${BRANCH}/.version.json
LASTCOMMIT=$(shell git log -1 --pretty=short | tail -n 1 | tr -d " " | tr -d "UPDATE:")


.PHONY: help build bootstrap all docs publish push version

.DEFAULT_GOAL := all

build:
	@echo ">>>> Build docker image branch:" ${BRANCH}
	@docker buildx build --build-arg TAG=${TAG} --build-arg BUILDDATE=${BUILDDATE} --build-arg VERSION_URL=${VERSION_URL} -t ${IMAGEFULLNAME}:latest .

build-bin:
	@echo ">>>> Build binary"
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags "-X main.BuildVersion=${BUILDDATE} -X main.GitVersion=${TAG} -X main.VersionURL=${VERSION_URL} -extldflags \"-static\"" .

controller-bin:
	@echo ">>>> Build controller"
	$(MAKE) -C controller_bin/
	@cp controller_bin/controller.amd64 bootstrap/

push:
	@echo ">>>> Publish docker image" ${BRANCH} ${BRANCHSHORT}
	-docker buildx create --use --name buildkit
	@docker buildx build --sbom=true --provenance=true --platform linux/amd64 --push --build-arg TAG=${TAG} --build-arg BUILDDATE=${BUILDDATE} --build-arg VERSION_URL=${VERSION_URL} -t ${IMAGEFULLNAME}:${BRANCH} .
	@docker buildx build --sbom=true --provenance=true --platform linux/amd64 --push --build-arg TAG=${TAG} --build-arg BUILDDATE=${BUILDDATE} --build-arg VERSION_URL=${VERSION_URL} -t ${IMAGEFULLNAME}:${BRANCHSHORT} .
	@docker buildx build --sbom=true --provenance=true --platform linux/amd64 --push --build-arg TAG=${TAG} --build-arg BUILDDATE=${BUILDDATE} --build-arg VERSION_URL=${VERSION_URL} -t ${IMAGEFULLNAME}:latest .
	-docker buildx rm buildkit

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

imagecheck:
	grype --add-cpes-if-none ${IMAGEFULLNAME}:latest > cve-report.md

go-fmt:
	@gofmt -w .

version:
	@echo ">>>> Generate version file"
	@echo "{\"m3sVersion\": {	\"gitVersion\": \"${TAG}\",	\"buildDate\": \"${BUILDDATE}\"}}" > .version.json
	@cat .version.json
	@echo "Saved under .version.json"

check: go-fmt sboom seccheck imagecheck
all: check version controller-bin build


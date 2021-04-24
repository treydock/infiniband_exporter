# Needs to be defined before including Makefile.common to auto-generate targets
DOCKER_ARCHS ?= amd64 ppc64le s390x
DOCKER_REPO	 ?= treydock

include Makefile.common

DOCKER_IMAGE_NAME ?= infiniband_exporter

coverage:
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...

# Makefile

# set the default shell to bash
SHELL = /bin/bash

# build version
# by default, all builds are development versions
VERSION ?= dev

# go build env vars
# |
# |- by default, build for the host operating system
GOOS ?= $(shell go env GOOS)
# |
# |- by default, build for the host architecture
GOARCH ?= $(shell go env GOARCH)

# name of the binary that is being built
BIN = cmd/stevebot/stevebot-$(VERSION)-$(GOOS)-$(GOARCH)

# docker image name
DOCKER_IMAGE_NAME ?= cezarmathe/stevebot

all: build run

# build stevebot
build:
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 go build \
		-o $(BIN) \
		-ldflags "-X 'main.Version=$(VERSION)'" \
		github.com/cezarmathe/stevebot/cmd/stevebot
.PHONY: build

# package stevebot - this builds the docker image
package:
	docker build \
		--build-arg version=$(VERSION) \
		--file build/package/Dockerfile \
		--tag $(DOCKER_IMAGE_NAME):$(VERSION) .
.PHONY: package

# run stevebot
run:
	STEVEBOT_RCON_HOST=$(shell jq .rcon_host < dev.json) \
		STEVEBOT_RCON_PORT=$(shell jq .rcon_port < dev.json) \
		STEVEBOT_RCON_PASSWORD=$(shell jq .rcon_password < dev.json) \
		STEVEBOT_DISCORD_TOKEN=$(shell jq .discord_token < dev.json) \
		STEVEBOT_COMMAND_PREFIX=$(shell jq .command_prefix < dev.json) \
		STEVEBOT_ALLOWED_COMMANDS=$(shell jq .allowed_commands < dev.json) \
		STEVEBOT_FORBIDDEN_COMMANDS=$(shell jq .forbidden_commands < dev.json) \
		./$(BIN)
.PHONY: run

# run stevebot (packaged version - Docker)
run-package:
	docker run \
		--env STEVEBOT_RCON_HOST=$(shell jq .rcon_host < dev.json) \
		--env STEVEBOT_RCON_PORT=$(shell jq .rcon_port < dev.json) \
		--env STEVEBOT_RCON_PASSWORD=$(shell jq .rcon_password < dev.json) \
		--env STEVEBOT_DISCORD_TOKEN=$(shell jq .discord_token < dev.json) \
		--env STEVEBOT_COMMAND_PREFIX=$(shell jq .command_prefix < dev.json) \
		--env STEVEBOT_ALLOWED_COMMANDS=$(shell jq .allowed_commands < dev.json) \
		--env STEVEBOT_FORBIDDEN_COMMANDS=$(shell jq .forbidden_commands < dev.json) \
		--rm \
		-it \
		$(DOCKER_IMAGE_NAME):$(VERSION)
.PHONY: run

# clean artifacts (just binaries, docker image is not removed)
clean:
	rm -r cmd/stevebot/stevebot-*-*-*
.PHONY: clean

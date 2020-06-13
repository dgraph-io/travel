SHELL := /bin/bash

# Building containers

all: travel-api travel-ui

travel-api:
	docker build \
		-f dockerfile.travel-api \
		-t travel-api-amd64:1.0 \
		--build-arg PACKAGE_NAME=travel-api \
		--build-arg VCS_REF=`git rev-parse HEAD` \
		--build-arg BUILD_DATE=`date -u +”%Y-%m-%dT%H:%M:%SZ”` \
		.

travel-ui:
	docker build \
		-f dockerfile.travel-ui \
		-t travel-ui-amd64:1.0 \
		--build-arg PACKAGE_NAME=travel-ui \
		--build-arg VCS_REF=`git rev-parse HEAD` \
		--build-arg BUILD_DATE=`date -u +”%Y-%m-%dT%H:%M:%SZ”` \
		.

# Running from within docker compose

run: up seed browse

up:
	docker-compose -f compose.yaml up --detach --remove-orphans

down:
	docker-compose -f compose.yaml down --remove-orphans

browse:
	python -m webbrowser "http://localhost"

logs:
	docker-compose -f compose.yaml logs -f

# Running from within the local with Slash

slash-run: slash-up seed slash-browse

slash-up:
	docker-compose -f compose-slash.yaml up --detach --remove-orphans

slash-down:
	docker-compose -f compose-slash.yaml down --remove-orphans

slash-browse:
	python -m webbrowser "http://localhost"

slash-logs:
	docker-compose -f compose-slash.yaml logs -f

# Running Local

local-run: local-up seed browse

local-up:
	go run cmd/travel-api/main.go &> api.log &
	cd cmd/travel-ui; \
	go run main.go &> ../../ui.log &

API := $(shell lsof -i tcp:4000 | cut -c9-13 | grep "[0-9]")
UI := $(shell lsof -i tcp:4080 | cut -c9-13 | grep "[0-9]")

ps:
	lsof -i tcp:4000; \
	lsof -i tcp:4080

local-down:
	kill -15 $(API); \
	kill -15 $(UI); \
	rm *.log

api-logs:
	tail -F api.log

ui-logs:
	tail -F ui.log

# Administration

schema:
	go run cmd/travel-admin/main.go schema

seed: schema
	go run cmd/travel-admin/main.go seed

# Running tests within the local computer

test:
	go test ./... -count=1
	staticcheck ./...

# Modules support

deps-reset:
	git checkout -- go.mod
	go mod tidy
	go mod vendor

tidy:
	go mod tidy
	go mod vendor

deps-upgrade:
	go get -u -t -d -v ./...
	go mod vendor

deps-cleancache:
	go clean -modcache

# Docker support

FILES := $(shell docker ps -aq)

down-local:
	docker stop $(FILES)
	docker rm $(FILES)

clean:
	docker system prune -f

logs-local:
	docker logs -f $(FILES)

# Git support

install-hooks:
	cp -r .githooks/pre-commit .git/hooks/pre-commit

remove-hooks:
	rm .git/hooks/pre-commit
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

run: up seed browser

up:
	docker-compose up --detach --remove-orphans

down:
	docker-compose down --remove-orphans

browser:
	python -m webbrowser "http://localhost"

logs:
	docker-compose logs -f

# Running from within the local computer

run-local: up-local seed browser-local ui-local

up-local:
	docker run -it -d -p 8080:8080 dgraph/standalone:v20.03.1

ui-local:
	cd cmd/travel-ui; \
	go run main.go --web-ui-host=0.0.0.0:81

FILES := $(shell docker ps -aq)

down-local:
	docker stop $(FILES)
	docker rm $(FILES)

browser-local:
	python -m webbrowser "http://localhost:81"

logs-local:
	docker logs -f

# Seeding the database

seed:
	go run cmd/travel-data/main.go

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

clean:
	docker system prune -f

# Git support

install-hooks:
	cp -r .githooks/ .git/hooks/
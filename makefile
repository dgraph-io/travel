SHELL := /bin/bash

all: travel-api

travel-api:
	docker build \
		-f dockerfile.travel-api \
		-t travel-api-amd64:1.0 \
		--build-arg PACKAGE_NAME=travel-api \
		--build-arg VCS_REF=`git rev-parse HEAD` \
		--build-arg BUILD_DATE=`date -u +”%Y-%m-%dT%H:%M:%SZ”` \
		.

up:
	docker-compose up

down:
	docker-compose down

seed:
	go run cmd/travel-data/main.go

ui:
	go run cmd/travel-ui/main.go

test:
	go test ./... -count=1
	staticcheck ./...

clean:
	docker system prune -f

stop-all:
	docker stop $(docker ps -aq)

remove-all:
	docker rm $(docker ps -aq)

deps-reset:
	git checkout -- go.mod
	go mod tidy
	go mod vendor

deps-upgrade:
	go get -u -t -d -v ./...
	go mod vendor

deps-cleancache:
	go clean -modcache
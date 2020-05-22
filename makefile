SHELL := /bin/bash

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

run: up seed browser

up:
	docker-compose up --detach

down:
	docker-compose down

logs:
	docker-compose logs -f

seed:
	go run cmd/travel-data/main.go

ui:
	cd cmd/travel-ui; \
	go run main.go --web-ui-host=0.0.0.0:81

database:
	docker run -it -d -p 8080:8080 dgraph/standalone:v20.03.1

test:
	go test ./... -count=1
	staticcheck ./...

browser:
	python -m webbrowser "http://localhost"

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

tidy:
	go mod tidy
	go mod vendor

deps-upgrade:
	go get -u -t -d -v ./...
	go mod vendor

deps-cleancache:
	go clean -modcache
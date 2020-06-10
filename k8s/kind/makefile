SHELL := /bin/bash

all:
	(cd ../.. && make all)

run: up load services seed browser

up:
	kind create cluster --name dgraph-travel-cluster --config kind-config.yaml

down:
	kind delete cluster --name dgraph-travel-cluster

seed:
	cd ../..; \
	go run cmd/travel-data/main.go --custom-functions-send-email-url=http://localhost:3000/v1/email

browser:
	(cd ../.. && make browser-compose)

load:
	kind load docker-image travel-ui-amd64:1.0 --name dgraph-travel-cluster
	kind load docker-image travel-api-amd64:1.0 --name dgraph-travel-cluster
	# kind load docker-image dgraph/dgraph:master --name dgraph-travel-cluster

services:
	kubectl create -f deploy-dgraph.yaml
	kubectl create -f deploy-travel-ui.yaml
	kubectl create -f deploy-travel-api.yaml

get-pods:
	kubectl get pods

status:
	kubectl get nodes
	kubectl get pods
	kubectl get services travel-api

delete:
	kubectl delete -f deploy-dgraph.yaml
	kubectl delete pvc -l app=dgraph
	kubectl delete -f deploy-travel-ui.yaml
	kubectl delete -f deploy-travel-api.yaml
	@echo ======================================================================

SHELL := /bin/bash

export PROJECT = dgraph-travel-project
CLUSTER = dgraph-travel-cluster
ZONE = us-central1-b

config:
	@echo Setting environment for $(PROJECT)
	gcloud config set project $(PROJECT)
	gcloud config set compute/zone $(ZONE)
	gcloud auth configure-docker
	@echo ======================================================================

project:
	gcloud projects create $(PROJECT)
	gcloud beta billing projects link $(PROJECT) --billing-account=$(ACCOUNT_ID)
	gcloud services enable container.googleapis.com
	@echo ======================================================================

cluster:
	gcloud container clusters create $(CLUSTER) --enable-ip-alias --num-nodes=2 --machine-type=n1-standard-2
	gcloud compute instances list
	@echo ======================================================================

upload:
	docker push gcr.io/$(PROJECT)/travel-api-amd64:1.0
	docker push gcr.io/$(PROJECT)/travel-ui-amd64:1.0
	@echo ======================================================================

services:
	# These scripts needs to be edited for the PROJECT and PRIVATE_DB_IP markers before running.
	kubectl create -f deploy-dgraph.yaml
	kubectl create -f deploy-travel-api.yaml
	kubectl create -f deploy-travel-ui.yaml
	kubectl create -f expose-dgraph.yaml
	kubectl create -f expose-travel-api.yaml
	kubectl create -f expose-travel-ui.yaml
	@echo ======================================================================

seed: DGRAPH_PUBLIC = http://$(shell kubectl get svc dgraph-public -o jsonpath='{.status.loadBalancer.ingress[0].ip}:{.spec.ports[0].port}')
seed:
	(export DATA_DGRAPH_URL=$(DGRAPH_PUBLIC); cd ../.. && make seed)

browser: TRAVEL_UI = http://$(shell kubectl get svc travel-ui -o jsonpath='{.status.loadBalancer.ingress[0].ip}:{.spec.ports[0].port}')
browser:
	python -m webbrowser "$(TRAVEL_UI)"

status:
	gcloud container clusters list
	kubectl get nodes
	kubectl get pods
	kubectl get services travel-api
	@echo ======================================================================

shell:
	# kubectl get pods
	kubectl exec -it <POD NAME> --container travel-api  -- /bin/sh
	# ./admin --db-disable-tls=1 migrate
	# ./admin --db-disable-tls=1 seed
	@echo ======================================================================

delete:
	kubectl delete -f deploy-dgraph.yaml
	kubectl delete pvc -l 'app in (dgraph-zero, dgraph-alpha)'
	kubectl delete -f deploy-travel-api.yaml
	kubectl delete -f deploy-travel-ui.yaml
	kubectl delete -f expose-dgraph.yaml
	kubectl delete -f expose-travel-api.yaml
	kubectl delete -f expose-travel-ui.yaml
	gcloud container clusters delete $(CLUSTER)
	# gcloud projects delete $(PROJECT)
	gcloud container images delete gcr.io/$(PROJECT)/travel-api-amd64:1.0 --force-delete-tags
	gcloud container images delete gcr.io/$(PROJECT)/travel-ui-amd64:1.0 --force-delete-tags
	docker image remove gcr.io/$(PROJECT)/travel-api-amd64:1.0
	docker image remove gcr.io/$(PROJECT)/travel-ui-amd64:1.0
	@echo ======================================================================

#===============================================================================
# GKE Installation
#
# Install the Google Cloud SDK. This contains the gcloud client needed to perform
# some operatings
# https://cloud.google.com/sdk/
#
# Installing the K8s kubectl client. 
# https://kubernetes.io/docs/tasks/tools/install-kubectl/

# SPDX-License-Identifier: MIT

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ debug prometheus

.PHONY: prom-port-forward
prom-port-forward: ## Port forward Prometheus.
	kubectl --kubeconfig /tmp/kcd.kubeconfig port-forward svc/prometheus-server 9090:80


.PHONY: k-debug-prometheus
k-debug-prometheus: get-kubeconfig ## Debug Prometheus.
	@echo "Prometheus is running at http://localhost:9090"

.PHONE: cat-prom-dockerfile
cat-prom-dockerfile: ## Cat Dockerfile for Prometheus.
	echo "FROM ubuntu:20.04"
	echo "RUN curl -LO prometheus.tar.gz"
	echo "RUN tar -xvf prometheus.tar.gz"

##@ Prepare Stage

.PHONY: kind-cluster
kind-cluster: ## Create a kind cluster.
	kind create cluster --name kcd

.PHONY: get-kubeconfig
get-kubeconfig: ## Export kubeconfig.
	kind export kubeconfig --name kcd --kubeconfig /tmp/kcd.kubeconfig

.PHONY: apply-prometheus
apply-prometheus: get-kubeconfig ## Apply Prometheus.
	helm upgrade prometheus prometheus-community/prometheus --install --kubeconfig /tmp/kcd.kubeconfig --values prometheus-values.yaml
	

.PHONY: add-community-chart
add-community-chart: ## Add community chart repository.
	helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
	helm repo update


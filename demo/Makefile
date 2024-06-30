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

.PHONY: exec-node
exec-node: ## start a shell on node level
	kubectl debug node/kcd-control-plane -it --image=nicolaka/netshoot:v0.13

.PHONY: exec-pod
exec-pod: ## Start a shell at the pod
	$(eval POD=$(shell kubectl get pod -l run=webapp -o=jsonpath='{.items[0].metadata.name}'))
	kubectl exec -it $(POD) -- /bin/sh


.PHONY: prom-port-forward
prom-port-forward: ## Port forward Prometheus.
	kubectl port-forward svc/prometheus-server 9090:80

.PHONY: app-port-forward
app-port-forward: ## Port forward webapp.
	$(eval POD=$(shell kubectl get pod -l run=webapp -o=jsonpath='{.items[0].metadata.name}'))
	kubectl port-forward $(POD) 8080

.PHONY: app-metrics-port-forward
app-metrics-port-forward: ## Port forward to webapp on metrics port
	$(eval POD=$(shell kubectl get pod -l run=webapp -o=jsonpath='{.items[0].metadata.name}'))
	kubectl port-forward $(POD) 9090

.PHONY: debug-pod
debug-pod: ## kubectl debug
	$(eval POD=$(shell kubectl get pod -l run=webapp -o=jsonpath='{.items[0].metadata.name}'))
	kubectl debug --image=nicolaka/netshoot:v0.13 $(POD) -it

.PHONY: debug-pod-with-profile
debug-pod-with-profile: ## kubectl custom profile
	$(eval POD=$(shell kubectl get pod -l run=webapp -o=jsonpath='{.items[0].metadata.name}'))
	KUBECTL_DEBUG_CUSTOM_PROFILE=true kubectl debug --image=nicolaka/netshoot:v0.13 --custom debug-profiles/app-profile-config.json $(POD) -it

.PHONY: show-dockerfile-diff
show-dockerfile-diff: ## show container build diff
	diff --color -Pru app/Dockerfile app/all.Dockerfile || /bin/true

.PHONY: show-debug-diff
show-debug-diff: ## show diff of k debug command
	diff --color -Pru <(echo "kubectl debug --image=nicolaka/netshoot:v0.13 PODNAME -it") <(echo -e "KUBECTL_DEBUG_CUSTOM_PROFILE=true \nkubectl debug --image=nicolaka/netshoot:v0.13 --custom debug-profiles/app-profile-config.json PODNAME -it") || /bin/true
	echo -e "\n\n"
	jq '.' debug-profiles/app-profile-config.json

.PHONY: show-dpm-config
show-dpm-config: ## print dpm config
	yq -P debug-profiles/debug-profiles.yaml

.PHONY: run-dpm
run-dpm: ## run k dpm
	kubectl dpm run --profile webapp --config debug-profiles/debug-profiles.yaml

.PHONY: qr-code
qr-code: ## print dpm qr code
	qrencode -o - -t UTF8 http://github.com/bavarianbidi/kubectl-dpm

##@ Prepare Stage

.PHONY: kind-cluster
kind-cluster: ## Create a kind cluster.
	kind create cluster --name kcd

.PHONY: get-kubeconfig
get-kubeconfig: ## Export kubeconfig.
	kind export kubeconfig --name kcd --kubeconfig /tmp/kcd.kubeconfig

.PHONY: apply-prometheus
apply-prometheus: get-kubeconfig add-community-chart ## Apply Prometheus.
	helm upgrade prometheus prometheus-community/prometheus --install --kubeconfig /tmp/kcd.kubeconfig

.PHONY: add-community-chart
add-community-chart: ## Add community chart repository.
	helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
	helm repo update

##@ Webapp

.PHONY: apply-webapp
apply-webapp: ## Apply webapp.
	kubectl apply -f app/manifest/

.PHONY: build-webapp
build-webapp: ## Build webapp.
	docker build -t quay.io/bavarianbidi/demo-app:latest app/
	docker push quay.io/bavarianbidi/demo-app:latest

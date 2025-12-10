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
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-30s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ debug prometheus

.PHONY: debug-pod-with-profile
debug-pod-with-profile: ## kubectl custom profile
	$(eval POD=$(shell kubectl get pod -l run=webapp -o=jsonpath='{.items[0].metadata.name}'))
	KUBECTL_DEBUG_CUSTOM_PROFILE=true kubectl debug --image=nicolaka/netshoot:v0.13 --custom debug-profiles/app-profile-config.json $(POD) -it

#.PHONY: show-dockerfile-diff
#show-dockerfile-diff: ## show container build diff
#	diff --color -Pru app/Dockerfile app/all.Dockerfile || /bin/true

#.PHONY: show-debug-diff
#show-debug-diff: ## show diff of k debug command
#	diff --color -Pru <(echo "kubectl debug --image=nicolaka/netshoot:v0.13 PODNAME -it") <(echo -e "KUBECTL_DEBUG_CUSTOM_PROFILE=true \nkubectl debug --image=nicolaka/netshoot:v0.13 --custom debug-profiles/app-profile-config.json PODNAME -it") || /bin/true
#	echo -e "\n\n"
#	jq '.' debug-profiles/app-profile-config.json

.PHONY: run-dpm
run-dpm: ## run k dpm
	kubectl dpm run --profile webapp --config debug-profiles/debug-profiles.yaml



##@ Prepare Stage

.PHONY: kind-cluster
kind-cluster: ## Create a kind cluster.
	kind-with-registry.sh

.PHONY: get-kubeconfig
get-kubeconfig: ## Export kubeconfig.
	kind export kubeconfig --name kcd --kubeconfig /tmp/kcd.kubeconfig

##@ Prepare

.PHONY: apply-webapp
apply-webapp: ## Apply webapp.
	kubectl apply -f app/manifest/
	kubectl rollout restart deploy/webapp

.PHONY: build-webapp
build-webapp: ## Build webapp.
	docker build -t localhost:5000/demo-app:latest app/
	docker push localhost:5000/demo-app:latest

.PHONY: port-forward-webapp
port-forward-webapp: ## Port forward webapp.
	kubectl port-forward svc/webapp 8080:8080

.PHONY: status-on-pod
status-on-pod: ## print status on pod
	$(eval POD=$(shell kubectl get pod -l run=webapp -o=jsonpath='{.items[0].metadata.name}'))
	kubectl get pod $(POD) -o yaml | yq '.status.ephemeralContainerStatuses[1].state'
	kubectl get pod $(POD) -o yaml | yq '.spec.ephemeralContainers[1]'

##@ Live Demo

.PHONY: 01-access-webapp
01-access-webapp: ## Access webapp.
	curl -i http://localhost:8080

.PHONY: 02-get-pod-logs
02-get-pod-logs: ## Get the pod logs
	$(eval POD=$(shell kubectl get pod -l run=webapp -o=jsonpath='{.items[0].metadata.name}'))
	kubectl logs $(POD)

.PHONY: 03-exec-pod
03-exec-pod: ## Start a shell at the pod
	$(eval POD=$(shell kubectl get pod -l run=webapp -o=jsonpath='{.items[0].metadata.name}'))
	kubectl exec -it $(POD) -- /bin/sh

.PHONY: 04-rebuild-webapp-with-alpine
04-rebuild-webapp-with-alpine: ## Build webapp.
#docker build -t localhost:5000/demo-app:alpine -f app/all.Dockerfile app/
	diff --color -Pru app/Dockerfile app/all.Dockerfile || /bin/true
	@echo ""
	@echo ""
	@GRYPE_DB_AUTO_UPDATE=false grype localhost:5000/demo-app:alpine -q

.PHONY: 05-debug-pod
05-debug-pod: ## k debug
	$(eval POD=$(shell kubectl get pod -l run=webapp -o=jsonpath='{.items[0].metadata.name}'))
	kubectl debug --image=nicolaka/netshoot:v0.13 $(POD) -it

.PHONY: 06-print-debug-profile
06-print-debug-profile: ## print debug profile
	@cat debug-profiles/app-profile-config.json | jq '.'

.PHONY: 07-show-kubectl-debug-command
07-show-kubectl-debug-command: ## show k debug command with custom profile
	$(eval POD=$(shell kubectl get pod -l run=webapp -o=jsonpath='{.items[0].metadata.name}'))
	@echo "kubectl debug \
	--namespace default \
	--custom app-profile-config.json \
	--image nicolaka/netshoot:v0.13 \
	$(POD) -it"

.PHONY: 08-debug-pod-with-profile
08-debug-pod-with-profile: ## k debug with custom profile
#kubectl dpm run --profile webapp --config debug-profiles/debug-profiles.yaml
	kubectl dpm run --profile webapp

.PHONY: 09-show-dpm-config
09-show-dpm-config: ## print dpm config
	@yq -P ~/.kube-dpm/debug-profiles.yaml

.PHONY: 10-qr-code
10-qr-code: ## print dpm qr code
	@qrencode -o - -t UTF8 http://github.com/bavarianbidi/kubectl-dpm



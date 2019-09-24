
# Image URL to use all building/pushing image targets
IMG ?= quay.io/awesomenix/drainsafe-manager:latest
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"
GOBIN ?= $(PWD)/bin
PLATFORM := $(shell go env GOOS;)
ARCH := $(shell go env GOARCH;)
HAS_KUBEBUILDER := $(shell command -v $(GOBIN)/kubebuilder;)
HAS_CONTROLLER_GEN := $(shell command -v $(GOBIN)/controller-gen;)
KUBEBUILDER_VERSION := 2.0.1
CONTROLLER_GEN := $(GOBIN)/controller-gen
BUILD_DIR := $(CURDIR)

all: manager

# Run tests
test: generate fmt vet manifests
	KUBEBUILDER_ASSETS=$(GOBIN) go test ./... -coverprofile cover.out

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet
	go run main.go

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	kustomize build config/default | kubectl apply -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:rbac:dir=./config/rbac

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths=./api/...

# Build the docker image
docker-build: test
	docker build . -t ${IMG}

# Push the docker image
docker-push:
	docker push ${IMG}
	@echo "updating kustomize image patch file for manager resource"
	sed -i'' -e 's@image: .*@image: '"${IMG}"'@' ./config/default/manager_image_patch.yaml
	sed -i'' -e 's@image: .*@image: '"${IMG}"'@' ./config/default/scheduledevent_manager_image_patch.yaml
	rm -f config/default/manager_image_patch.yaml-e
	rm -f config/default/scheduledevent_manager_image_patch.yaml-e

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
	mkdir -p $(GOBIN)
ifndef HAS_KUBEBUILDER
	curl -L --fail -O \
		"https://github.com/kubernetes-sigs/kubebuilder/releases/download/v$(KUBEBUILDER_VERSION)/kubebuilder_$(KUBEBUILDER_VERSION)_$(PLATFORM)_$(ARCH).tar.gz" && \
		tar -zxvf kubebuilder_$(KUBEBUILDER_VERSION)_$(PLATFORM)_$(ARCH).tar.gz && \
		rm kubebuilder_$(KUBEBUILDER_VERSION)_$(PLATFORM)_$(ARCH).tar.gz && \
		mv ./kubebuilder_$(KUBEBUILDER_VERSION)_$(PLATFORM)_$(ARCH)/bin/* $(GOBIN) && \
		rm -rf ./kubebuilder_$(KUBEBUILDER_VERSION)_$(PLATFORM)_$(ARCH)
endif
ifndef HAS_CONTROLLER_GEN
	cp tools.mod $(GOBIN)/go.mod
	cp tools.sum $(GOBIN)/go.sum
	cd $(GOBIN) && GOBIN=$(GOBIN) go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.1
	cd $(BUILD_DIR)
endif

GO_SOURCES = $(shell find . -type f -name '*.go' ! -path '**/mocks/*')

ifeq ($(OS),Windows_NT)
	OUTPUT=tck.exe
else
	OUTPUT=tck
endif

.PHONY: all
all: build test

.PHONY: build
build: $(OUTPUT) ## Build the executable for current architecture (local dev)

$(OUTPUT): $(GO_SOURCES)
	go build -o $(OUTPUT) -gcflags="all=-N -l" main.go

.PHONY: clean
clean: ## Clean generated files
	rm -f $(OUTPUT)
	rm -f streaming-http-adapter-linux-amd64.tgz

.PHONY: test
test: fmt vet ## Run tests
	go test ./... -timeout 30s -coverprofile cover.out

# Run go fmt against code
.PHONY: fmt
fmt: goimports
	$(GOIMPORTS) -w --local github.com/projectriff main.go framework/

# Run go vet against code
.PHONY: vet
vet:
	go vet ./...

# find or download goimports, download goimports if necessary
goimports:
ifeq (, $(shell which goimports))
	# avoid go.* mutations from go get
	cp go.mod go.mod~ && cp go.sum go.sum~
	go get golang.org/x/tools/cmd/goimports@release-branch.go1.13
	mv go.mod~ go.mod && mv go.sum~ go.sum
GOIMPORTS=$(GOBIN)/goimports
else
GOIMPORTS=$(shell which goimports)
endif

# Absolutely awesome: http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help: ## Print help for each make target
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

########## riff specific targets ############
DOCKER_REPO ?= projectriff
BUILDER ?= projectriff/builder:0.6.0-snapshot-20200213090111-227ea45ceb4bd145

.PHONY: samples
samples: java-samples node-samples

.PHONY: java-samples
java-samples:
	pack build --builder $(BUILDER) --path samples/java --env RIFF=true --env RIFF_HANDLER=io.projectriff.tck.Counter        $(DOCKER_REPO)/counter-java
	pack build --builder $(BUILDER) --path samples/java --env RIFF=true --env RIFF_HANDLER=io.projectriff.tck.HundredDivider $(DOCKER_REPO)/divider-java
	pack build --builder $(BUILDER) --path samples/java --env RIFF=true --env RIFF_HANDLER=io.projectriff.tck.MD5            $(DOCKER_REPO)/md5-java
	pack build --builder $(BUILDER) --path samples/java --env RIFF=true --env RIFF_HANDLER=io.projectriff.tck.Uppercase      $(DOCKER_REPO)/uppercase-java

.PHONY: node-samples
node-samples:
	pack build --builder $(BUILDER) --path samples/node --env RIFF=true --env RIFF_ARTIFACT=counter.js   $(DOCKER_REPO)/counter-node
	pack build --builder $(BUILDER) --path samples/node --env RIFF=true --env RIFF_ARTIFACT=divider.js   $(DOCKER_REPO)/divider-node
	pack build --builder $(BUILDER) --path samples/node --env RIFF=true --env RIFF_ARTIFACT=md5.js       $(DOCKER_REPO)/md5-node
	pack build --builder $(BUILDER) --path samples/node --env RIFF=true --env RIFF_ARTIFACT=uppercase.js $(DOCKER_REPO)/uppercase-node

.PHONY: push-samples
push-samples: push-java-samples push-node-samples

.PHONY: push-java-samples
push-java-samples: java-samples
	docker push $(DOCKER_REPO)/counter-java
	docker push $(DOCKER_REPO)/divider-java
	docker push $(DOCKER_REPO)/md5-java
	docker push $(DOCKER_REPO)/uppercase-java

.PHONY: push-node-samples
push-node-samples: node-samples
	docker push $(DOCKER_REPO)/counter-node
	docker push $(DOCKER_REPO)/divider-node
	docker push $(DOCKER_REPO)/md5-node
	docker push $(DOCKER_REPO)/uppercase-node

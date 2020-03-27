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

framework/rpc/riff-rpc.pb.go: ../riff-rpc.proto
	protoc -I .. riff-rpc.proto --go_out=plugins=grpc:framework/rpc

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
	cp go.mod go.mod~ && cp go.sum go.sum~ # avoid go.* mutations from go get
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
BUILDER ?= projectriff/builder:0.6.0-snapshot-20200324134805-fbecd0add506d0df

.PHONY: samples
samples: java-samples node-samples command-samples

.PHONY: java-samples
java-samples:
	pack build --builder $(BUILDER) --path samples/java --env RIFF=true --env RIFF_HANDLER=io.projectriff.tck.Counter        $(DOCKER_REPO)/counter-java
	pack build --builder $(BUILDER) --path samples/java --env RIFF=true --env RIFF_HANDLER=io.projectriff.tck.HundredDivider $(DOCKER_REPO)/divider-java
	pack build --builder $(BUILDER) --path samples/java --env RIFF=true --env RIFF_HANDLER=io.projectriff.tck.MD5            $(DOCKER_REPO)/md5-java
	pack build --builder $(BUILDER) --path samples/java --env RIFF=true --env RIFF_HANDLER=io.projectriff.tck.Uppercase      $(DOCKER_REPO)/uppercase-java
	pack build --builder $(BUILDER) --path samples/java --env RIFF=true --env RIFF_HANDLER=io.projectriff.tck.Repeater       $(DOCKER_REPO)/repeater-java

.PHONY: node-samples
node-samples:
	pack build --builder $(BUILDER) --path samples/node --env RIFF=true --env RIFF_ARTIFACT=counter.js   $(DOCKER_REPO)/counter-node
	pack build --builder $(BUILDER) --path samples/node --env RIFF=true --env RIFF_ARTIFACT=divider.js   $(DOCKER_REPO)/divider-node
	pack build --builder $(BUILDER) --path samples/node --env RIFF=true --env RIFF_ARTIFACT=md5.js       $(DOCKER_REPO)/md5-node
	pack build --builder $(BUILDER) --path samples/node --env RIFF=true --env RIFF_ARTIFACT=uppercase.js $(DOCKER_REPO)/uppercase-node
	pack build --builder $(BUILDER) --path samples/node --env RIFF=true --env RIFF_ARTIFACT=repeater.js	 $(DOCKER_REPO)/repeater-node

.PHONY: command-samples
command-samples:
	#pack build --builder $(BUILDER) --path samples/command --env RIFF=true --env RIFF_ARTIFACT=counter.sh   $(DOCKER_REPO)/counter-command
	#pack build --builder $(BUILDER) --path samples/command --env RIFF=true --env RIFF_ARTIFACT=divider.sh   $(DOCKER_REPO)/divider-command
	pack build --builder $(BUILDER) --path samples/command --env RIFF=true --env RIFF_ARTIFACT=md5.sh       $(DOCKER_REPO)/md5-command
	pack build --builder $(BUILDER) --path samples/command --env RIFF=true --env RIFF_ARTIFACT=uppercase.sh $(DOCKER_REPO)/uppercase-command

.PHONY: push-samples
push-samples: push-java-samples push-node-samples push-command-samples

.PHONY: push-java-samples
push-java-samples: java-samples
	docker push $(DOCKER_REPO)/counter-java
	docker push $(DOCKER_REPO)/divider-java
	docker push $(DOCKER_REPO)/md5-java
	docker push $(DOCKER_REPO)/uppercase-java
	docker push $(DOCKER_REPO)/repeater-java

.PHONY: push-node-samples
push-node-samples: node-samples
	docker push $(DOCKER_REPO)/counter-node
	docker push $(DOCKER_REPO)/divider-node
	docker push $(DOCKER_REPO)/md5-node
	docker push $(DOCKER_REPO)/uppercase-node

.PHONY: push-command-samples
push-command-samples: command-samples
	#docker push $(DOCKER_REPO)/counter-command
	#docker push $(DOCKER_REPO)/divider-command
	docker push $(DOCKER_REPO)/md5-command
	docker push $(DOCKER_REPO)/uppercase-command

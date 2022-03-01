
GOPATH := $(shell go env GOPATH)
GOBIN := $(GOPATH)/bin

swagger-ui:
	@echo "generate swagger..."
	@swag init -g ./server/server.go --output ./swagger

build: go.sum
ifeq ($(OS),Windows_NT)
	@echo "building mc-subscriber binary..."
	@go build -mod=readonly $(BUILD_FLAGS) -o build/mc-subscriber.exe main.go
else
	@echo "building mc-subscriber binary..."
	@go build -mod=readonly $(BUILD_FLAGS) -o build/mc-subscriber main.go
endif

.PHONY: build swagger-ui
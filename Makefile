HOST_NAME        ?=thingz.io
RELEASE_VERSION  ?=v0.2.1
GRPC_PORT        ?=50505
HTTP_PORT        ?=8080
IMAGE_NAME       ?=grpc-ping
IMAGE_OWNER      ?=$(shell git config --get user.username)

.PHONY: all 
all: help

.PHONY: init 
init: ## Setup deps
	go install \
	  github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
	  github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
	  google.golang.org/protobuf/cmd/protoc-gen-go \
	  google.golang.org/grpc/cmd/protoc-gen-go-grpc

.PHONY: protos 
protos: ## Generats gRPC proto clients
	protoc \
	  --proto_path=proto proto/v1/*.proto \
	  --go_out=:pkg/api/v1 \
	  --go-grpc_out=:pkg/api/v1 \
	  --grpc-gateway_out=:pkg/api/v1 \
	  --openapiv2_out=:swagger

.PHONY: certs
certs: ## Create wildcard TLS certificates using letsencrypt for k8s ingress
	sudo certbot certonly --manual --preferred-challenges dns -d "*.${HOST_NAME}"
	sudo cp "/etc/letsencrypt/live/${HOST_NAME}/fullchain.pem" certs/ingress-cert.pem	
	sudo cp "/etc/letsencrypt/live/${HOST_NAME}/privkey.pem" certs/ingress-key.pem
	sudo chmod 644 certs/*.pem

.PHONY: tidy 
tidy: ## Updates go modules and vendors deps
	go mod tidy
	go mod vendor

.PHONY: test 
test: tidy ## Tests the entire project 
	go test -count=1 -race -covermode=atomic -coverprofile=cover.out \
	  ./...

.PHONY: cover 
cover: test ## Runs test and displays test coverage 
	go tool cover -html=cover.out

.PHONY: debug
debug: ## Sets up gRPC degub env vars
	export GRPC_TRACE=all 
	export GRPC_VERBOSITY=DEBUG 
	export GRPC_GO_LOG_VERBOSITY_LEVEL=2 
	export GRPC_GO_LOG_SEVERITY_LEVEL=info 
	export DEBUG=true

.PHONY: server
server: tidy debug ## Starts the Ping server using gRPC protocol
    GRPC_PORT=$(GRPC_PORT) \
    HTTP_PORT=$(HTTP_PORT) \
    go run cmd/server/main.go

.PHONY: client 
client: tidy ## Starts the Ping client
	go run cmd/client/main.go \
	  --address=localhost:$(GRPC_PORT) \
	  --client="ping-client" \
	  --debug=true

.PHONY: stream 
stream: tidy ## Runs Ping client in streaming mode
	go run cmd/client/main.go \
	  --address=localhost:$(GRPC_PORT) \
	  --client="stream-client" \
	  --stream=100 \
	  --debug=true

.PHONY: gping
gping: ## Invokes ping method using grpcurl
	grpcurl -plaintext \
	  -d '{"id":"id1", "message":"hello"}' \
	  -authority="ping.${HOST_NAME}" \
	  localhost:$(GRPC_PORT) \
	  io.thingz.grpc.v1.Service/Ping

.PHONY: hping
hping: ## Invokes ping method using curl
	curl -i -k -d '{"id":"id1", "message":"hello"}'\
      -H "Content-type: application/json" \
      http://localhost:$(HTTP_PORT)/v1/ping

.PHONY: spell 
spell: ## Checks spelling across the entire project 
	# go get github.com/client9/misspell/cmd/misspell
	misspell -locale="US" -error -source="text" **/*

.PHONY: lint 
lint: ## Lints the entire project
	# brew install golangci-lint
	golangci-lint run --timeout=3m

.PHONY: image
image: tidy ## Builds and publish image 
	docker build \
		-f build/Dockerfile \
		-t "ghcr.io/$(IMAGE_OWNER)/$(IMAGE_NAME):$(RELEASE_VERSION)" .
	docker push "ghcr.io/$(IMAGE_OWNER)/$(IMAGE_NAME):$(RELEASE_VERSION)"

.PHONY: tag 
tag: ## Creates release tag 
	git tag $(RELEASE_VERSION)
	git push origin $(RELEASE_VERSION)

.PHONY: clean 
clean: ## Cleans go and generated files
	go clean
	rm -f certs/*

.PHONY: test  
help: ## Display available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk \
		'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'



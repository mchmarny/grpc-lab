APP_NAME         ?=ping
HOST_NAME        ?=thingz.io
RELEASE_VERSION  ?=v0.0.2
SERVER_ADDRESS   ?=:50505
IMAGE_NAME       ?=grpc-ping
IMAGE_OWNER      ?=$(shell git config --get user.username)

.PHONY: all 
all: test

.PHONY: protos 
protos: ## Generats gRPC proto clients
	protoc \
	  --proto_path=proto proto/*.proto \
	  --go_out=pkg/proto/v1 \
	  --go_opt=paths=source_relative \
	  --go-grpc_out=pkg/proto/v1 \
	  --go-grpc_opt=paths=source_relative

.PHONY: certs
certs: ## Create wildcard TLS certificates using letsencrypt for k8s ingress
	sudo certbot certonly --manual --preferred-challenges dns -d "*.${HOST_NAME}"
	sudo cp "/etc/letsencrypt/live/${HOST_NAME}/fullchain.pem" certs/ingress-cert.pem	
	sudo cp "/etc/letsencrypt/live/${HOST_NAME}/privkey.pem" certs/ingress-key.pem
	sudo chmod 644 certs/*.pem

.PHONY: tidy 
tidy: ## Updates the go modules
	go mod tidy
	go mod vendor

.PHONY: test 
test: tidy ## Tests the entire project 
	go test -v -count=1 -race -covermode=atomic -coverprofile=coverage.txt \
	  ./...

.PHONY: server 
server: tidy ## Starts the Ping server
	ADDRESS=$(SERVER_ADDRESS) \
	DEBUG=true \
	go run cmd/server/main.go

.PHONY: client 
client: tidy ## Starts the Ping client
	go run cmd/client/main.go \
	  --address=$(SERVER_ADDRESS) \
	  --client="${APP_NAME}-client" \
	  --debug=true

.PHONY: stream 
stream: tidy ## Starts the Ping client
	go run cmd/client/main.go \
	  --address=$(SERVER_ADDRESS) \
	  --client="${APP_NAME}-client" \
	  --stream=100 \
	  --debug=true

.PHONY: call
call: ## Lists processes using the app addresss
	grpcurl -plaintext \
	  -d '{"id":"id1", "message":"hello"}' \
	  -authority="${APP_NAME}.${HOST_NAME}" \
	  $(SERVER_ADDRESS) \
	  io.thingz.grpc.v1.Service/Ping

.PHONY: spellcheck 
spellcheck: ## Checks spelling across the entire project 
	@command -v misspell > /dev/null 2>&1 || (cd tools && go get github.com/client9/misspell/cmd/misspell)
	@misspell -locale="US" -error -source="text" **/*

.PHONY: cover 
cover: tidy ## Displays test coverage in the service and service packages
	go test -coverprofile=cover.out ./service && go tool cover -html=cover.out

.PHONY: lint 
lint: ## Lints the entire project
	golangci-lint run --timeout=3m

.PHONY: image
image: tidy ## Builds and publish image 
	docker build \
		-f cmd/server/Dockerfile \
		-t "ghcr.io/$(IMAGE_OWNER)/$(IMAGE_NAME):$(RELEASE_VERSION)" .
	docker push "ghcr.io/$(IMAGE_OWNER)/$(IMAGE_NAME):$(RELEASE_VERSION)"

.PHONY: tag 
tag: ## Creates release tag 
	git tag $(RELEASE_VERSION)
	git push origin $(RELEASE_VERSION)

.PHONY: clean 
clean: ## Cleans go and generated files
	go clean
	rm -f certs/server.*

.PHONY: pids
pids: ## Lists processes using the app addresss
	sudo lsof -i $(SERVER_ADDRESS)
	# kill -9 <pid>

.PHONY: test  
help: ## Display available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk \
		'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'



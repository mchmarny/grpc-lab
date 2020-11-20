RELEASE_VERSION  =v0.0.1

.PHONY: all 
all: test

.PHONY: certs 
certs: ## Updates the go modules
	# brew install certstrap
	certstrap init --common-name "demo.thingz.io" --key-bits 4096 --passphrase ""

.PHONY: openssl-certs 
openssl-certs: ## Updates the go modules
	openssl genrsa -out certs/server.key 4096
	openssl req -new -x509 -sha256 -key certs/server.key -out certs/server.crt -days 3650
	openssl req -new -sha256 -key certs/server.key -out certs/server.csr
	openssl x509 -req -sha256 -in certs/server.csr -signkey certs/server.key -out certs/server.crt -days 3650

.PHONY: tidy 
tidy: ## Updates the go modules
	go mod tidy
	go mod vendor

.PHONY: test 
test: tidy ## Tests the entire project 
	go test -v \
			-count=1 \
			-race \
			-coverprofile=coverage.txt \
			-covermode=atomic \
			./...

.PHONY: server 
server: tidy ## Starts the Ping Server
	go run cmd/server/main.go --debug true

.PHONY: client 
client: tidy ## Starts the Ping Server
	go run cmd/client/main.go --debug true

.PHONY: spellcheck 
spellcheck: ## Checks spelling across the entire project 
	@command -v misspell > /dev/null 2>&1 || (cd tools && go get github.com/client9/misspell/cmd/misspell)
	@misspell -locale="US" -error -source="text" **/*

.PHONY: cover 
cover: tidy ## Displays test coverage in the service and service packages
	go test -coverprofile=cover-service.out ./service && go tool cover -html=cover-service.out

.PHONY: lint 
lint: ## Lints the entire project
	golangci-lint run --timeout=3m

.PHONY: docs 
docs: ## Runs godoc (in container due to mod support)
	docker run \
			--rm \
			-e "GOPATH=/tmp/go" \
			-p 127.0.0.1:$(GDOC_PORT):$(GDOC_PORT) \
			-v $(PWD):/tmp/go/src/ \
			--name godoc golang \
			bash -c "go get golang.org/x/tools/cmd/godoc && echo http://localhost:7777/pkg/ && /tmp/go/bin/godoc -http=:7777"
	open http://localhost:7777/pkg/service/

.PHONY: tag 
tag: ## Creates release tag 
	git tag $(RELEASE_VERSION)
	git push origin $(RELEASE_VERSION)

.PHONY: clean 
clean: ## Cleans go and generated files
	go clean

.PHONY: protos 
protos: ## Generats gRPC proto clients
	protoc --go_out=. --go_opt=paths=source_relative \
	  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	  pkg/proto/v1/service/service.proto

.PHONY: test  
help: ## Display available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk \
		'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

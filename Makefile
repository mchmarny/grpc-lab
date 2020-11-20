RELEASE_VERSION   =v0.0.1
APP_NAME         ?=demo
HOST_NAME        ?=grpc.thingz.io
SERVER_ADDRESS   ?=:50505

.PHONY: all 
all: test

.PHONY: certs
certs: ## Updates the go modules
	rm -f certs/*.pem
	openssl req -x509 -newkey rsa:4096 -days 365 -nodes \
	  -keyout certs/ca-key.pem \
	  -out certs/ca-cert.pem \
	  -subj "/C=US/ST=Oregon/L=Portland/O=Demo/OU=Dev/CN=*.${HOST_NAME}/emailAddress=demo@thingz.io"
	openssl x509 -in certs/ca-cert.pem -noout -text
	# server
	openssl req -newkey rsa:4096 -nodes \
	  -keyout certs/server-key.pem \
	  -out certs/server-req.pem \
	  -subj "/C=US/ST=Oregon/L=Portland/O=Demo/OU=Dev/CN=*.${HOST_NAME}/emailAddress=demo@thingz.io"
	openssl x509 -req -days 365 \
	  -in certs/server-req.pem \
	  -CA certs/ca-cert.pem \
	  -CAkey certs/ca-key.pem \
	  -CAcreateserial \
	  -out certs/server-cert.pem \
	  -extfile certs/cert.cnf
	openssl x509 -in certs/server-cert.pem -noout -text
	# client 
	openssl req -newkey rsa:4096 -nodes \
	  -keyout certs/client-key.pem \
	  -out certs/client-req.pem \
	  -subj "/C=US/ST=Oregon/L=Portland/O=Demo/OU=Dev/CN=*.${HOST_NAME}/emailAddress=demo@thingz.io"
	openssl x509 -req -days 60 \
	  -in certs/client-req.pem \
	  -CA certs/ca-cert.pem \
	  -CAkey certs/ca-key.pem \
	  -CAcreateserial \
	  -out certs/client-cert.pem \
	  -extfile certs/cert.cnf
	openssl x509 -in certs/client-cert.pem -noout -text

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
	go run cmd/server/main.go \
	  --address=$(SERVER_ADDRESS) \
	  --debug=true

.PHONY: client 
client: tidy ## Starts the Ping client
	go run cmd/client/main.go \
	  --address=$(SERVER_ADDRESS) \
	  --client="${APP_NAME}-client" \
	  --debug=true

.PHONY: server-tls
server-tls: tidy ## Starts the Ping server with TLS certs
	GRPC_VERBOSITY=debug GRPC_TRACE=tcp,http,api \
	go run cmd/server/main.go \
	  --address=$(SERVER_ADDRESS) \
	  --ca=certs/ca-cert.pem \
	  --cert=certs/server-cert.pem \
	  --key=certs/server-key.pem \
	  --debug=true

.PHONY: client-tls 
client-tls: tidy ## Starts the Ping client with TLS cert
	GRPC_VERBOSITY=debug GRPC_TRACE=tcp,http,api \
	go run cmd/client/main.go \
	  --address=$(SERVER_ADDRESS) \
	  --host="${APP_NAME}.${HOST_NAME}" \
	  --client="${APP_NAME}-client" \
	  --ca=certs/ca-cert.pem \
	  --cert=certs/client-cert.pem \
	  --key=certs/client-key.pem \
	  --debug=true

.PHONY: pids
pids: ## Lists processes using the app addresss
	sudo lsof -i $(SERVER_ADDRESS)
	# kill -9 <pid>

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

.PHONY: tag 
tag: ## Creates release tag 
	git tag $(RELEASE_VERSION)
	git push origin $(RELEASE_VERSION)

.PHONY: clean 
clean: ## Cleans go and generated files
	go clean
	rm -f certs/server.*

.PHONY: protos 
protos: ## Generats gRPC proto clients
	protoc --go_out=. --go_opt=paths=source_relative \
	  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	  pkg/proto/v1/service/service.proto

.PHONY: test  
help: ## Display available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk \
		'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

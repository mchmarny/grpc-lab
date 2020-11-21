APP_NAME         ?=ping
HOST_NAME        ?=thingz.io
RELEASE_VERSION  ?=v0.0.1
SERVER_ADDRESS   ?=:50505
IMAGE_NAME       ?=grpc-ping
IMAGE_OWNER      ?=$(shell git config --get user.username)


.PHONY: all 
all: test

.PHONY: ingress-certs
ingress-certs: ## Create wildcard TLS certificates using letsencrypt for k8s ingress
	sudo certbot certonly --manual --preferred-challenges dns -d "*.${HOST_NAME}"
	sudo cp "/etc/letsencrypt/live/${HOST_NAME}/fullchain.pem" certs/ingress-cert.pem	
	sudo cp "/etc/letsencrypt/live/${HOST_NAME}/privkey.pem" certs/ingress-key.pem
	sudo chmod 644 certs/${HOST_NAME}/*.pem

.PHONY: certs
certs: ## Creates MTL certificates for use directlty in go
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
	  -extfile cert.cnf
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
	  -extfile cert.cnf
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
	ADDRESS=$(SERVER_ADDRESS) \
	DEBUG=true \
	go run cmd/server/main.go

.PHONY: server-tls
server-tls: tidy ## Starts the Ping server with TLS certs
	GRPC_VERBOSITY=debug GRPC_TRACE=tcp,http,api \
	ADDRESS=$(SERVER_ADDRESS) \
	CA_CERT=certs/ca-cert.pem \
	SERVER_CERT=certs/server-cert.pem \
	SERVER_KEY=certs/server-key.pem \
	DEBUG=true \
	go run cmd/server/main.go

.PHONY: client 
client: tidy ## Starts the Ping client
	go run cmd/client/main.go \
	  --address=$(SERVER_ADDRESS) \
	  --client="${APP_NAME}-client" \
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

.PHONY: call-tls
call-tls: ## Lists processes using the app addresss
	grpcurl \
	 -d '{"id":"id1", "message":"hello"}' \
	 -authority="${APP_NAME}.${HOST_NAME}" \
	 -cacert=certs/ca-cert.pem \
	 -cert=certs/client-cert.pem \
	 -key=certs/client-key.pem \
	 $(SERVER_ADDRESS) \
	 io.thingz.grpc.v1.Service/Ping

.PHONY: call
call: ## Lists processes using the app addresss
	grpcurl -plaintext \
	 -d '{"id":"id1", "message":"hello"}' \
	 -authority="${APP_NAME}.${HOST_NAME}" \
	 $(SERVER_ADDRESS) \
	 io.thingz.grpc.v1.Service/Ping

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

.PHONY: protos 
protos: ## Generats gRPC proto clients
	protoc \
	  --proto_path=proto proto/*.proto \
	  --go_out=pkg/proto/v1 \
	  --go_opt=paths=source_relative \
	  --go-grpc_out=pkg/proto/v1 \
	  --go-grpc_opt=paths=source_relative \
	  --grpc-gateway_out=pkg/proto/v1

.PHONY: test  
help: ## Display available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk \
		'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

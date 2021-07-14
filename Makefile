.PHONY: build-protected-endpoint
build-protected-endpoint: 
	export GO111MODULE=on
	env GOOS=linux go build -ldflags="-s -w" -o bin/protected-endpoint api/protected-endpoint/*.go

.PHONY: deploy-protected-endpoint
deploy-protected-endpoint: build-protected-endpoint
	sls deploy -v -f protected-endpoint

.PHONY: build-authorizer
build-authorizer: 
	export GO111MODULE=on
	env GOOS=linux go build -ldflags="-s -w" -o bin/authorizer api/authorizer/*.go

.PHONY: deploy-authorizer
deploy-authorizer: build-authorizer
	sls deploy -v -f authorizer

.PHONY: build-register
build-register: 
	export GO111MODULE=on
	env GOOS=linux go build -ldflags="-s -w" -o bin/register api/register/*.go

.PHONY: build-login
build-login: 
	export GO111MODULE=on
	env GOOS=linux go build -ldflags="-s -w" -o bin/login api/login/*.go

.PHONY: deploy-register
deploy-register: build-register
	sls deploy -v -f register

.PHONY: build-all
build-all: 
	make build-protected-endpoint
	make build-authorizer
	make build-register
	make build-login

.PHONY: deploy-all
deploy-all: build-all
	sls deploy -v	

	



include .env-api
BIN := $(shell pwd)/bin
VERSION ?= $(shell git rev-parse --short HEAD)
GO?=$(shell which go)
export GOBIN := $(BIN)
export PATH := $(BIN):$(PATH)

BUILD_CMD := $(GO) install -ldflags "-X main.build=${VERSION}"

LOCAL_DEV_PATH = $(shell pwd)/infrastructure/local
DOCKER_COMPOSE_FILE := $(LOCAL_DEV_PATH)/docker-compose.yml
DOCKER_COMPOSE_FILE_INFRA := $(LOCAL_DEV_PATH)/docker-compose-infra.yml
DOCKER_COMPOSE_CMD := docker compose -p issuer -f $(DOCKER_COMPOSE_FILE)
DOCKER_COMPOSE_INFRA_CMD := docker compose -p issuer -f $(DOCKER_COMPOSE_FILE_INFRA)
ENVIRONMENT := ${ISSUER_API_ENVIRONMENT}


# Local environment overrides via godotenv
DOTENV_CMD = $(BIN)/godotenv
ENV = $(DOTENV_CMD) -f .env-issuer

.PHONY: build-local
build-local:
	$(BUILD_CMD) ./cmd/...

.PHONY: build/docker
build/docker: ## Build the docker image.
	DOCKER_BUILDKIT=1 \
	docker build \
		-f ./Dockerfile \
		-t issuer/api:$(VERSION) \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		.

.PHONY: clean
clean: ## Go clean
	$(GO) clean ./...

.PHONY: tests
tests:
	$(GO) test -v ./... --count=1

.PHONY: test-race
test-race:
	$(GO) test -v --race ./...

$(BIN)/oapi-codegen: tools.go go.mod go.sum ## install code generator for API files.
	$(GO) install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

.PHONY: api
api: $(BIN)/oapi-codegen
	$(BIN)/oapi-codegen -config ./api/config-oapi-codegen.yaml ./api/api.yaml > ./internal/api/api.gen.go


.PHONY: api-ui
api-ui: $(BIN)/oapi-codegen
	$(BIN)/oapi-codegen -config ./api_ui/config-oapi-codegen.yaml ./api_ui/api.yaml > ./internal/api_ui/api.gen.go

.PHONY: up
up:
	$(DOCKER_COMPOSE_INFRA_CMD) up -d redis postgres vault

.PHONY: run
run:
	$(eval DELETE_FILE = $(shell if [ -f ./.env-ui ]; then echo "false"; else echo "true"; fi))
	@if [ -f ./.env-ui ]; then echo "false"; else touch ./.env-ui; fi
	COMPOSE_DOCKER_CLI_BUILD=1 DOCKER_FILE="Dockerfile" $(DOCKER_COMPOSE_CMD) up -d api pending_publisher
	@if [ $(DELETE_FILE) = "true" ] ; then rm ./.env-ui; fi

.PHONY: run-arm
run-arm:
	@echo "WARN: Running ARM version is deprecated. 'make run' will be executed instead."
	@make run

.PHONY: run-ui
run-ui: add-host-url-swagger
	COMPOSE_DOCKER_CLI_BUILD=1 DOCKER_FILE="Dockerfile" $(DOCKER_COMPOSE_CMD) up -d api-ui ui notifications pending_publisher --build

.PHONY: run-ui-arm
run-ui-arm: add-host-url-swagger
	@echo "WARN: Running ARM version is deprecated. 'make run-ui' will be executed instead."
	@make run-ui
	
.PHONY: build
build:
	COMPOSE_DOCKER_CLI_BUILD=1 DOCKER_FILE="Dockerfile" $(DOCKER_COMPOSE_CMD) build api pending_publisher

.PHONY: build-arm
build-arm:
	@echo "WARN: Running ARM version is deprecated. 'make build' will be executed instead."
	@make build

.PHONY: build-ui
build-ui:
	COMPOSE_DOCKER_CLI_BUILD=1 DOCKER_FILE="Dockerfile" $(DOCKER_COMPOSE_CMD) build api-ui ui notifications pending_publisher

.PHONY: build-ui-arm
build-ui-arm:
	@echo "WARN: Running ARM version is deprecated. 'make build-ui' will be executed instead."
	@make build-ui

.PHONY: down
down:
	$(DOCKER_COMPOSE_INFRA_CMD) down --remove-orphans
	$(DOCKER_COMPOSE_CMD) down --remove-orphans

.PHONY: stop
stop:
	$(DOCKER_COMPOSE_INFRA_CMD) stop
	$(DOCKER_COMPOSE_CMD) stop

.PHONY: up-test
up-test:
	$(DOCKER_COMPOSE_INFRA_CMD) up -d test_postgres vault test_local_files_apache

.PHONY: clean-vault
clean-vault:
	rm -R infrastructure/local/.vault/data/init.out
	rm -R infrastructure/local/.vault/file/core/
	rm -R infrastructure/local/.vault/file/logical/
	rm -R infrastructure/local/.vault/file/sys/

$(BIN)/platformid-migrate:
	$(BUILD_CMD) ./cmd/migrate

$(BIN)/install-goose: go.mod go.sum
	$(GO) install github.com/pressly/goose/v3

$(BIN)/golangci-lint: go.mod go.sum
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint

$(BIN)/godotenv: tools.go go.mod go.sum
	$(GO) install github.com/joho/godotenv/cmd/godotenv

.PHONY: db/migrate
db/migrate: $(BIN)/install-goose $(BIN)/godotenv $(BIN)/platformid-migrate ## Install goose and apply migrations.
	$(ENV) sh -c '$(BIN)/migrate'

.PHONY: lint
lint: $(BIN)/golangci-lint
	  $(BIN)/golangci-lint run

.PHONY: lint-fix
lint-fix: $(BIN)/golangci-lint
		  $(BIN)/golangci-lint run --fix

# usage: make private_key=xxx add-private-key
.PHONY: add-private-key
add-private-key:
	docker exec issuer-vault-1 \
	vault write iden3/import/pbkey key_type=ethereum private_key=$(private_key)

.PHONY: print-vault-token
print-vault-token:
	$(eval TOKEN = $(shell docker logs issuer-vault-1 2>&1 | grep " .hvs" | awk  '{print $$2}' | tail -1 ))
	@echo $(TOKEN)

.PHONY: add-vault-token
add-vault-token:
	$(eval TOKEN = $(shell docker logs issuer-vault-1 2>&1 | grep " .hvs" | awk  '{print $$2}' | tail -1 ))
	sed '/ISSUER_KEY_STORE_TOKEN/d' .env-issuer > .env-issuer.tmp
	@echo ISSUER_KEY_STORE_TOKEN=$(TOKEN) >> .env-issuer.tmp
	mv .env-issuer.tmp .env-issuer


.PHONY: run-initializer
run-initializer:
	COMPOSE_DOCKER_CLI_BUILD=1 DOCKER_FILE="Dockerfile" $(DOCKER_COMPOSE_CMD) up -d initializer
	sleep 5

.PHONY: generate-issuer-did
generate-issuer-did: run-initializer
	docker logs issuer-initializer-1
	$(eval DID = $(shell docker logs -f --tail 1 issuer-initializer-1 | grep "did"))
	@echo $(DID)
	sed '/ISSUER_API_UI_ISSUER_DID/d' .env-api > .env-api.tmp
	@echo ISSUER_API_UI_ISSUER_DID=$(DID) >> .env-api.tmp
	mv .env-api.tmp .env-api
	docker stop issuer-initializer-1
	docker rm issuer-initializer-1


.PHONY: generate-issuer-did-arm
generate-issuer-did-arm:
	@echo "WARN: Running ARM version is deprecated. 'make generate-issuer-did' will be executed instead."
	@make generate-issuer-did

.PHONY: add-host-url-swagger
add-host-url-swagger:
	@if [ $(ENVIRONMENT) != "" ] && [ $(ENVIRONMENT) != "local" ]; then \
		sed -i -e  "s#server-url = [^ ]*#server-url = \""${ISSUER_API_UI_SERVER_URL}"\"#g" api_ui/spec.html; \
	fi

.PHONY: rm-issuer-imgs
rm-issuer-imgs: stop
	$(DOCKER_COMPOSE_CMD) rm -f
	docker rmi -f issuer-api issuer-ui issuer-api-ui issuer-pending_publisher

.PHONY: restart-ui
restart-ui: rm-issuer-imgs up run run-ui

.PHONY: restart-ui-arm
restart-ui-arm:
	@echo "WARN: Running ARM version is deprecated. 'make restart-ui' will be executed instead."
	@make restart-ui

.PHONY: print-did
print-did:
	docker exec issuer-vault-1 \
	vault kv get -mount=kv did

# use this to delete the did from vault. It will not be deleted from the database
.PHONY: delete-did
delete-did:
	docker exec issuer-vault-1 \
	vault kv delete kv/did

# use this to add the did to vault. It will not be added to the database
# usage: make did=xxx add-did
.PHONY: add-did
add-did:
	docker exec issuer-vault-1 \
	vault kv put kv/did did=$(did)

# usage: make vault_token=xxx vault-export-keys
.PHONY: vault-export-keys
vault-export-keys:
	docker build -t issuer-vault-export-keys .
	docker run --rm -it --network=issuer-network -v $(shell pwd):/keys issuer-vault-export-keys ./vault-migrator -operation=export -output-file=keys.json -vault-token=$(vault_token) -vault-addr=http://vault:8200

# usage: make vault_token=xxx vault-import-keys
.PHONY: vault-import-keys
vault-import-keys:
	docker build -t issuer-vault-import-keys .
	docker run --rm -it --network=issuer-network -v $(shell pwd)/keys.json:/keys.json issuer-vault-import-keys ./vault-migrator -operation=import -input-file=keys.json -vault-token=$(vault_token) -vault-addr=http://vault:8200


# usage: make new_password=xxx change-vault-password
.PHONY: change-vault-password
change-vault-password:
	docker exec issuer-vault-1 \
	vault write auth/userpass/users/issuernode password=$(new_password)
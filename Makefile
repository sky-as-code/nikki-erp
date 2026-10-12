env ?= local
goarch ?= amd64
goos ?= windows
outfile ?= nikki-erp.exe

cwdnikki := $(dir $(lastword $(MAKEFILE_LIST)))

ifndef cwd
cwd := $(dir $(lastword $(MAKEFILE_LIST)))
endif

.PHONY: build build-mods build-static build-dynamic clean \
	db-init db-gen db-current db-hash db-migration db-apply db-revert db-infra-apply \
	infra-up infra-down infra-swarm-up infra-swarm-down infra-swarm-svc install-tools gen-sql run \
	test-api-rest test-api-rest-deps cert-ca cert-client cert-server-nikki keypair-jwt-nikki


# START: Go builds

clean:
	@echo "Cleaning dist directory..."
	@rm -rf dist
	@echo "Clean completed"

# Production build: static binary (modules linked in with staticmods) into dist/.
build-static: clean
	@echo "Building static binary..."
	GOOS=$(goos) GOARCH=$(goarch) go build -tags staticmods -work -o dist/$(outfile) main.go
	@echo "Static build completed"

# Build all modules as plugins
build-mods:
	@echo "Building all modules..."
	@mkdir -p ./modules
	@for module_dir in ./modules/*/; do \
		if [ -d "$$module_dir" ]; then \
			module_name=$$(basename "$$module_dir"); \
			echo "Building $$module_name..."; \
			CGO_ENABLED=1 GOOS=$(goos) GOARCH=$(goarch) go build -buildmode=plugin -o ./modules/$$module_name.so ./modules/$$module_name; \
		fi \
	done
	@echo "All modules built successfully"

build-dynamic: build-mods clean
	@echo "Building dynamic binary..."
	GOOS=$(goos) GOARCH=$(goarch) go build -tags dynamicmods -work -tags dynamicmods -o dist/$(outfile) main.go
	@cp -rf ./modules dist/
	@echo "Dynamic build completed"

# Build application and copy config files
build: build-dynamic
	@echo "Copying config files..."
	@mkdir -p dist/config
	@cp ./config/config.json dist/config/config.json
	@echo "Build completed successfully"

# END: Go builds

# START: ORM & Database
migration_dir := file://${cwd}scripts/migrations

# Legacy: db-init and db-gen generate infra/ent/ code. Do not use; model schemas are JSON now.
db-init:
	@if [ -z "$(module)" ]; then \
		echo "Error: module folder is required. Usage: make db-init module=<module_name> name=<schema_name>"; \
		exit 1; \
	fi
	@if [ -z "$(name)" ]; then \
		echo "Error: Entity name (in PascalCase) is required. Usage: make db-init module=<module_name> name=<schema_name>"; \
		exit 1; \
	fi
	@module_path="./modules/$(module)/infra/ent/schema"; \
	echo "Initializing ent schema '$(name)' in '$$module_path'..."; \
	go run -mod=mod entgo.io/ent/cmd/ent new $(name) --target $$module_path; \
	printf "package ent\n\n//go:generate go run -mod=mod entgo.io/ent/cmd/ent generate ./schema\n" > "./modules/$(module)/infra/ent/generate.go"

db-gen:
	@if [ -z "$(module)" ]; then \
		echo "Error: module parameter is required. Usage: make db-gen module=<module_name>"; \
		exit 1; \
	fi
	@if [ ! -d "./modules/$(module)/infra/ent" ]; then \
		echo "Error: ent schema directory not found for module '$(module)'"; \
		exit 1; \
	fi
	@echo "Generating ent code for module '$(module)'..."
	go generate ./modules/$(module)/infra/ent

db-current:
	@echo "Generating script of current state to '$(migration_dir)'..."
	atlas migrate diff current_state.tmp \
		--dir "$(migration_dir)" \
		--to "postgres://nikki_admin:nikki_password@localhost:5432/nikki_erp?sslmode=disable" \
		--config file://${cwd}scripts/atlas.hcl \
		--env local

# Re-hash atlas.sum after a migration file changed.
db-hash:
	@echo "Hashing migrations in '$(migration_dir)'..."
	@atlas migrate hash --dir "$(migration_dir)"

# Generate the module's migration SQL into $(cwd)scripts/migrations: run from the repo root for the
# multi-tenant tree, from nikkierp/ for the single-tenant tree.
db-migration:
	@if [ -z "$(module)" ]; then \
		echo "Error: module parameter is required. Usage: make db-migration module=<module_name>"; \
		exit 1; \
	fi

	"${cwd}scripts/db-migrations.sh" "$(module)" '${cwd}'

# Apply pending migrations to local postgres.
db-apply:
	@echo "Applying migration files in '$(migration_dir)'..."
	atlas migrate apply \
		--dir "$(migration_dir)" \
		--url "postgres://nikki_admin:nikki_password@localhost:5432/nikki_erp?search_path=public&sslmode=disable"

# Roll back the latest applied migration file.
db-revert:
	@echo "Undoing the LATEST APPLIED migration file in '$(migration_dir)'..."
	atlas migrate down \
		--dir "$(migration_dir)" \
		--url "postgres://nikki_admin:nikki_password@localhost:5432/nikki_erp?search_path=public&sslmode=disable" \
		--config file://${cwd}scripts/atlas.hcl \
		--env local

db-infra-apply:
	make infra-down && make infra-up && make db-hash && make db-apply

# END: ORM & Database

# START: Local development

# Start local infra (docker compose); allow 10-20s for initialization.
infra-up:
	"${cwdnikki}scripts/infra-up.sh" "${cwd}scripts/docker/docker-compose.local.yml" "${cwd}config"

# Stop local infra AND delete its volumes (down -v): local data is lost.
infra-down:
	docker compose -f "${cwd}scripts/docker/docker-compose.local.yml" down -v

infra-swarm-up:
	./scripts/docker/infra-up.sh $(shell pwd)

infra-swarm-down:
	./scripts/docker/infra-down.sh

infra-swarm-svc:
	@if [ -z "$(svc)" ]; then \
		echo "Error: svc parameter is required. Usage: make infra-swarm-svc svc=<service_name>"; \
		exit 1; \
	fi
	docker service ps nikki_infra_$(svc) --no-trunc --format "table {{.ID}}\t{{.Node}}\t{{.DesiredState}}\t{{.CurrentState}}\t{{.Error}}"

install-tools:
	go install go.uber.org/mock/mockgen@latest
# curl -sSf https://atlasgo.sh | sh

gen-sql:
	@[ -f config/local.env ] || cp config/local.env.sample config/local.env
	go run -tags=staticmods $(cwd)*.go -createsql -dialect=postgres -module=$(module)

run:
	@[ -f config/local.env ] || cp config/local.env.sample config/local.env
	APP_ENV=$(env) WORKING_DIR="$(cwd)" GENERAL_LOG_LEVEL="debug" go run -tags=staticmods $(cwd)*.go $(ARGS)

# END: Local development

# START: API testing
robot_outdir ?= $(cwd)tests/api-rest/output
# Falls back to "python -m robot" when the robot launcher is not on PATH.
robot_bin ?= $(shell command -v robot >/dev/null 2>&1 && echo robot || echo "python -m robot")

# Robot Framework API tests. Runs the tree under the CALLING app's tests/api-rest
# (cd coremart && make test-api-rest -> coremart tests; from nikkierp -> nikkierp tests).
# Optional params:
#   t=<folder-or-file>   relative to tests/api-rest, e.g. t=iam/user or t=iam/user/03_update.robot
#   env=<name>           environment variable file, default "local"
#   include=<tag>        Robot --include tag filter
#   robot_args=...       extra raw robot arguments
test-api-rest:
	@target="$(cwd)tests/api-rest$(if $(t),/$(t),)"; \
	echo "Running Robot API tests: $$target (env=$(env))"; \
	$(robot_bin) \
		--pythonpath "$(cwdnikki)tests/api-rest" \
		--variablefile "$(cwd)tests/api-rest/environments/$(env).py" \
		--outputdir "$(robot_outdir)" \
		--name "api-rest" \
		$(if $(include),--include $(include),) \
		$(robot_args) \
		"$$target"

# Run once: install the API-test requirements.
test-api-rest-deps:
	pip install -r "$(cwdnikki)tests/api-rest/requirements.txt"
# END: API testing

# START: Certificate

# Local PKI: cert-ca (root, client and server CA), cert-client, cert-server-nikki. Run by a human, not an agent.
cert-ca:
	@set -euo pipefail; \
	echo "==> Root CA"; \
	CWD="${cwd}scripts/cert" SDIR="${cwdnikki}scripts/cert" ${cwdnikki}scripts/cert/gen-root-ca.sh

	@set -euo pipefail; \
	echo "==> Client CA"; \
	CWD="${cwd}scripts/cert" SDIR="${cwdnikki}scripts/cert" ${cwdnikki}scripts/cert/gen-intermediate-ca.sh client-ca

	@set -euo pipefail; \
	echo "==> Server CA"; \
	CWD="${cwd}scripts/cert" SDIR="${cwdnikki}scripts/cert" ${cwdnikki}scripts/cert/gen-intermediate-ca.sh server-ca

cert-client:
	@set -euo pipefail; \
	echo "==> Client cert"; \
	CWD="${cwd}scripts/cert" SDIR="${cwdnikki}scripts/cert" ${cwdnikki}scripts/cert/gen-client-cert.sh

cert-server-nikki:
	@set -euo pipefail; \
	echo "==> Server cert"; \
	CWD="${cwd}scripts/cert" SDIR="${cwdnikki}scripts/cert" ./scripts/cert/gen-server-cert.sh

keypair-jwt-nikki:
	@set -euo pipefail; \
	echo "==> JWT keypair"; \
	CWD="${cwd}scripts/cert" ./scripts/cert/gen-keypair-ed25519.sh

# END: Certificate

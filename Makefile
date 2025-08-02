ifneq (,$(wildcard ./.env))
    include .env
    export
endif

SCRIPT_FOLDER = migrations
MIGRATION_FOLDER= migrations/postgres
GOLANG_MIGRATE_VERSION = 4.18.3
GOLANG_MIGRATE_LINUX_ZIP = migrate.linux-amd64.tar.gz
POSTGRES_DSN = postgresql://$(POSTGRES_USERNAME):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DATABASE)?sslmode=disable

install-migrate-linux:
	curl -OL https://github.com/golang-migrate/migrate/releases/download/v$(GOLANG_MIGRATE_VERSION)/$(GOLANG_MIGRATE_LINUX_ZIP)
	sudo tar xvf $(GOLANG_MIGRATE_LINUX_ZIP) -C /usr/local/bin/ migrate
	rm -f $(GOLANG_MIGRATE_LINUX_ZIP)

install-migrate-windows:
	# https://scoop.sh/
	scoop install migrate

.PHONY: migrate-create
migrate-create:
	$(SCRIPT_FOLDER)/migrate-db create $(name)

.PHONY: migrate-up
migrate-up:
	$(SCRIPT_FOLDER)/migrate-db up

.PHONY: migrate-down
migrate-down:
	$(SCRIPT_FOLDER)/migrate-db down

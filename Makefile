ifneq (,$(wildcard ./.env))
    include .env
    export
endif

SCRIPT_FOLDER = migrations
MIGRATION_FOLDER = migrations/postgres
GOLANG_MIGRATE_VERSION = 4.18.1
GOLANG_MIGRATE_LINUX_ZIP = migrate.linux-amd64.tar.gz
POSTGRESQL_URI = postgresql://$(PG_USERNAME):$(PG_PASSWORD)@$(PG_HOST):$(PG_PORT)/$(PG_DATABASE)?sslmode=disable

install-migrate-linux:
	curl -OL https://github.com/golang-migrate/migrate/releases/download/v$(GOLANG_MIGRATE_VERSION)/$(GOLANG_MIGRATE_LINUX_ZIP)
	tar xvf $(GOLANG_MIGRATE_LINUX_ZIP) -C /usr/local/bin/ migrate
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

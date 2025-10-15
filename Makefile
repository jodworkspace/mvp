MIGRATIONS_FOLDER=migrations
ENV_FILE=.env

install-goose:
	go install github.com/pressly/goose/v3/cmd/goose@latest
	ls "$(shell go env GOPATH)/bin/" | grep goose

migrate-sql:
	goose -dir=$(MIGRATIONS_FOLDER)/postgres create $(NAME) sql

migrate-go:
	goose -dir=$(MIGRATIONS_FOLDER)/go create $(NAME) go

migrate-up:
	goose -env $(ENV_FILE) up

migrate-down:
	goose -env $(ENV_FILE) down

env-example:
	awk -F'=' 'BEGIN {OFS="="} \
    	/^[[:space:]]*#/ {print; next} \
    	/^[[:space:]]*$$/ {print ""; next} \
    	NF>=1 {gsub(/^[[:space:]]+|[[:space:]]+$$/, "", $$1); print $$1"="}' .env > .env.example
	echo ".env.example generated successfully."

# Default Go version (can be overridden with `make install-go VERSION=1.25.0`)
VERSION ?= 1.25.0
GO_URL = https://go.dev/dl/go$(VERSION).linux-amd64.tar.gz

install-go:
	echo "Installing Go version $(VERSION)..."
	cd ~
	wget -q $(GO_URL) -O go$(VERSION).linux-amd64.tar.gz
	sudo rm -rf /usr/local/go
	sudo tar -C /usr/local -xzf go$(VERSION).linux-amd64.tar.gz
	echo 'export PATH=$$PATH:/usr/local/go/bin' >> ~/.bashrc
	source ~/.bashrc
	/usr/local/go/bin/go version
	rm go$(VERSION).linux-amd64.tar.gz
	echo "Go $(VERSION) installation complete."
# RSJ-V3 Monorepo Makefile

.PHONY: all clean build test fmt lint run-audit

# Build Output Directory
BIN_DIR := bin

# Binaries to build
BINARIES := audit-engine mesh-node rsj-agent rsj-guardian cloud-worker gatekeeper

all: build

build: $(BINARIES)

$(BIN_DIR):
	mkdir -p $(BIN_DIR)

audit-engine: $(BIN_DIR)
	go build -o $(BIN_DIR)/audit-engine ./cmd/audit-engine

mesh-node: $(BIN_DIR)
	go build -o $(BIN_DIR)/mesh-node ./cmd/mesh-node

rsj-agent: $(BIN_DIR)
	go build -o $(BIN_DIR)/rsj-agent ./cmd/rsj-agent

rsj-guardian: $(BIN_DIR)
	go build -o $(BIN_DIR)/rsj-guardian ./cmd/rsj-guardian

cloud-worker: $(BIN_DIR)
	go build -o $(BIN_DIR)/cloud-worker ./services/cloud-worker

gatekeeper: $(BIN_DIR)
	go build -o $(BIN_DIR)/gatekeeper ./services/gatekeeper

clean:
	rm -rf $(BIN_DIR)

test:
	go test -v ./...

run: build
	chmod +x ./run.sh
	./run.sh

fmt:
	go fmt ./...

lint:
	# Assuming golangci-lint is installed, otherwise skip or install it
	# golangci-lint run
	@echo "Linting not configured (golangci-lint missing)"

run-audit: build
	./$(BIN_DIR)/audit-engine

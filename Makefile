APP       = gopaper
ADMIN_DIR = web/admin
BIN_DIR   = bin
VERSION   ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
RELEASE    = -ldflags "-s -w -X main.version=$(VERSION)"
GOBUILD    = go build $(RELEASE)

.PHONY: one all front build run dev clean

one: front
	@echo "Build $(APP) (local) ..."
	mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 $(GOBUILD) -o $(BIN_DIR)/$(APP) ./

all: clean one build

build: front
	@echo "Cross-compiling ..."
	mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64  $(GOBUILD) -o $(BIN_DIR)/$(APP)-$(VERSION).darwin-arm64  ./
	CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64  $(GOBUILD) -o $(BIN_DIR)/$(APP)-$(VERSION).darwin-amd64  ./
	CGO_ENABLED=0 GOOS=linux   GOARCH=arm64  $(GOBUILD) -o $(BIN_DIR)/$(APP)-$(VERSION).linux-arm64   ./
	CGO_ENABLED=0 GOOS=linux   GOARCH=amd64  $(GOBUILD) -o $(BIN_DIR)/$(APP)-$(VERSION).linux-amd64   ./
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64  $(GOBUILD) -o $(BIN_DIR)/$(APP)-$(VERSION).windows-amd64.exe ./
	@echo "✅ Build success."

front:
	@echo "Build frontend ..."
	cd $(ADMIN_DIR) && npm run build

run: front
	go run ./

dev:
	@which air > /dev/null 2>&1 || go install github.com/air-verse/air@latest
	@if [ ! -d "$(ADMIN_DIR)/node_modules" ]; then cd $(ADMIN_DIR) && npm install; fi
	(cd $(ADMIN_DIR) && npm run dev &)
	air

clean:
	rm -rf $(BIN_DIR) tmp/ web/public/admin
	@echo "✅ Clean complete."

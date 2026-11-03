BINARY  := digtast
PKG     := ./cmd
BIN_DIR := bin

GO      ?= go
LDFLAGS := -s -w

.PHONY: all build run test vet fmt tidy clean

all: build

## build: 编译二进制到 bin/digtast
build:
	$(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY) $(PKG)

## run: 直接运行
run:
	$(GO) run $(PKG)

test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

fmt:
	$(GO) fmt ./...

tidy:
	$(GO) mod tidy

clean:
	rm -rf $(BIN_DIR)

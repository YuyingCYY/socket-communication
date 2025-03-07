# 變數
BINARY_SERVER=server
BINARY_CLIENT=client.dll
BUILD_DIR=build
CERT_DIR=cert

# Go 相關變數
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GMMOD=$(GOMOD) mod

# 版本
VERSION?=1.0.0
BUILD_TIME=$(shell date +%FT%T%z)

# 標誌
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}"

.PHONY: all build clean test deps cert help

# 默認目標
all: clean deps cert build test

# 建立目錄
$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)/server
	mkdir -p $(BUILD_DIR)/client

# 建構伺服器
server: $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/server/$(BINARY_SERVER) ./cmd/server

# 建構客戶端
client: $(BUILD_DIR)
	$(GOBUILD) -buildmode=c-shared -o $(BUILD_DIR)/client/$(BINARY_CLIENT) ./cmd/client

# 清理建構檔案
clean:
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f $(CERT_DIR)/*.pem $(CERT_DIR)/*.key $(CERT_DIR)/*.crt

# 執行測試
test:
	$(GOTEST) -v ./...

# 安裝依賴
deps:
	$(GOMOD) download
	$(GOMOD) verify

# 生成 SSL 證書
cert:
	mkdir -p $(CERT_DIR)
	openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
		-keyout $(CERT_DIR)/server.key \
		-out $(CERT_DIR)/server.crt \
		-subj "/C=TW/ST=Taiwan/L=Taipei/O=Supreme/CN=localhost"

# 執行服務器
run-server: server
	./$(BUILD_DIR)/server/$(BINARY_SERVER)

# Dokcer 相關
docker-build:
	docker-compose build

# Docker 啟動服務
docker-up:
	docker-compose up

# Docker 停止服務
docker-down:
	docker-compose down

# 查看 Docker 日誌
docker-logs:
	docker-compose logs -f

# 清理 Docker 容器
docker-clean:
	docker-compose down -v --rmi all

# 幫助信息
help:
	@echo "Make commands:"
	@echo "make all        - 清理、安裝依賴、生成證書、建構並測試"
	@echo "make server     - 建構服務器"
	@echo "make client     - 建構客戶端"
	@echo "make clean      - 清理建構檔案"
	@echo "make test       - 執行測試"
	@echo "make deps       - 安裝依賴"
	@echo "make cert       - 生成 SSL 證書"
	@echo "make run-server - 建構並執行服務器"
	@echo "make docker-build - 建構 Docker 容器"
	@echo "make docker-up  - 啟動 Docker 容器"
	@echo "make docker-down - 停止 Docker 容器"
	@echo "make docker-logs - 查看 Docker 容器日誌"
	@echo "make docker-clean - 清理 Docker 容器"
# Makefile

# 定义变量
APP_NAME := url_shortener
BUILD_DIR := build
COVERAGE_DIR := coverage

# 默认目标
.PHONY: all
all: build

# 安装依赖
.PHONY: deps
deps:
	go mod tidy && go mod vendor

# 运行应用
.PHONY: run
run:
	go run .

# 编译应用
.PHONY: build
build: build-default

.PHONY: build-default
build-default:
	go build -o $(BUILD_DIR)/$(APP_NAME) .

# 运行测试
.PHONY: test
test:
	go test -v ./...

# 运行测试并生成覆盖率报告
.PHONY: test-coverage
test-coverage:
	mkdir -p $(COVERAGE_DIR)
	go test -v -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	go tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html

# 清理生成的文件
.PHONY: clean
clean:
	rm -rf $(BUILD_DIR) $(COVERAGE_DIR)

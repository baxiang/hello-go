# NATS MVP Example Makefile

.PHONY: all build run clean test

# 默认 NATS 地址
NATS_URL ?= nats://localhost:4222

# 构建二进制
all: build

build:
	go build -o nats-mvp-example .

# 运行 Hub
run-hub:
	go run main.go hub --nats=$(NATS_URL)

# 运行设备
run-device:
	go run main.go device --id=device-001 --nats=$(NATS_URL)

# 运行客户端
run-client:
	go run main.go client --device=device-001 --query="ls -la" --nats=$(NATS_URL)

# 列出设备
list:
	go run main.go list --nats=$(NATS_URL)

# 监听设备变化
watch:
	go run main.go list --watch --nats=$(NATS_URL)

# 并发测试
bench:
	go run main.go bench --devices=5 --requests=10 --nats=$(NATS_URL)

# 清理
clean:
	rm -f nats-mvp-example

# 运行测试
test:
	go test ./...

# 下载依赖
deps:
	go mod tidy
	go mod download

# 验证连接
verify:
	@echo "验证 NATS 连接: $(NATS_URL)"
	@go run main.go list --nats=$(NATS_URL) 2>&1 | head -5

# 帮助
help:
	@echo "NATS MVP Example"
	@echo ""
	@echo "用法:"
	@echo "  make run-hub      启动 Hub 服务"
	@echo "  make run-device   启动设备模拟器"
	@echo "  make run-client   发送指令到设备"
	@echo "  make list         列出所有设备"
	@echo "  make watch        监听设备状态变化"
	@echo "  make bench        并发测试"
	@echo ""
	@echo "自定义 NATS 地址:"
	@echo "  make run-hub NATS_URL=nats://localhost:4222"
# NATS 边学边练 Makefile

NATS_URL ?= nats://localhost:4222

.PHONY: all deps subject pubsub rpc jetstream kv clean

all: deps

# 安装依赖
deps:
	go mod tidy

# 第一章：Subject 寻址
subject:
	go run ./01-subjects/main.go

# 第二章：Pub/Sub
pubsub:
	go run ./02-pub-sub/main.go

# 第三章：Request/Reply
rpc:
	go run ./03-request-reply/main.go

# 第四章：JetStream
jetstream:
	go run ./04-jetstream/main.go

# 第五章：KV Store
kv:
	go run ./05-kv/main.go

# 运行所有章节
all-chapters: subject pubsub rpc jetstream kv

# 清理
clean:
	go clean ./...

# 帮助
help:
	@echo "NATS 边学边练"
	@echo ""
	@echo "用法:"
	@echo "  make subject     # 第一章：Subject 寻址"
	@echo "  make pubsub      # 第二章：Pub/Sub 发布订阅"
	@echo "  make rpc         # 第三章：Request/Reply"
	@echo "  make jetstream   # 第四章：JetStream 持久化"
	@echo "  make kv          # 第五章：KV Store"
	@echo "  make all-chapters # 运行所有章节"
	@echo ""
	@echo "NATS 服务器: $(NATS_URL)"
# 9.1 Docker容器化进阶

## Dockerfile最佳实践

### 多阶段构建

```dockerfile
# 第一阶段：构建
FROM golang:1.21-alpine AS builder

WORKDIR /build

# 安装依赖
RUN apk add --no-cache git make

# 复制依赖文件（利用缓存）
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建（优化参数）
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w -X main.Version=1.0.0" \
    -o /build/app ./cmd/myapp

# 第二阶段：运行
FROM alpine:3.18

# 安装必要工具
RUN apk --no-cache add ca-certificates tzdata

# 创建非root用户
RUN adduser -D -g '' appuser

WORKDIR /app

# 从builder复制二进制文件
COPY --from=builder /build/app .
COPY --from=builder /build/configs ./configs

# 切换用户
USER appuser

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s \
    CMD wget -q --spider http://localhost:8080/health || exit 1

EXPOSE 8080

ENTRYPOINT ["./app"]
```

**优势**:
- 最终镜像体积小（Alpine + 二进制）
- 安全性高（无源代码、非root用户）
- 构建缓存优化（分层COPY）

---

## 镜像优化技巧

### 体积优化对比

```dockerfile
# ❌ 基础镜像：800MB
FROM golang:1.21
COPY . .
RUN go build -o app ./cmd/myapp
CMD ["./app"]

# ✅ Alpine镜像：15MB
FROM golang:1.21-alpine AS builder
# ... 多阶段构建
FROM alpine:3.18
# ... 最终镜像

# ✅ Distroless镜像：8MB（最小）
FROM gcr.io/distroless/static-debian12
COPY --from=builder /build/app /
ENTRYPOINT ["./app"]
```

### 优化策略

**1. 使用Alpine或Distroless基础镜像**

```dockerfile
# Alpine（推荐，包含必要工具）
FROM alpine:3.18

# Distroless（最小，适合生产）
FROM gcr.io/distroless/static-debian12
```

**2. 减少层数**

```dockerfile
# ❌ 多个RUN层
RUN apk add git
RUN apk add make
RUN apk add gcc

# ✅ 合并为一个层
RUN apk add --no-cache git make gcc
```

**3. 清理缓存**

```dockerfile
# 下载后立即清理
RUN go mod download && \
    go mod verify && \
    rm -rf /go/pkg/mod/cache
```

**4. 使用.dockerignore**

```
# .dockerignore
.git
.gitignore
README.md
Makefile
*.test
*.out
tmp/
vendor/
```

---

## Docker Compose编排

### 多服务编排示例

```yaml
# docker-compose.yml
version: '3.8'

services:
  # Go应用服务
  app:
    build:
      context: .
      dockerfile: Dockerfile
      target: production
    ports:
      - "8080:8080"
    environment:
      - APP_ENV=production
      - DB_HOST=db
      - REDIS_HOST=redis
    depends_on:
      - db
      - redis
    networks:
      - backend
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "-q", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 3s
      retries: 3

  # PostgreSQL数据库
  db:
    image: postgres:15-alpine
    environment:
      - POSTGRES_USER=myapp
      - POSTGRES_PASSWORD=secret
      - POSTGRES_DB=myappdb
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - backend
    restart: unless-stopped

  # Redis缓存
  redis:
    image: redis:7-alpine
    command: redis-server --appendonly yes
    volumes:
      - redis_data:/data
    networks:
      - backend
    restart: unless-stopped

volumes:
  postgres_data:
  redis_data:

networks:
  backend:
    driver: bridge
```

### 开发环境配置

```yaml
# docker-compose.dev.yml
version: '3.8'

services:
  app:
    build:
      context: .
      dockerfile: Dockerfile.dev
    volumes:
      - .:/app
      - /app/vendor  # 排除vendor目录
    environment:
      - APP_ENV=development
      - DEBUG=true
    ports:
      - "8080:8080"
      - "4000:4000"  # Delve调试端口
    command: ["go", "run", "cmd/myapp/main.go"]
```

---

## Docker网络管理

### 网络模式

```bash
# Bridge模式（默认）
docker network create mynet
docker run --network mynet myapp

# Host模式（性能最好）
docker run --network host myapp

# None模式（无网络）
docker run --network none myapp
```

### 自定义网络

```yaml
# docker-compose.yml
networks:
  frontend:
    driver: bridge
    ipam:
      config:
        - subnet: 172.20.0.0/16

  backend:
    driver: bridge
    internal: true  # 内部网络，无外部访问
```

---

## 容器安全实践

### 安全配置清单

```dockerfile
# 1. 使用非root用户
RUN adduser -D -g '' appuser
USER appuser

# 2. 限制能力
# docker run --cap-drop=ALL --cap-add=NET_BIND_SERVICE myapp

# 3. 只读文件系统
# docker run --read-only myapp

# 4. 禁用特权模式
# docker run --security-opt=no-new-privileges myapp

# 5. 资源限制
# docker run --memory="256m" --cpus="0.5" myapp
```

### 安全扫描

```bash
# 使用Trivy扫描镜像
trivy image myapp:latest

# 使用Docker Scout
docker scout quickview myapp:latest

# Snyk扫描
snyk container test myapp:latest
```

---

## 最佳实践总结

### Dockerfile最佳实践

```
[ ] 使用多阶段构建减小体积
[ ] 利用构建缓存（分层COPY）
[ ] 使用Alpine或Distroless基础镜像
[ ] 创建非root用户运行
[ ] 添加健康检查
[ ] 使用.dockerignore排除无关文件
[ ] 合并RUN命令减少层数
[ ] 清理不必要的文件和缓存
[ ] 设置合理的资源限制
[ ] 使用明确的版本标签
```

### 生产环境推荐配置

```dockerfile
FROM golang:1.21-alpine AS builder
# ... 构建过程

FROM gcr.io/distroless/static-debian12

COPY --from=builder /build/app /

# 最小化镜像（<10MB）
# 无shell、无包管理器
# 最大安全性

ENTRYPOINT ["./app"]
```

### 性能对比数据

| 镜像类型 | 体积 | 构建时间 | 安全性 | 适用场景 |
|---------|------|----------|--------|----------|
| golang:1.21 | 800MB | 快 | 低 | 开发环境 |
| alpine | 15MB | 中 | 高 | 生产环境 |
| distroless | 8MB | 中 | 最高 | 高安全要求 |

---

## 实战案例

### Go应用完整示例

```dockerfile
# Dockerfile
# === 构建阶段 ===
FROM golang:1.21-alpine AS builder

WORKDIR /build

# 缓存依赖
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# 复制源代码
COPY . .

# 构建二进制
RUN CGO_ENABLED=0 GOOS=linux \
    go build -ldflags="-s -w" \
    -trimpath -o /build/server ./cmd/server

# === 最终镜像 ===
FROM alpine:3.18

# 时区和证书
RUN apk --no-cache add ca-certificates tzdata

# 非root用户
RUN adduser -D -g '' appuser

WORKDIR /app

# 复制文件
COPY --from=builder /build/server .
COPY --from=builder /build/configs ./configs

USER appuser

HEALTHCHECK --interval=30s --timeout=3s \
    CMD wget -q --spider http://localhost:8080/health || exit 1

EXPOSE 8080

ENTRYPOINT ["./server"]
CMD ["serve"]
```

```yaml
# docker-compose.yml
version: '3.8'

services:
  server:
    build:
      context: .
      dockerfile: Dockerfile
    image: myapp:latest
    ports:
      - "8080:8080"
    environment:
      - LOG_LEVEL=info
      - DB_URL=postgres://user:pass@db:5432/mydb
    depends_on:
      - db
    networks:
      - backend
    restart: unless-stopped
    deploy:
      resources:
        limits:
          memory: 256M
          cpus: '0.5'

  db:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: user
      POSTGRES_PASSWORD: pass
      POSTGRES_DB: mydb
    volumes:
      - db_data:/var/lib/postgresql/data
    networks:
      - backend

volumes:
  db_data:

networks:
  backend:
```

---

## 学习检查点

完成本章节后，验证你的掌握程度：

- [ ] 编写多阶段Dockerfile
- [ ] 镜像体积优化至<50MB
- [ ] 使用Docker Compose编排3个服务
- [ ] 配置健康检查和资源限制
- [ ] 执行镜像安全扫描
- [ ] 理解网络模式和存储管理
- [ ] 配置开发环境热重载

---

## 延伸阅读

- [Docker官方最佳实践](https://docs.docker.com/develop/develop-images/dockerfile_best-practices/)
- [Distroless镜像项目](https://github.com/GoogleContainerTools/distroless)
- [Docker安全指南](https://docs.docker.com/engine/security/)
- [多阶段构建详解](https://docs.docker.com/build/building/multi-stage/)
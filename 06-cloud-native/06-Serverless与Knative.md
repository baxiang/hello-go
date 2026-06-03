# 9.6 Serverless与Knative

## Serverless概念

### 什么是Serverless？

Serverless是一种云原生开发模型：
- 无需管理服务器
- 自动扩缩容（缩容到0）
- 按需计费
- 事件驱动

### Knative组件

```
Knative
├── Serving（服务运行时）
│   ├── Service
│   ├── Route
│   ├── Configuration
│   └── Revision
├── Eventing（事件处理）
│   ├── Broker
│   ├── Trigger
│   └── Source
└── Build（构建，已废弃）
```

---

## Knative Serving实战

### 安装Knative

```bash
# 安装Knative CRD
kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.12.0/serving-crds.yaml

# 安装Knative核心组件
kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.12.0/serving-core.yaml

# 安装网络层（Istio）
kubectl apply -f https://github.com/knative/net-istio/releases/download/knative-v1.12.0/release.yaml

# 验证安装
kubectl get pods -n knative-serving
```

### Knative Service部署

```yaml
# kservice.yaml
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: myapp
  namespace: default
spec:
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/target: "10"        # 并发目标
        autoscaling.knative.dev/minScale: "0"       # 最小实例数
        autoscaling.knative.dev/maxScale: "10"      # 最大实例数
    spec:
      containers:
        - image: myapp:v1.0.0
          ports:
            - containerPort: 8080
          env:
            - name: APP_ENV
              value: "production"
          resources:
            limits:
              memory: "256Mi"
              cpu: "500m"
            requests:
              memory: "128Mi"
              cpu: "250m"
```

### 部署命令

```bash
# 部署服务
kubectl apply -f kservice.yaml

# 查看服务
kubectl get ksvc

# 查看路由
kubectl get route

# 访问服务
kubectl get ksvc myapp -o jsonpath='{.status.url}'
```

---

## 自动扩缩容配置

### KPA (Knative Pod Autoscaler)

```yaml
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: myapp
spec:
  template:
    metadata:
      annotations:
        # 并发配置
        autoscaling.knative.dev/target: "10"         # 目标并发
        autoscaling.knative.dev/targetUtilizationPercentage: "70"
        
        # 实例数限制
        autoscaling.knative.dev/minScale: "0"        # 缩容到0
        autoscaling.knative.dev/maxScale: "100"
        
        # 扩缩容窗口
        autoscaling.knative.dev/window: "60s"
        
        # 稳定窗口
        autoscaling.knative.dev/scaleDownDelay: "30s"
    spec:
      containers:
        - image: myapp:v1.0.0
```

### 流量分发

```yaml
# 流量分割
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: myapp
spec:
  traffic:
    - percent: 90
      tag: current
      revisionName: myapp-00001
    - percent: 10
      tag: canary
      revisionName: myapp-00002
    - percent: 0
      tag: latest
      latestRevision: true
```

---

## Knative Eventing事件处理

### Broker和Trigger

```yaml
# broker.yaml
apiVersion: eventing.knative.dev/v1
kind: Broker
metadata:
  name: default
  namespace: default
---
# trigger.yaml
apiVersion: eventing.knative.dev/v1
kind: Trigger
metadata:
  name: myapp-trigger
  namespace: default
spec:
  broker: default
  filter:
    attributes:
      type: com.example.myevent
  subscriber:
    ref:
      apiVersion: serving.knative.dev/v1
      kind: Service
      name: myapp
```

### 事件源配置

```yaml
# PingSource（定时事件）
apiVersion: sources.knative.dev/v1
kind: PingSource
metadata:
  name: myapp-cron
spec:
  schedule: "*/5 * * * *"  # 每5分钟
  sink:
    ref:
      apiVersion: serving.knative.dev/v1
      kind: Service
      name: myapp

---
# ApiServerSource（K8s事件）
apiVersion: sources.knative.dev/v1
kind: ApiServerSource
metadata:
  name: k8s-events
spec:
  mode: Resource
  resources:
    - apiVersion: v1
      kind: Event
  sink:
    ref:
      apiVersion: serving.knative.dev/v1
      kind: Service
      name: event-handler
```

---

## Go Serverless应用示例

### 示例代码

```go
// main.go
package main

import (
    "context"
    "fmt"
    "net/http"
    "os"
    
    "github.com/cloudevents/sdk-go/v2/event"
    "knative.dev/pkg/signals"
)

func main() {
    ctx := signals.NewContext()
    
    // HTTP服务
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Hello from Serverless Go!")
    })
    
    // 健康检查
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    })
    
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    
    fmt.Printf("Server listening on port %s\n", port)
    if err := http.ListenAndServe(":"+port, nil); err != nil {
        log.Fatal(err)
    }
}
```

### Dockerfile

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o server main.go

FROM alpine:3.18
COPY --from=builder /app/server /server
ENV PORT=8080
EXPOSE 8080
ENTRYPOINT ["/server"]
```

---

## 性能优化

### 冷启动优化

```yaml
# 使用容器预热
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: myapp
spec:
  template:
    metadata:
      annotations:
        # 保持最小实例
        autoscaling.knative.dev/minScale: "1"
        
        # 优先级
        scheduler.alpha.kubernetes.io/critical-pod: ""
    spec:
      containers:
        - image: myapp:v1.0.0
          # 减少镜像体积
          # 快速健康检查
          readinessProbe:
            httpGet:
              path: /health
            initialDelaySeconds: 0
            periodSeconds: 1
```

### 资源配置优化

```yaml
spec:
  containers:
    - image: myapp:v1.0.0
      resources:
        requests:
          memory: "128Mi"
          cpu: "100m"
        limits:
          memory: "256Mi"
          cpu: "500m"
      env:
        - name: GOMAXPROCS
          valueFrom:
            resourceFieldRef:
              resource: limits.cpu
```

---

## 监控与调试

### 查看日志

```bash
# 查看服务日志
kubectl logs -l serving.knative.dev/service=myapp -c user-container

# 实时日志
kubectl logs -f -l serving.knative.dev/service=myapp -c user-container
```

### 性能监控

```bash
# 查看Revision状态
kubectl get revisions

# 查看自动扩缩容状态
kubectl get pods -l serving.knative.dev/service=myapp

# 查看指标
kubectl port-forward -n knative-monitoring svc/grafana 3000:3000
```

---

## 最佳实践

```
[ ] 最小化镜像体积（< 50MB）
[ ] 配置健康检查端点
[ ] 设置合理的资源限制
[ ] 使用环境变量配置
[ ] 实现优雅关闭
[ ] 结构化日志输出
[ ] 错误处理和重试
[ ] 测试冷启动性能
[ ] 监控和告警配置
[ ] 文档化事件格式
```

---

## 学习检查点

- [ ] 安装Knative Serving和Eventing
- [ ] 部署Knative Service
- [ ] 配置自动扩缩容
- [ ] 实现流量分割
- [ ] 配置事件源和Trigger
- [ ] 优化冷启动性能

---

## 延伸阅读

- [Knative官方文档](https://knative.dev/docs/)
- [Serverless框架对比](https://knative.dev/docs/comparison/)
- [CloudEvents规范](https://cloudevents.io/)
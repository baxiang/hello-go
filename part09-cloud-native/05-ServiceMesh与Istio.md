# 9.5 Service Mesh与Istio

## Service Mesh概念

### 什么是Service Mesh？

Service Mesh是处理服务间通信的基础设施层，解决微服务架构中的：
- 服务发现
- 负载均衡
- 故障恢复
- 指标收集
- 分布式追踪
- 安全通信

### Istio架构

```
┌─────────────────────────────────────┐
│           Control Plane              │
│  ┌──────────┐  ┌──────────┐        │
│  │  Pilot   │  │  Mixer   │        │
│  │ (流量管理)│  │ (策略检查)│        │
│  └──────────┘  └──────────┘        │
│  ┌──────────┐  ┌──────────┐        │
│  │ Citadel  │  │ Galley   │        │
│  │ (证书管理)│  │ (配置验证)│        │
│  └──────────┘  └──────────┘        │
└─────────────────────────────────────┘
              ↓ 推送配置
┌─────────────────────────────────────┐
│           Data Plane                │
│  ┌──────────┐  ┌──────────┐        │
│  │ Sidecar  │  │ Sidecar  │        │
│  │  Envoy   │  │  Envoy   │        │
│  │  ┌────┐  │  │  ┌────┐  │        │
│  │  │App │  │  │  │App │  │        │
│  │  └────┘  │  │  └────┘  │        │
│  └──────────┘  └──────────┘        │
└─────────────────────────────────────┘
```

---

## Istio安装与配置

### 安装Istio

```bash
# 下载Istio
curl -L https://istio.io/downloadIstio | sh -
cd istio-1.20.0
export PATH=$PWD/bin:$PATH

# 安装Istio
istioctl install --set profile=demo -y

# 验证安装
kubectl get pods -n istio-system

# 启用自动注入
kubectl label namespace default istio-injection=enabled
```

### 部署示例应用

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
spec:
  replicas: 2
  selector:
    matchLabels:
      app: myapp
  template:
    metadata:
      labels:
        app: myapp
        version: v1
    spec:
      containers:
      - name: app
        image: myapp:v1
        ports:
        - containerPort: 8080
---
apiVersion: v1
kind: Service
metadata:
  name: myapp
spec:
  ports:
  - port: 80
    targetPort: 8080
  selector:
    app: myapp
```

---

## 流量管理

### VirtualService路由规则

```yaml
# virtualservice.yaml
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: myapp
spec:
  hosts:
    - myapp
  http:
    - match:
        - headers:
            version:
              exact: v2
      route:
        - destination:
            host: myapp
            subset: v2
    - route:
        - destination:
            host: myapp
            subset: v1
          weight: 90
        - destination:
            host: myapp
            subset: v2
          weight: 10
```

### DestinationRule目标规则

```yaml
# destinationrule.yaml
apiVersion: networking.istio.io/v1beta1
kind: DestinationRule
metadata:
  name: myapp
spec:
  host: myapp
  trafficPolicy:
    connectionPool:
      tcp:
        maxConnections: 100
      http:
        h2UpgradePolicy: UPGRADE
        http1MaxPendingRequests: 100
        http2MaxRequests: 1000
    outlierDetection:
      consecutive5xxErrors: 5
      interval: 30s
      baseEjectionTime: 30s
      maxEjectionPercent: 50
  subsets:
    - name: v1
      labels:
        version: v1
    - name: v2
      labels:
        version: v2
```

### 金丝雀发布

```yaml
# 90%流量到v1, 10%到v2
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: myapp-canary
spec:
  hosts:
    - myapp
  http:
    - route:
        - destination:
            host: myapp
            subset: v1
          weight: 90
        - destination:
            host: myapp
            subset: v2
          weight: 10
```

---

## 故障注入与测试

### 注入延迟

```yaml
# fault-injection-delay.yaml
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: myapp
spec:
  hosts:
    - myapp
  http:
    - fault:
        delay:
          percentage:
            value: 100
          fixedDelay: 7s
      route:
        - destination:
            host: myapp
```

### 注入中断

```yaml
# fault-injection-abort.yaml
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: myapp
spec:
  hosts:
    - myapp
  http:
    - fault:
        abort:
          percentage:
            value: 50
          httpStatus: 500
      route:
        - destination:
            host: myapp
```

---

## 可观测性集成

### Kiali可视化

```bash
# 安装Kiali
kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.20/samples/addons/kiali.yaml

# 访问Kiali
istioctl dashboard kiali
```

### Jaeger分布式追踪

```bash
# 安装Jaeger
kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.20/samples/addons/jaeger.yaml

# 访问Jaeger
istioctl dashboard jaeger
```

### Prometheus指标收集

```yaml
# 自定义指标
apiVersion: telemetry.istio.io/v1alpha1
kind: Telemetry
metadata:
  name: myapp-telemetry
spec:
  metrics:
    - providers:
        - name: prometheus
      overrides:
        - match:
            metric: REQUEST_COUNT
          tagOverrides:
            request_method:
              value: "request.method"
            request_path:
              value: "request.path"
```

---

## 安全配置

### mTLS双向认证

```yaml
# PeerAuthentication - 严格mTLS
apiVersion: security.istio.io/v1beta1
kind: PeerAuthentication
metadata:
  name: default
  namespace: istio-system
spec:
  mtls:
    mode: STRICT
```

### AuthorizationPolicy授权策略

```yaml
# 允许特定服务访问
apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
  name: myapp-policy
  namespace: production
spec:
  selector:
    matchLabels:
      app: myapp
  action: ALLOW
  rules:
    - from:
        - source:
            principals: ["cluster.local/ns/frontend/sa/frontend"]
      to:
        - operation:
            methods: ["GET", "POST"]
            paths: ["/api/*"]
```

### 速率限制

```yaml
apiVersion: networking.istio.io/v1alpha3
kind: EnvoyFilter
metadata:
  name: ratelimit
spec:
  workloadSelector:
    labels:
      app: myapp
  configPatches:
    - applyTo: HTTP_FILTER
      match:
        context: SIDECAR_INBOUND
      patch:
        operation: INSERT_BEFORE
        value:
          name: envoy.filters.http.ratelimit
          typed_config:
            "@type": type.googleapis.com/envoy.extensions.filters.http.ratelimit.v3.RateLimit
            domain: myapp-ratelimit
            rate_limit_service:
              grpc_service:
                envoy_grpc:
                  cluster_name: rate_limit_cluster
```

---

## 常用命令

```bash
# 检查代理状态
istioctl proxy-status

# 查看代理配置
istioctl proxy-config cluster myapp-pod

# 分析配置
istioctl analyze

# 查看指标
istioctl dashboard prometheus

# 调试
istioctl describe pod myapp-xxx
```

---

## 最佳实践

```
[ ] 启用自动Sidecar注入
[ ] 配置资源限制
[ ] 使用严格mTLS
[ ] 配置授权策略
[ ] 实施金丝雀发布
[ ] 监控指标收集
[ ] 分布式追踪集成
[ ] 定期备份配置
[ ] 测试故障注入
[ ] 文档化流量规则
```

---

## 学习检查点

- [ ] 安装Istio并理解架构
- [ ] 配置VirtualService和DestinationRule
- [ ] 实现金丝雀发布
- [ ] 注入故障测试
- [ ] 集成Kiali/Jaeger监控
- [ ] 配置mTLS和授权策略

---

## 延伸阅读

- [Istio官方文档](https://istio.io/latest/docs/)
- [Envoy代理](https://www.envoyproxy.io/)
- [Service Mesh模式](https://philcalcado.com/2017/08/03/pattern_service_mesh.html)
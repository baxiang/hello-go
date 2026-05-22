# 9.2 Kubernetes基础实战

## 核心概念解析

### Pod（最小部署单元）

```yaml
# pod.yaml
apiVersion: v1
kind: Pod
metadata:
  name: myapp-pod
  labels:
    app: myapp
    version: v1.0
spec:
  containers:
  - name: app
    image: myapp:latest
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
    livenessProbe:
      httpGet:
        path: /health
        port: 8080
      initialDelaySeconds: 30
      periodSeconds: 10
    readinessProbe:
      httpGet:
        path: /ready
        port: 8080
      initialDelaySeconds: 5
      periodSeconds: 5
```

### Deployment（声明式部署）

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp-deployment
spec:
  replicas: 3  # 副本数
  selector:
    matchLabels:
      app: myapp
  strategy:
    type: RollingUpdate  # 滚动更新
    rollingUpdate:
      maxSurge: 1         # 最多额外1个Pod
      maxUnavailable: 0   # 不可用Pod数为0
  template:
    metadata:
      labels:
        app: myapp
    spec:
      containers:
      - name: app
        image: myapp:v1.0
        ports:
        - containerPort: 8080
        resources:
          limits:
            memory: "256Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
```

### Service（服务发现）

```yaml
# service.yaml
apiVersion: v1
kind: Service
metadata:
  name: myapp-service
spec:
  type: ClusterIP  # 内部服务
  selector:
    app: myapp
  ports:
  - port: 80        # Service端口
    targetPort: 8080 # Pod端口
    protocol: TCP

---
# NodePort Service（外部访问）
apiVersion: v1
kind: Service
metadata:
  name: myapp-nodeport
spec:
  type: NodePort
  selector:
    app: myapp
  ports:
  - port: 80
    targetPort: 8080
    nodePort: 30080  # 外部端口(30000-32767)

---
# LoadBalancer Service（云平台）
apiVersion: v1
kind: Service
metadata:
  name: myapp-lb
spec:
  type: LoadBalancer
  selector:
    app: myapp
  ports:
  - port: 80
    targetPort: 8080
```

---

## Go应用部署实战

### 完整部署配置

```yaml
# k8s-full-deployment.yaml
# ConfigMap配置
apiVersion: v1
kind: ConfigMap
metadata:
  name: myapp-config
data:
  APP_ENV: "production"
  LOG_LEVEL: "info"
  DB_HOST: "postgres-service"

---
# Secret密钥
apiVersion: v1
kind: Secret
metadata:
  name: myapp-secret
type: Opaque
data:
  DB_PASSWORD: c2VjcmV0  # base64编码
  JWT_SECRET: amV0LXNlY3JldC1rZXk=

---
# Deployment
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
spec:
  replicas: 3
  selector:
    matchLabels:
      app: myapp
  template:
    metadata:
      labels:
        app: myapp
    spec:
      containers:
      - name: app
        image: myregistry/myapp:v1.0
        ports:
        - containerPort: 8080
        envFrom:
        - configMapRef:
            name: myapp-config
        - secretRef:
            name: myapp-secret
        resources:
          limits:
            memory: "256Mi"
            cpu: "500m"
          requests:
            memory: "128Mi"
            cpu: "250m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5

---
# Service
apiVersion: v1
kind: Service
metadata:
  name: myapp-service
spec:
  type: ClusterIP
  selector:
    app: myapp
  ports:
  - port: 80
    targetPort: 8080
```

### 部署命令

```bash
# 应用配置
kubectl apply -f k8s-full-deployment.yaml

# 查看部署状态
kubectl get deployments
kubectl get pods
kubectl get services

# 查看Pod详情
kubectl describe pod myapp-xxx

# 查看日志
kubectl logs myapp-xxx
kubectl logs -f myapp-xxx  # 实时日志

# 进入容器
kubectl exec -it myapp-xxx -- sh

# 端口转发（本地调试）
kubectl port-forward pod/myapp-xxx 8080:8080

# 扩缩容
kubectl scale deployment myapp --replicas=5

# 更新镜像
kubectl set image deployment/myapp app=myapp:v2.0

# 回滚
kubectl rollout undo deployment/myapp
kubectl rollout history deployment/myapp
```

---

## 健康检查与自动扩缩容

### 健康检查配置

```yaml
# 探针配置详解
livenessProbe:  # 存活探针（失败重启Pod）
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 30  # 启动后延迟
  periodSeconds: 10        # 检查间隔
  timeoutSeconds: 3        # 超时时间
  failureThreshold: 3      # 失败阈值

readinessProbe:  # 就绪探针（失败停止流量）
  httpGet:
    path: /ready
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 5
  successThreshold: 1      # 成功阈值
  failureThreshold: 3

startupProbe:  # 启动探针（慢启动应用）
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 0
  periodSeconds: 10
  failureThreshold: 30     # 最多等待300秒
```

### HPA自动扩缩容

```yaml
# hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: myapp-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: myapp
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70  # CPU使用率70%
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80  # 内存使用率80%
  behavior:  # 扩缩容行为控制
    scaleDown:
      stabilizationWindowSeconds: 300  # 缩容稳定期
      policies:
      - type: Percent
        value: 10
        periodSeconds: 60
    scaleUp:
      stabilizationWindowSeconds: 0
      policies:
      - type: Percent
        value: 100
        periodSeconds: 15
```

---

## 配置管理

### ConfigMap使用

```yaml
# 创建ConfigMap
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  # 单个键值
  APP_NAME: "myapp"
  
  # 配置文件
  config.yaml: |
    server:
      port: 8080
    database:
      host: postgres
      port: 5432

---
# 使用ConfigMap（环境变量）
spec:
  containers:
  - name: app
    envFrom:
    - configMapRef:
        name: app-config
    
    # 或单个变量
    env:
    - name: APP_NAME
      valueFrom:
        configMapKeyRef:
          name: app-config
          key: APP_NAME

---
# 使用ConfigMap（挂载文件）
spec:
  containers:
  - name: app
    volumeMounts:
    - name: config-volume
      mountPath: /etc/config
  volumes:
  - name: config-volume
    configMap:
      name: app-config
```

### Secret管理

```bash
# 创建Secret
kubectl create secret generic app-secret \
  --from-literal=username=admin \
  --from-literal=password=secret

# 或从文件
kubectl create secret generic app-secret \
  --from-file=username.txt \
  --from-file=password.txt

# 加密存储（启用加密）
kubectl create secret tls tls-secret \
  --cert=path/to/tls.crt \
  --key=path/to/tls.key
```

---

## 资源管理

### 资源配额

```yaml
# namespace资源配额
apiVersion: v1
kind: ResourceQuota
metadata:
  name: compute-quota
  namespace: production
spec:
  hard:
    requests.cpu: "4"
    requests.memory: "8Gi"
    limits.cpu: "8"
    limits.memory: "16Gi"
    pods: "10"
    services: "5"
```

### LimitRange（默认限制）

```yaml
apiVersion: v1
kind: LimitRange
metadata:
  name: default-limits
spec:
  limits:
  - default:      # 默认limits
      cpu: "500m"
      memory: "512Mi"
    defaultRequest:  # 默认requests
      cpu: "250m"
      memory: "256Mi"
    type: Container
```

---

## 网络策略

```yaml
# NetworkPolicy（网络隔离）
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: app-network-policy
spec:
  podSelector:
    matchLabels:
      app: myapp
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - podSelector:
        matchLabels:
          role: frontend
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to:
    - podSelector:
        matchLabels:
          role: database
    ports:
    - protocol: TCP
      port: 5432
```

---

## 常用命令速查

### 集群管理

```bash
# 集群信息
kubectl cluster-info
kubectl get nodes
kubectl describe node node-1

# 节点标签
kubectl label node node-1 disk=ssd
kubectl get nodes --show-labels

# 资源查看
kubectl get all
kubectl get pods -o wide
kubectl get deployments -n production

# 清理资源
kubectl delete pod myapp-xxx
kubectl delete deployment myapp
kubectl delete all -l app=myapp
```

### 故障排查

```bash
# Pod状态检查
kubectl describe pod myapp-xxx
kubectl logs myapp-xxx
kubectl logs myapp-xxx --previous  # 上一个容器日志

# 事件查看
kubectl get events
kubectl describe pod myapp-xxx | grep -A 10 Events

# 调试
kubectl exec -it myapp-xxx -- /bin/sh
kubectl run debug --image=busybox --rm -it --restart=Never -- sh

# 资源使用
kubectl top pods
kubectl top nodes
```

---

## 最佳实践

### 生产环境配置清单

```
[ ] 使用Deployment而非Pod（副本管理）
[ ] 配置资源限制（limits/requests）
[ ] 实现健康检查（liveness/readiness）
[ ] 使用ConfigMap/Secret管理配置
[ ] 配置HPA自动扩缩容
[ ] 使用Namespace资源隔离
[ ] 配置NetworkPolicy网络策略
[ ] 使用PersistentVolume持久化存储
[ ] 配置Service服务发现
[ ] 使用Ingress外部访问管理
```

---

## 学习检查点

完成本章节后，验证：

- [ ] 理解Pod、Deployment、Service概念
- [ ] 成功部署Go应用到K8s
- [ ] 配置健康检查和自动扩缩容
- [ ] 使用ConfigMap/Secret管理配置
- [ ] 实现服务发现和负载均衡
- [ ] 完成滚动更新和回滚操作
- [ ] 掌握故障排查命令

---

## 延伸阅读

- [Kubernetes官方文档](https://kubernetes.io/docs/)
- [K8s概念详解](https://kubernetes.io/docs/concepts/)
- [K8s最佳实践](https://kubernetes.io/docs/concepts/configuration/overview/)
- [K8s交互式教程](https://kubernetes.io/docs/tutorials/)
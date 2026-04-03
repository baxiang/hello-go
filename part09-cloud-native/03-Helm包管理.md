# 9.3 Helm包管理

## Helm核心概念

### 什么是Helm？

Helm是Kubernetes的包管理器，类似于：
- Ubuntu的apt
- CentOS的yum
- macOS的brew
- Node.js的npm

**核心优势**:
- 打包应用为Chart（可复用）
- 版本管理和回滚
- 配置参数化（多环境部署）
- 依赖管理

### 三大概念

**1. Chart**: 应用包（模板+默认配置）
**2. Release**: Chart的部署实例
**3. Repository**: Chart仓库

---

## Chart结构详解

### 标准结构

```
mychart/
├── Chart.yaml          # Chart元数据
├── values.yaml         # 默认配置值
├── charts/             # 依赖Chart
├── templates/          # K8s模板文件
│   ├── deployment.yaml
│   ├── service.yaml
│   ├── configmap.yaml
│   ├── ingress.yaml
│   └── NOTES.txt      # 安装后说明
├── .helmignore         # 打包忽略文件
└── README.md           # Chart说明
```

### Chart.yaml详解

```yaml
# Chart.yaml
apiVersion: v2
name: myapp
description: A Helm chart for my Go application
type: application
version: 1.0.0        # Chart版本
appVersion: "2.0.0"   # 应用版本

keywords:
  - go
  - web
  - microservice

maintainers:
  - name: Your Name
    email: your@email.com

dependencies:  # 依赖
  - name: postgresql
    version: "12.x.x"
    repository: "https://charts.bitnami.com/bitnami"
    condition: postgresql.enabled
```

---

## 编写自定义Chart

### 创建Chart

```bash
# 创建新Chart
helm create myapp

# 验证语法
helm lint myapp

# 模板渲染测试
helm template myapp ./mychart

# 安装测试
helm install --dry-run --debug myapp ./mychart
```

### Deployment模板示例

```yaml
# templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "myapp.fullname" . }}
  labels:
    {{- include "myapp.labels" . | nindent 4 }}
spec:
  replicas: {{ .Values.replicaCount }}
  selector:
    matchLabels:
      {{- include "myapp.selectorLabels" . | nindent 6 }}
  template:
    metadata:
      labels:
        {{- include "myapp.selectorLabels" . | nindent 8 }}
    spec:
      containers:
        - name: {{ .Chart.Name }}
          image: "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
          imagePullPolicy: {{ .Values.image.pullPolicy }}
          ports:
            - name: http
              containerPort: {{ .Values.service.port }}
              protocol: TCP
          {{- if .Values.livenessProbe.enabled }}
          livenessProbe:
            httpGet:
              path: {{ .Values.livenessProbe.path }}
              port: http
            initialDelaySeconds: {{ .Values.livenessProbe.initialDelaySeconds }}
            periodSeconds: {{ .Values.livenessProbe.periodSeconds }}
          {{- end }}
          {{- if .Values.readinessProbe.enabled }}
          readinessProbe:
            httpGet:
              path: {{ .Values.readinessProbe.path }}
              port: http
            initialDelaySeconds: {{ .Values.readinessProbe.initialDelaySeconds }}
            periodSeconds: {{ .Values.readinessProbe.periodSeconds }}
          {{- end }}
          resources:
            {{- toYaml .Values.resources | nindent 12 }}
          env:
            {{- range $key, $value := .Values.env }}
            - name: {{ $key }}
              value: {{ $value | quote }}
            {{- end }}
```

### values.yaml配置

```yaml
# values.yaml
replicaCount: 3

image:
  repository: myregistry/myapp
  tag: "v1.0.0"
  pullPolicy: IfNotPresent

nameOverride: ""
fullnameOverride: ""

service:
  type: ClusterIP
  port: 80

resources:
  limits:
    cpu: 500m
    memory: 512Mi
  requests:
    cpu: 250m
    memory: 256Mi

livenessProbe:
  enabled: true
  path: /health
  initialDelaySeconds: 30
  periodSeconds: 10

readinessProbe:
  enabled: true
  path: /ready
  initialDelaySeconds: 5
  periodSeconds: 5

env:
  APP_NAME: "myapp"
  LOG_LEVEL: "info"

ingress:
  enabled: false
  hosts:
    - host: myapp.example.com
      paths: ["/"]

autoscaling:
  enabled: false
  minReplicas: 2
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70
```

---

## 多环境配置管理

### 目录结构

```
mychart/
├── values.yaml              # 默认配置
├── values-dev.yaml          # 开发环境
├── values-staging.yaml      # 测试环境
├── values-prod.yaml         # 生产环境
└── templates/
```

### 环境配置示例

```yaml
# values-dev.yaml
replicaCount: 1

image:
  tag: "latest"

resources:
  limits:
    cpu: 200m
    memory: 256Mi

env:
  APP_ENV: "development"
  LOG_LEVEL: "debug"

# values-staging.yaml
replicaCount: 2

image:
  tag: "v1.0.0-rc1"

resources:
  limits:
    cpu: 300m
    memory: 384Mi

env:
  APP_ENV: "staging"
  LOG_LEVEL: "info"

# values-prod.yaml
replicaCount: 5

image:
  tag: "v1.0.0"

resources:
  limits:
    cpu: 1000m
    memory: 1Gi
  requests:
    cpu: 500m
    memory: 512Mi

autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10

env:
  APP_ENV: "production"
  LOG_LEVEL: "error"
```

### 部署命令

```bash
# 开发环境
helm upgrade --install myapp ./mychart \
  -f mychart/values-dev.yaml \
  -n development

# 测试环境
helm upgrade --install myapp ./mychart \
  -f mychart/values-staging.yaml \
  -n staging

# 生产环境
helm upgrade --install myapp ./mychart \
  -f mychart/values-prod.yaml \
  -n production \
  --atomic  # 失败自动回滚
```

---

## Chart高级特性

### 条件渲染

```yaml
# templates/ingress.yaml
{{- if .Values.ingress.enabled -}}
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: {{ include "myapp.fullname" . }}
  annotations:
    kubernetes.io/ingress.class: nginx
spec:
  rules:
  {{- range .Values.ingress.hosts }}
    - host: {{ .host }}
      http:
        paths:
        {{- range .paths }}
          - path: {{ . }}
            backend:
              service:
                name: {{ include "myapp.fullname" $ }}
                port:
                  number: {{ $.Values.service.port }}
        {{- end }}
  {{- end }}
{{- end }}
```

### 循环渲染

```yaml
# 遍历ConfigMap
{{- range $key, $value := .Values.config }}
  {{ $key }}: {{ $value | quote }}
{{- end }}

# 遍历列表
{{- range .Values.hosts }}
  - {{ . }}
{{- end }}
```

### 助手函数

```yaml
# templates/_helpers.tpl
{{- define "myapp.labels" -}}
app.kubernetes.io/name: {{ .Chart.Name }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion }}
{{- end }}

{{- define "myapp.selectorLabels" -}}
app.kubernetes.io/name: {{ .Chart.Name }}
{{- end }}

# 使用
metadata:
  labels:
    {{- include "myapp.labels" . | nindent 4 }}
```

---

## Chart依赖管理

### 定义依赖

```yaml
# Chart.yaml
dependencies:
  - name: postgresql
    version: "12.x.x"
    repository: "https://charts.bitnami.com/bitnami"
    condition: postgresql.enabled
  
  - name: redis
    version: "17.x.x"
    repository: "https://charts.bitnami.com/bitnami"
    condition: redis.enabled
```

### 依赖操作

```bash
# 下载依赖
helm dependency update ./mychart

# 查看依赖
helm dependency list ./mychart

# 构建依赖
helm dependency build ./mychart
```

---

## Chart仓库管理

### 创建私有仓库

```bash
# 本地仓库
helm serve &

# GitHub Pages
helm package mychart
helm repo index ./ --url https://username.github.io/charts
git push

# 添加仓库
helm repo add myrepo https://username.github.io/charts
helm repo update
```

### Chart发布流程

```bash
# 1. 打包Chart
helm package mychart

# 2. 生成索引
helm repo index ./ --url https://myrepo.com/charts

# 3. 上传到仓库
# (使用Git、S3、或ChartMuseum)

# 4. 更新仓库
helm repo update

# 5. 搜索安装
helm search repo myapp
helm install myapp myrepo/myapp
```

---

## 常用命令速查

```bash
# Chart管理
helm create mychart          # 创建Chart
helm lint mychart            # 验证语法
helm package mychart         # 打包Chart
helm template myapp mychart  # 渲染模板

# Release管理
helm install myapp mychart          # 安装
helm upgrade myapp mychart          # 升级
helm rollback myapp 1               # 回滚到版本1
helm uninstall myapp                # 卸载
helm list                           # 列出Release
helm status myapp                   # 查看状态

# 历史管理
helm history myapp           # 查看历史
helm get values myapp        # 查看配置值
helm get manifest myapp      # 查看清单

# 仓库管理
helm repo add myrepo https://charts.example.com
helm repo update              # 更新仓库
helm search repo nginx        # 搜索Chart
helm show values myrepo/nginx # 查看默认值
```

---

## 最佳实践

### Chart开发清单

```
[ ] 清晰的README文档
[ ] 完整的values.yaml注释
[ ] 资源限制配置
[ ] 健康检查配置
[ ] 多环境配置文件
[ ] 条件渲染减少冗余
[ ] 使用_helpers.tpl复用模板
[ ] 版本号遵循语义化版本
[ ] Chart测试用例
[ ] NOTICES.txt安装说明
```

### 安全建议

```yaml
# 不要在values.yaml中存储敏感信息
# 使用Secret或外部Secret管理

# 使用镜像摘要而非标签（生产）
image:
  repository: myapp
  digest: sha256:abc123...

# 资源限制必须设置
resources:
  limits: {...}
  requests: {...}

# 网络策略
networkPolicy:
  enabled: true
```

---

## 实战案例：完整Go应用Chart

```bash
# 项目结构
myapp-chart/
├── Chart.yaml
├── values.yaml
├── values-dev.yaml
├── values-staging.yaml
├── values-prod.yaml
├── templates/
│   ├── deployment.yaml
│   ├── service.yaml
│   ├── configmap.yaml
│   ├── ingress.yaml
│   ├── hpa.yaml
│   ├── _helpers.tpl
│   └── NOTES.txt
└── README.md
```

```bash
# 部署流程
# 1. 验证
helm lint ./myapp-chart

# 2. 预览
helm template myapp ./myapp-chart -f values-prod.yaml

# 3. 安装
helm upgrade --install myapp ./myapp-chart \
  -f values-prod.yaml \
  -n production \
  --create-namespace \
  --atomic \
  --timeout 5m

# 4. 验证
kubectl get all -n production
helm status myapp -n production
```

---

## 学习检查点

完成本章节后，验证：

- [ ] 创建自定义Chart
- [ ] 编写Deployment和Service模板
- [ ] 配置多环境values文件
- [ ] 使用条件渲染和循环
- [ ] 管理Chart依赖
- [ ] 发布Chart到仓库
- [ ] 实现版本回滚

---

## 延伸阅读

- [Helm官方文档](https://helm.sh/docs/)
- [Chart模板指南](https://helm.sh/docs/chart_template_guide/)
- [Chart最佳实践](https://helm.sh/docs/chart_best_practices/)
- [Artifact Hub](https://artifacthub.io/) - Chart仓库搜索
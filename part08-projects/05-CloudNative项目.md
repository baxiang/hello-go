# 8.5 Cloud Native 项目

## Kubernetes Operator 开发

### Operator 概念

```
Operator = Custom Resource Definition (CRD) + Controller

Operator 模式:
1. 定义 Custom Resource (自定义资源)
2. 编写 Controller (控制器) 监听资源变化
3. 自动执行运维操作

典型场景:
- 数据库自动备份
- 应用自动扩缩容
- 证书自动续期
- 配置自动同步
```

### Kubebuilder 入门

```bash
# 安装 kubebuilder
curl -L -o kubebuilder https://github.com/kubernetes-sigs/kubebuilder/releases/download/v3.11.1/kubebuilder_$(go env GOOS)_$(go env GOARCH)
chmod +x kubebuilder
mv kubebuilder /usr/local/bin/

# 创建项目
mkdir my-operator
cd my-operator
kubebuilder init --domain example.com --repo github.com/example/my-operator

# 创建 API
kubebuilder create api --group apps --version v1 --kind MyApp

# 生成代码
make manifests
make generate
```

### 项目结构

```
my-operator/
├── api/
│   └── apps/v1/
│       ├── myapp_types.go      # CRD 类型定义
│       ├── groupversion_info.go
│       └── zz_generated.deepcopy.go
├── controllers/
│   ├── myapp_controller.go     # Controller 逻辑
│   └── suite_test.go
├── config/
│   ├── crd/                    # CRD 配置
│   ├── default/                # Kustomize 默认配置
│   ├── manager/                # Manager 配置
│   ├── rbac/                   # RBAC 配置
│   └── samples/                # 示例资源
├── main.go
├── Makefile
└── go.mod
```

### 定义 Custom Resource

```go
// api/apps/v1/myapp_types.go

package v1

import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MyAppSpec 定义 MyApp 的期望状态
type MyAppSpec struct {
    // 副本数量
    // +kubebuilder:validation:Minimum=1
    // +kubebuilder:validation:Maximum=100
    Replicas int32 `json:"replicas"`

    // 镜像
    Image string `json:"image"`

    // 端口
    Port int32 `json:"port"`

    // 资源限制
    Resources ResourceRequirements `json:"resources,omitempty"`

    // 环境变量
    Env []EnvVar `json:"env,omitempty"`
}

type ResourceRequirements struct {
    CPU    string `json:"cpu"`
    Memory string `json:"memory"`
}

type EnvVar struct {
    Name  string `json:"name"`
    Value string `json:"value"`
}

// MyAppStatus 定义 MyApp 的实际状态
type MyAppStatus struct {
    // 可用副本数
    AvailableReplicas int32 `json:"availableReplicas"`

    // 当前状态
    Conditions []metav1.Condition `json:"conditions,omitempty"`

    // 当前镜像
    CurrentImage string `json:"currentImage"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Replicas",type=integer,JSONPath=`.spec.replicas`
// +kubebuilder:printcolumn:name="Image",type=string,JSONPath=`.spec.image`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// MyApp 是自定义资源
type MyApp struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`

    Spec   MyAppSpec   `json:"spec,omitempty"`
    Status MyAppStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MyAppList 包含 MyApp 列表
type MyAppList struct {
    metav1.TypeMeta `json:",inline"`
    metav1.ListMeta `json:"metadata,omitempty"`
    Items           []MyApp `json:"items"`
}

func init() {
    SchemeBuilder.Register(&MyApp{}, &MyAppList{})
}
```

### 编写 Controller

```go
// controllers/myapp_controller.go

package controllers

import (
    "context"
    "fmt"

    appsv1 "k8s.io/api/apps/v1"
    corev1 "k8s.io/api/core/v1"
    "k8s.io/apimachinery/pkg/api/errors"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/runtime"
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "sigs.k8s.io/controller-runtime/pkg/log"

    appsv1alpha1 "github.com/example/my-operator/api/apps/v1"
)

// MyAppReconciler 协调器
type MyAppReconciler struct {
    client.Client
    Scheme *runtime.Scheme
}

// Reconcile 协调函数
func (r *MyAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    logger := log.FromContext(ctx)

    // 获取 MyApp 资源
    var myApp appsv1alpha1.MyApp
    if err := r.Get(ctx, req.NamespacedName, &myApp); err != nil {
        if errors.IsNotFound(err) {
            // 资源被删除，无需处理
            return ctrl.Result{}, nil
        }
        return ctrl.Result{}, err
    }

    logger.Info("Reconciling MyApp", "name", myApp.Name)

    // 创建或更新 Deployment
    deployment, err := r.reconcileDeployment(ctx, &myApp)
    if err != nil {
        return ctrl.Result{}, err
    }

    // 创建或更新 Service
    service, err := r.reconcileService(ctx, &myApp)
    if err != nil {
        return ctrl.Result{}, err
    }

    // 更新状态
    if err := r.updateStatus(ctx, &myApp, deployment); err != nil {
        return ctrl.Result{}, err
    }

    return ctrl.Result{}, nil
}

func (r *MyAppReconciler) reconcileDeployment(ctx context.Context, myApp *appsv1alpha1.MyApp) (*appsv1.Deployment, error) {
    // 定义期望的 Deployment
    desired := &appsv1.Deployment{
        ObjectMeta: metav1.ObjectMeta{
            Name:      myApp.Name,
            Namespace: myApp.Namespace,
        },
        Spec: appsv1.DeploymentSpec{
            Replicas: &myApp.Spec.Replicas,
            Selector: &metav1.LabelSelector{
                MatchLabels: map[string]string{
                    "app": myApp.Name,
                },
            },
            Template: corev1.PodTemplateSpec{
                ObjectMeta: metav1.ObjectMeta{
                    Labels: map[string]string{
                        "app": myApp.Name,
                    },
                },
                Spec: corev1.PodSpec{
                    Containers: []corev1.Container{
                        {
                            Name:  myApp.Name,
                            Image: myApp.Spec.Image,
                            Ports: []corev1.ContainerPort{
                                {
                                    ContainerPort: myApp.Spec.Port,
                                },
                            },
                            Resources: corev1.ResourceRequirements{
                                Requests: corev1.ResourceList{
                                    corev1.ResourceCPU:    resource.MustParse(myApp.Spec.Resources.CPU),
                                    corev1.ResourceMemory: resource.MustParse(myApp.Spec.Resources.Memory),
                                },
                            },
                        },
                    },
                },
            },
        },
    }

    // 设置所有者引用
    ctrl.SetControllerReference(myApp, desired, r.Scheme)

    // 检查是否已存在
    var existing appsv1.Deployment
    err := r.Get(ctx, client.ObjectKey{
        Name:      myApp.Name,
        Namespace: myApp.Namespace,
    }, &existing)

    if errors.IsNotFound(err) {
        // 创建 Deployment
        if err := r.Create(ctx, desired); err != nil {
            return nil, err
        }
        return desired, nil
    } else if err != nil {
        return nil, err
    }

    // 更新现有 Deployment
    existing.Spec = desired.Spec
    if err := r.Update(ctx, &existing); err != nil {
        return nil, err
    }

    return &existing, nil
}

func (r *MyAppReconciler) reconcileService(ctx context.Context, myApp *appsv1alpha1.MyApp) (*corev1.Service, error) {
    // 类似 Deployment 的逻辑创建 Service
    // ...
    return nil, nil
}

func (r *MyAppReconciler) updateStatus(ctx context.Context, myApp *appsv1alpha1.MyApp, deployment *appsv1.Deployment) error {
    myApp.Status.AvailableReplicas = deployment.Status.AvailableReplicas
    myApp.Status.CurrentImage = myApp.Spec.Image

    return r.Status().Update(ctx, myApp)
}

// SetupWithManager 注册 Controller
func (r *MyAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&appsv1alpha1.MyApp{}).
        Owns(&appsv1.Deployment{}).
        Owns(&corev1.Service{}).
        Complete(r)
}
```

### 示例资源

```yaml
# config/samples/apps_v1_myapp.yaml

apiVersion: apps.example.com/v1
kind: MyApp
metadata:
  name: myapp-sample
  namespace: default
spec:
  replicas: 3
  image: nginx:latest
  port: 80
  resources:
    cpu: "100m"
    memory: "128Mi"
  env:
    - name: ENV
      value: production
```

### 部署 Operator

```bash
# 本地运行测试
make run

# 构建镜像
make docker-build docker-push IMG=example.com/my-operator:v0.1.0

# 安装 CRD
make install

# 部署 Operator
make deploy IMG=example.com/my-operator:v0.1.0

# 创建示例资源
kubectl apply -f config/samples/apps_v1_myapp.yaml

# 查看资源
kubectl get myapp
kubectl describe myapp myapp-sample

# 卸载
make uninstall
make undeploy
```

---

## Helm Chart 开发

### Chart 结构

```
my-chart/
├── Chart.yaml          # Chart 元数据
├── values.yaml         # 默认配置值
├── values-prod.yaml    # 生产环境配置
├── templates/          # 模板文件
│   ├── _helpers.tpl    # 模板辅助函数
│   ├── deployment.yaml
│   ├── service.yaml
│   ├── configmap.yaml
│   ├── ingress.yaml
│   └── tests/
│       └── test-connection.yaml
├── charts/             # 子 Chart (依赖)
└── crds/               # Custom Resource Definitions
```

### Chart.yaml

```yaml
apiVersion: v2
name: myapp
description: A Helm chart for MyApp application
type: application
version: 0.1.0        # Chart 版本
appVersion: "1.0.0"   # 应用版本

keywords:
  - myapp
  - web

maintainers:
  - name: Your Name
    email: your@email.com

dependencies:
  - name: postgresql
    version: "12.x"
    repository: "https://charts.bitnami.com/bitnami"
    condition: postgresql.enabled

  - name: redis
    version: "17.x"
    repository: "https://charts.bitnami.com/bitnami"
    condition: redis.enabled
```

### values.yaml

```yaml
# 应用配置
replicaCount: 3

image:
  repository: myapp
  tag: "latest"
  pullPolicy: IfNotPresent

imagePullSecrets: []
nameOverride: ""
fullnameOverride: ""

# 服务配置
service:
  type: ClusterIP
  port: 80
  targetPort: 8080

# Ingress 配置
ingress:
  enabled: true
  className: nginx
  annotations:
    kubernetes.io/ingress.class: nginx
  hosts:
    - host: myapp.example.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: myapp-tls
      hosts:
        - myapp.example.com

# 资源限制
resources:
  limits:
    cpu: 500m
    memory: 512Mi
  requests:
    cpu: 100m
    memory: 128Mi

# 自动扩缩容
autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10
  targetCPUUtilizationPercentage: 80
  targetMemoryUtilizationPercentage: 80

# 健康检查
livenessProbe:
  httpGet:
    path: /health
    port: http
  initialDelaySeconds: 30
  periodSeconds: 10

readinessProbe:
  httpGet:
    path: /ready
    port: http
  initialDelaySeconds: 5
  periodSeconds: 5

# 环境变量
env:
  - name: ENV
    value: "production"
  - name: LOG_LEVEL
    value: "info"

# 依赖配置
postgresql:
  enabled: true
  auth:
    username: myapp
    password: secret
    database: myapp

redis:
  enabled: false
```

### templates/deployment.yaml

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "myapp.fullname" . }}
  labels:
    {{- include "myapp.labels" . | nindent 4 }}
spec:
  {{- if not .Values.autoscaling.enabled }}
  replicas: {{ .Values.replicaCount }}
  {{- end }}
  selector:
    matchLabels:
      {{- include "myapp.selectorLabels" . | nindent 6 }}
  template:
    metadata:
      labels:
        {{- include "myapp.selectorLabels" . | nindent 8 }}
      annotations:
        checksum/config: {{ include (print $.Template.BasePath "/configmap.yaml") . | sha256sum }}
    spec:
      {{- with .Values.imagePullSecrets }}
      imagePullSecrets:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      containers:
        - name: {{ .Chart.Name }}
          image: "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
          imagePullPolicy: {{ .Values.image.pullPolicy }}
          ports:
            - name: http
              containerPort: {{ .Values.service.targetPort }}
              protocol: TCP
          {{- with .Values.livenessProbe }}
          livenessProbe:
            {{- toYaml . | nindent 12 }}
          {{- end }}
          {{- with .Values.readinessProbe }}
          readinessProbe:
            {{- toYaml . | nindent 12 }}
          {{- end }}
          resources:
            {{- toYaml .Values.resources | nindent 12 }}
          env:
            {{- range .Values.env }}
            - name: {{ .name }}
              value: {{ .value | quote }}
            {{- end }}
            {{- if .Values.postgresql.enabled }}
            - name: DB_HOST
              value: "{{ .Release.Name }}-postgresql"
            - name: DB_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: {{ .Release.Name }}-postgresql
                  key: password
            {{- end }}
```

### templates/_helpers.tpl

```
{{/*
Expand the name of the chart.
*/}}
{{- define "myapp.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "myapp.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "myapp.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "myapp.labels" -}}
helm.sh/chart: {{ include "myapp.chart" . }}
{{ include "myapp.selectorLabels" . }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "myapp.selectorLabels" -}}
app.kubernetes.io/name: {{ include "myapp.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
```

### Helm 命令

```bash
# 创建 Chart
helm create my-chart

# 语法检查
helm lint ./my-chart

# 渲染模板
helm template my-release ./my-chart
helm template my-release ./my-chart -f values-prod.yaml

# 安装 Chart
helm install my-release ./my-chart
helm install my-release ./my-chart -f values-prod.yaml
helm install my-release ./my-chart --set replicaCount=5

# 升级
helm upgrade my-release ./my-chart -f values-prod.yaml

# 回滚
helm rollback my-release 1

# 查看状态
helm status my-release
helm list

# 卸载
helm uninstall my-release

# 依赖更新
helm dependency update
helm dependency build

# 打包
helm package ./my-chart

# 发布到仓库
helm push my-chart-0.1.0.tgz oci://registry.example.com/charts
```

---

## 云原生检查清单

```
[ ] Operator 实现资源自动化管理
[ ] Helm Chart 支持多环境配置
[ ] 使用 CRD 扩展 Kubernetes API
[ ] 实现健康检查和自动恢复
[ ] 配置合理的资源限制
[ ] 使用 HPA 实现自动扩缩容
[ ] 实现优雅关闭和启动
[ ] 添加监控指标导出
[ ] 配置日志收集和聚合
[ ] 实现配置热重载
```

# 9.4 GitOps持续部署

## GitOps核心理念

### 什么是GitOps？

GitOps是一种现代化的持续交付方式，核心思想：

**Git是唯一的真实来源（Single Source of Truth）**

```
传统部署流程：
代码提交 → CI构建 → CD推送 → 生产环境

GitOps部署流程：
代码提交 → Git仓库更新 → GitOps Operator同步 → 生产环境
```

### 核心原则

1. **声明式**: 基础设施即代码（IaC）
2. **版本化**: Git管理所有配置
3. **自动化**: Git变更自动同步到集群
4. **可审计**: Git历史即变更历史
5. **可回滚**: Git回滚即环境回滚

---

## ArgoCD实战

### 安装配置

```bash
# 安装ArgoCD
kubectl create namespace argocd
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml

# 获取初始密码
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d

# 访问UI
kubectl port-forward svc/argocd-server -n argocd 8080:443

# CLI安装
brew install argocd
argocd login localhost:8080
```

### Application配置

```yaml
# application.yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: myapp
  namespace: argocd
spec:
  project: default
  
  # Git仓库配置
  source:
    repoURL: https://github.com/myorg/myapp-config.git
    targetRevision: main
    path: overlays/production
    
    # Helm配置
    helm:
      valueFiles:
        - values-prod.yaml
      parameters:
        - name: image.tag
          value: v1.2.0
  
  # 目标集群
  destination:
    server: https://kubernetes.default.svc
    namespace: production
  
  # 同步策略
  syncPolicy:
    automated:
      prune: true     # 自动清理
      selfHeal: true  # 自动修复
    syncOptions:
      - CreateNamespace=true
    retry:
      limit: 5
      backoff:
        duration: 5s
        factor: 2
        maxDuration: 3m
```

### 多环境管理

```
# Git仓库结构
myapp-config/
├── bases/              # 基础配置
│   ├── deployment.yaml
│   ├── service.yaml
│   └── kustomization.yaml
├── overlays/           # 环境覆盖
│   ├── development/
│   │   ├── kustomization.yaml
│   │   └── patches/
│   ├── staging/
│   │   ├── kustomization.yaml
│   │   └── patches/
│   └── production/
│       ├── kustomization.yaml
│       └── patches/
└── apps/
    ├── dev.yaml
    ├── staging.yaml
    └── prod.yaml
```

```yaml
# apps/prod.yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: myapp-prod
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/myorg/myapp-config.git
    targetRevision: main
    path: overlays/production
  destination:
    server: https://kubernetes.default.svc
    namespace: production
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
```

---

## FluxCD实战

### 安装配置

```bash
# 安装Flux CLI
brew install fluxcd/tap/flux

# 安装到集群
flux install --namespace=flux-system

# 引导Git仓库
flux bootstrap github \
  --owner=myorg \
  --repository=myapp-config \
  --branch=main \
  --path=./clusters/production \
  --personal
```

### GitRepository配置

```yaml
# gitrepository.yaml
apiVersion: source.toolkit.fluxcd.io/v1beta2
kind: GitRepository
metadata:
  name: myapp
  namespace: flux-system
spec:
  interval: 1m0s
  ref:
    branch: main
  url: https://github.com/myorg/myapp-config
  secretRef:
    name: myapp-ssh-key  # SSH密钥
```

### Kustomization配置

```yaml
# kustomization.yaml
apiVersion: kustomize.toolkit.fluxcd.io/v1beta2
kind: Kustomization
metadata:
  name: myapp
  namespace: flux-system
spec:
  interval: 5m0s
  path: ./overlays/production
  prune: true
  sourceRef:
    kind: GitRepository
    name: myapp
  timeout: 2m0s
  validation: client
  healthChecks:
    - apiVersion: apps/v1
      kind: Deployment
      name: myapp
      namespace: production
```

### HelmRelease配置

```yaml
# helmrelease.yaml
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  name: myapp
  namespace: flux-system
spec:
  interval: 5m
  chart:
    spec:
      chart: mychart
      version: '1.0.x'
      sourceRef:
        kind: HelmRepository
        name: myrepo
      interval: 1m
  values:
    replicaCount: 3
    image:
      repository: myapp
      tag: v1.0.0
    resources:
      limits:
        cpu: 500m
        memory: 512Mi
```

---

## 工作流程对比

### ArgoCD工作流

```
开发者工作流：
1. 修改代码并提交
2. CI构建并推送镜像
3. 更新Git配置仓库（镜像tag）
4. ArgoCD自动检测并同步
5. 生产环境自动更新

回滚流程：
1. Git revert提交
2. ArgoCD自动同步回滚
```

### FluxCD工作流

```
开发者工作流：
1. 修改代码并提交
2. CI构建并推送镜像
3. Flux检测Git变更
4. 自动更新K8s资源
5. 生产环境自动更新

回滚流程：
1. Git revert提交
2. Flux自动同步回滚
```

---

## CI/CD集成

### GitHub Actions集成

```yaml
# .github/workflows/deploy.yaml
name: Deploy to Production

on:
  push:
    branches: [main]

jobs:
  build-and-deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v3
      
      - name: Build Docker image
        run: |
          docker build -t myapp:${{ github.sha }} .
          docker push myapp:${{ github.sha }}
      
      - name: Update GitOps repo
        run: |
          git clone https://github.com/myorg/myapp-config.git
          cd myapp-config
          # 更新镜像tag
          yq e '.image.tag = "${{ github.sha }}"' -i overlays/production/values.yaml
          git commit -am "Update image to ${{ github.sha }}"
          git push
```

### GitLab CI集成

```yaml
# .gitlab-ci.yml
stages:
  - build
  - deploy

build:
  stage: build
  script:
    - docker build -t myapp:$CI_COMMIT_SHA .
    - docker push myapp:$CI_COMMIT_SHA

deploy:
  stage: deploy
  script:
    - git clone https://gitlab.com/myorg/myapp-config.git
    - cd myapp-config
    - sed -i "s|image:.*|image: myapp:$CI_COMMIT_SHA|" overlays/production/deployment.yaml
    - git commit -am "Update image to $CI_COMMIT_SHA"
    - git push
  only:
    - main
```

---

## 演进式发布策略

### 蓝绿部署

```yaml
# ArgoCD Rollouts
apiVersion: argoproj.io/v1alpha1
kind: Rollout
metadata:
  name: myapp
spec:
  replicas: 5
  strategy:
    blueGreen:
      activeService: myapp-active
      previewService: myapp-preview
      autoPromotionEnabled: false
      prePromotionAnalysis:
        templates:
          - templateName: success-rate
  selector:
    matchLabels:
      app: myapp
  template:
    # Pod模板...
```

### 金丝雀发布

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Rollout
metadata:
  name: myapp
spec:
  replicas: 5
  strategy:
    canary:
      steps:
        - setWeight: 20
        - pause: {duration: 10m}
        - setWeight: 40
        - pause: {duration: 10m}
        - setWeight: 60
        - pause: {duration: 10m}
        - setWeight: 80
        - pause: {duration: 10m}
      analysis:
        templates:
          - templateName: success-rate
        startingStep: 2
        args:
          - name: service-name
            value: myapp-canary
```

---

## 监控与告警

### ArgoCD监控

```yaml
# Prometheus ServiceMonitor
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: argocd-metrics
  namespace: argocd
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: argocd-metrics
  endpoints:
    - port: metrics
```

### 同步状态告警

```yaml
# PrometheusRule
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: argocd-alerts
  namespace: argocd
spec:
  groups:
    - name: argocd
      rules:
        - alert: ApplicationOutOfSync
          expr: argocd_app_info{sync_status="OutOfSync"} == 1
          for: 10m
          labels:
            severity: warning
          annotations:
            summary: "Application {{ $labels.name }} is out of sync"
        
        - alert: ApplicationSyncFailed
          expr: argocd_app_sync_total{phase="Failed"} > 0
          for: 5m
          labels:
            severity: critical
          annotations:
            summary: "Application {{ $labels.name }} sync failed"
```

---

## 最佳实践

### GitOps实践清单

```
[ ] Git作为唯一真实来源
[ ] 声明式配置（YAML）
[ ] 所有变更通过Git提交
[ ] 自动同步策略配置
[ ] 多环境配置隔离
[ ] 密钥管理（Sealed Secrets）
[ ] PR审查流程
[ ] 自动化测试
[ ] 回滚策略明确
[ ] 监控告警配置
```

### 密钥管理

```bash
# 使用Sealed Secrets
kubectl apply -f https://github.com/bitnami-labs/sealed-secrets/releases/download/v0.18.0/controller.yaml

# 加密Secret
kubeseal --format=yaml < secret.yaml > sealed-secret.yaml

# 提交到Git
git add sealed-secret.yaml
git commit -m "Add sealed secret"
git push
```

---

## 学习检查点

完成本章节后，验证：

- [ ] 安装并配置ArgoCD/FluxCD
- [ ] 创建Application/ Kustomization资源
- [ ] 实现Git提交自动部署
- [ ] 配置多环境管理
- [ ] 集成CI/CD流程
- [ ] 实现蓝绿/金丝雀发布
- [ ] 配置监控告警
- [ ] 掌握回滚操作

---

## 延伸阅读

- [ArgoCD官方文档](https://argo-cd.readthedocs.io/)
- [FluxCD官方文档](https://fluxcd.io/docs/)
- [GitOps最佳实践](https://www.gitops.tech/)
- [Sealed Secrets](https://github.com/bitnami-labs/sealed-secrets)
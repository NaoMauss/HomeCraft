# HomeCraft Backend Deployment Guide

## Overview

The HomeCraft backend is deployed using:
- **Helm** for packaging and templating Kubernetes resources
- **GitHub Actions** for CI/CD (build Docker images)
- **Flux GitOps** for automated deployment
- **GitHub Container Registry (GHCR)** for Docker images

## Architecture

```
GitHub Push (main)
  ↓
GitHub Actions Workflow
  ├─ Build Docker Image → ghcr.io/naomauss/homecraft-backend:main-<sha>
  └─ Update HelmRelease YAML → infra/releases/homecraft-backend.yaml
       ↓
Git Commit & Push
  ↓
Flux Detects Change
  ↓
Flux Deploys Helm Chart
  ↓
Backend Pods Running in Kubernetes
```

## Image Tagging Strategy

To solve the "latest tag won't trigger rollout" problem, we use:

✅ **Git SHA-based tags**: `main-abc1234` (unique per commit)
✅ **imagePullPolicy: Always** (fallback safety)
✅ **Automated HelmRelease updates** (GitHub Actions updates the tag)

### How It Works

1. **Push to main** → GitHub Actions triggered
2. **Build Docker image** → Tagged with `main-<short-sha>`
3. **Push to GHCR** → Image available at `ghcr.io/naomauss/homecraft-backend:main-abc1234`
4. **Update HelmRelease** → GitHub Actions updates `infra/releases/homecraft-backend.yaml`
5. **Commit change** → Automated commit pushed to repo
6. **Flux syncs** → Detects new tag in HelmRelease
7. **Helm upgrade** → Deploys new version with new image tag
8. **Pods restart** → New image pulled and deployed

## Directory Structure

```
HomeCraft/
├── backend/
│   ├── chart/                        # Helm chart
│   │   ├── Chart.yaml                # Chart metadata
│   │   ├── values.yaml               # Default values
│   │   ├── templates/
│   │   │   ├── _helpers.tpl          # Template helpers
│   │   │   ├── deployment.yaml       # Deployment
│   │   │   ├── service.yaml          # Service
│   │   │   ├── serviceaccount.yaml   # RBAC
│   │   │   ├── configmap.yaml        # Environment variables
│   │   │   └── NOTES.txt             # Post-install notes
│   │   └── .helmignore
│   └── Dockerfile                    # Multi-stage Docker build
├── .github/
│   └── workflows/
│       └── build-and-deploy.yaml     # CI/CD workflow
└── infra/
    └── releases/
        └── homecraft-backend.yaml    # Flux HelmRelease
```

## Helm Chart Details

### Chart Information
- **Name**: `homecraft-backend`
- **Type**: `application`
- **Version**: `0.1.0`
- **App Version**: `1.0.0`

### Key Features
- ✅ **2 replicas** for high availability
- ✅ **Health checks** (liveness & readiness probes)
- ✅ **RBAC** for Kubernetes API access
- ✅ **Security contexts** (non-root, read-only filesystem)
- ✅ **Resource limits** (CPU/Memory)
- ✅ **ConfigMap** for environment variables
- ✅ **ClusterRole** for MinecraftServer CRD access
- ✅ **Role** for minecraft-servers namespace access

### Resources Created
- **Deployment**: `homecraft-backend`
- **Service**: `homecraft-backend` (ClusterIP on port 80)
- **ServiceAccount**: `homecraft-api`
- **ClusterRole** + **ClusterRoleBinding**: CRD permissions
- **Role** + **RoleBinding**: Namespace permissions
- **ConfigMap**: Environment variables

## GitHub Actions Workflow

### Trigger Conditions
- Push to `main` branch
- Changes in `backend/**` or workflow file
- Manual trigger (`workflow_dispatch`)

### Jobs

#### 1. `build-and-push`
- Checks out code
- Sets up Docker Buildx
- Logs into GHCR
- Builds Docker image
- Tags with:
  - `main-<short-sha>` (primary)
  - `main` (branch tag)
  - `latest` (for main branch)
- Pushes to GHCR
- Uses GitHub Actions cache for faster builds

#### 2. `update-helm-values`
- Checks out code
- Extracts Git SHA
- Updates `infra/releases/homecraft-backend.yaml`
- Sets image tag to `main-<short-sha>`
- Commits and pushes change
- Commit message: `chore: update backend image to <sha>`

## Flux HelmRelease

### Configuration
```yaml
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  name: homecraft-backend
  namespace: flux-system
spec:
  interval: 5m
  chart:
    spec:
      chart: ./backend/chart
      sourceRef:
        kind: GitRepository
        name: flux-system
  targetNamespace: default
  values:
    image:
      repository: ghcr.io/naomauss/homecraft-backend
      tag: "latest"  # Updated by GitHub Actions
      pullPolicy: Always
    replicaCount: 2
    # ... more values
```

### Behavior
- **Syncs every 5 minutes** or on Git change
- **Reads chart** from `./backend/chart` in Git repo
- **Deploys to** `default` namespace
- **Retries** 3 times on failure
- **Automatically upgrades** when values change (new image tag)

## Deployment Flow

### Initial Setup

1. **Commit and push** all changes:
   ```bash
   git add .
   git commit -m "feat: add Helm chart and CI/CD"
   git push origin main
   ```

2. **GitHub Actions runs** automatically
   - Builds Docker image
   - Pushes to GHCR (might require GHCR setup)
   - Updates HelmRelease YAML
   - Pushes commit

3. **Flux syncs** within 5 minutes (or force):
   ```bash
   flux reconcile kustomization flux-system --with-source
   ```

4. **Verify deployment**:
   ```bash
   # Check HelmRelease status
   kubectl get helmrelease homecraft-backend -n flux-system
   
   # Check pods
   kubectl get pods -l app.kubernetes.io/name=homecraft-backend
   
   # Check service
   kubectl get svc homecraft-backend
   
   # View logs
   kubectl logs -l app.kubernetes.io/name=homecraft-backend --tail=50
   ```

### Subsequent Updates

Any push to `backend/**` on main branch:
1. **GitHub Actions** builds new image with new SHA tag
2. **GitHub Actions** updates HelmRelease with new tag
3. **Flux** detects change and upgrades release
4. **Kubernetes** performs rolling update
5. **Old pods** terminate, **new pods** start

## Troubleshooting

### GitHub Actions Not Running
```bash
# Check workflow file syntax
cat .github/workflows/build-and-deploy.yaml

# View workflow runs
gh run list --workflow=build-and-deploy.yaml

# View specific run
gh run view <run-id>
```

### Image Not Building
```bash
# Check Docker build locally
cd backend
docker build -t test:local .

# Check GHCR login
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin
```

### HelmRelease Not Updating
```bash
# Check HelmRelease status
kubectl describe helmrelease homecraft-backend -n flux-system

# Check Flux logs
kubectl logs -n flux-system deployment/helm-controller

# Force reconciliation
flux reconcile helmrelease homecraft-backend

# Check if chart is valid
kubectl get helmchart -n flux-system
```

### Pods Not Starting
```bash
# Check pod status
kubectl get pods -l app.kubernetes.io/name=homecraft-backend

# Describe pod
kubectl describe pod <pod-name>

# Check events
kubectl get events --sort-by='.lastTimestamp'

# Check logs
kubectl logs <pod-name>
```

### Image Pull Errors
```bash
# Check if image exists in GHCR
docker pull ghcr.io/naomauss/homecraft-backend:main-<sha>

# Check image pull secret (if needed)
kubectl get secrets

# Make GHCR package public
# Go to: https://github.com/users/NaoMauss/packages/container/homecraft-backend/settings
```

### RBAC Permissions Issues
```bash
# Check ServiceAccount
kubectl get sa homecraft-api

# Check ClusterRole
kubectl get clusterrole | grep homecraft

# Check ClusterRoleBinding
kubectl describe clusterrolebinding | grep homecraft

# Test API access from pod
kubectl exec -it <pod-name> -- sh
# (if shell available, test kubectl commands)
```

## Configuration

### Environment Variables
Set in `values.yaml` under `env:`:
- `GIN_MODE`: "release" (production mode)
- `PORT`: "8080" (API port)

### Resource Limits
Default limits in `values.yaml`:
```yaml
resources:
  limits:
    cpu: 500m
    memory: 128Mi
  requests:
    cpu: 100m
    memory: 64Mi
```

### Scaling
Adjust replicas in HelmRelease:
```yaml
values:
  replicaCount: 3  # Scale to 3 replicas
```

Or enable autoscaling in `values.yaml`:
```yaml
autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 10
  targetCPUUtilizationPercentage: 80
```

## Security

### GHCR Package Visibility
Make package public to avoid pull authentication issues:
1. Go to: `https://github.com/users/NaoMauss/packages`
2. Select `homecraft-backend` package
3. Package settings → Change visibility → Public

### Image Pull Secrets (if private)
Create secret:
```bash
kubectl create secret docker-registry ghcr-secret \
  --docker-server=ghcr.io \
  --docker-username=NaoMauss \
  --docker-password=$GITHUB_TOKEN \
  --docker-email=your-email@example.com
```

Update HelmRelease:
```yaml
values:
  imagePullSecrets:
    - name: ghcr-secret
```

## Monitoring

### Check Backend Health
```bash
# Port-forward to API
kubectl port-forward svc/homecraft-backend 8080:80

# Test health endpoint
curl http://localhost:8080/health
```

### Watch Deployments
```bash
# Watch pods
watch kubectl get pods -l app.kubernetes.io/name=homecraft-backend

# Watch HelmRelease
watch kubectl get helmrelease -n flux-system

# Stream logs
kubectl logs -f -l app.kubernetes.io/name=homecraft-backend
```

## Next Steps

After successful deployment:
1. ✅ Set up ingress for external access
2. ✅ Configure monitoring/alerting
3. ✅ Set up backup for persistent data
4. ✅ Implement controller/operator for MinecraftServer CRDs
5. ✅ Add frontend deployment

## References

- [Helm Documentation](https://helm.sh/docs/)
- [Flux HelmRelease](https://fluxcd.io/docs/components/helm/)
- [GitHub Actions](https://docs.github.com/en/actions)
- [GHCR Documentation](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry)

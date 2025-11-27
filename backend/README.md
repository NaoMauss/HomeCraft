# HomeCraft Backend API

A Kubernetes-native REST API for managing Minecraft servers using the Operator Pattern. This API creates and manages `MinecraftServer` Custom Resources that are reconciled by a separate controller.

## Architecture

- **Framework**: Gin (Go HTTP framework)
- **K8s Integration**: client-go for CRD management
- **Pattern**: Operator Pattern (API creates CRDs, controller manages workloads)
- **Workload**: Sidecar pattern (Minecraft + SFTP containers)

## Project Structure

```
backend/
├── cmd/
│   └── api/
│       └── main.go                    # Application entry point
├── pkg/
│   ├── apis/
│   │   └── homecraft/
│   │       └── v1alpha1/
│   │           ├── types.go           # MinecraftServer CRD types
│   │           ├── register.go        # Scheme registration
│   │           └── zz_generated.deepcopy.go
│   ├── handlers/
│   │   └── server_handler.go         # HTTP handlers
│   ├── k8s/
│   │   └── client.go                 # Kubernetes client wrapper
│   └── models/
│       └── request.go                # API request/response models
├── config/
│   └── crd/
│       └── minecraftserver-crd.yaml  # CRD manifest
├── Dockerfile                         # Multi-stage Docker build
├── Makefile                           # Build and deployment targets
└── README.md
```

## Prerequisites

- Go 1.25.4+
- kubectl configured to access your K3s cluster
- (Optional) Docker for building images

## Quick Start

### 1. Deploy Infrastructure with Flux

This project uses **Flux GitOps** for deployment. The namespace and CRD are automatically deployed from the `infra/` directory.

**Prerequisites:**
- Flux installed and configured on your cluster
- Repository connected to Flux

**Deploy HomeCraft resources:**

```bash
# Commit the infrastructure changes
git add infra/
git commit -m "Add HomeCraft namespace and CRD"
git push

# Wait for Flux to sync (automatic) or force reconciliation
flux reconcile kustomization flux-system --with-source

# Verify Flux has applied the resources
make verify
```

The Flux setup will automatically create:
1. `minecraft-servers` namespace
2. `MinecraftServer` CRD

**Manual verification:**

```bash
# Check namespace
kubectl get namespace minecraft-servers

# Check CRD
kubectl get crd minecraftservers.homecraft.io

# Check Flux Kustomization status
kubectl get kustomization homecraft -n flux-system
```

### 2. Run Locally

```bash
# Download dependencies
make deps

# Build and run
make run
```

The API will start on `http://localhost:8080`

### 3. Test the API

Health check:
```bash
curl http://localhost:8080/health
```

## API Endpoints

### Health Check
```
GET /health
```

Response:
```json
{
  "status": "healthy",
  "message": "HomeCraft API is running"
}
```

### Create Minecraft Server
```
POST /api/v1/servers
```

Request body:
```json
{
  "name": "my-server",
  "eula": true,
  "memory": "4Gi",
  "storageSize": "2Gi",
  "version": "1.20.1",
  "serverType": "VANILLA",
  "maxPlayers": 20,
  "difficulty": "normal",
  "gamemode": "survival"
}
```

**Note:** SFTP credentials are automatically generated and returned in the response.

Response (201 Created):
```json
{
  "name": "my-server",
  "namespace": "minecraft-servers",
  "eula": true,
  "sftpUsername": "mc-my-server",
  "sftpPassword": "Xa9K2mP7nQ4vL8tY",
  "memory": "4Gi",
  "storageSize": "2Gi",
  "version": "1.20.1",
  "serverType": "VANILLA",
  "maxPlayers": 20,
  "difficulty": "normal",
  "gamemode": "survival",
  "phase": "Pending",
  "createdAt": "2025-11-27T10:00:00Z"
}
```

### List All Servers
```
GET /api/v1/servers
```

Response:
```json
{
  "items": [
    {
      "name": "my-server",
      "namespace": "minecraft-servers",
      "phase": "Running",
      "endpoint": "my-server.example.com:25565",
      ...
    }
  ],
  "count": 1
}
```

### Get Specific Server
```
GET /api/v1/servers/:name
```

Response:
```json
{
  "name": "my-server",
  "namespace": "minecraft-servers",
  "eula": true,
  "memory": "4Gi",
  ...
}
```

### Delete Server
```
DELETE /api/v1/servers/:name
```

Response:
```json
{
  "message": "Server deleted successfully",
  "name": "my-server"
}
```

## Configuration

Environment variables:

- `PORT` - API port (default: 8080)
- `GIN_MODE` - Gin mode: debug/release (default: debug)
- `KUBECONFIG` - Path to kubeconfig file (for local development)

## Development

### Build Commands

```bash
# Build binary
make build

# Run tests
make test

# Format code
make fmt

# Run linters
make lint

# Clean build artifacts
make clean
```

### Docker

Build Docker image:
```bash
make docker-build
```

Build with custom tag:
```bash
make docker-build DOCKER_TAG=v1.0.0
```

### Local Development Setup

```bash
# Install CRD and prepare environment
make local-dev

# Run the API
make run
```

## Infrastructure

All Kubernetes resources are managed via **Flux GitOps** in the `infra/` directory:

```
infra/
├── flux-system/              # Flux GitOps components
├── homecraft/
│   └── base/                 # HomeCraft base resources
│       ├── namespace.yaml    # minecraft-servers namespace
│       ├── minecraftserver-crd.yaml # MinecraftServer CRD
│       └── kustomization.yaml
├── repositories/             # Helm chart repositories (HelmRepository)
│   └── kustomization.yaml
├── releases/                 # Helm releases (HelmRelease)
│   └── kustomization.yaml
└── kustomization.yaml        # Root kustomization (includes all directories)
```

**Directory Structure:**
- `homecraft/base/` - Raw Kubernetes manifests for HomeCraft
- `repositories/` - Flux HelmRepository resources for Helm chart repos
- `releases/` - Flux HelmRelease resources for Helm chart deployments

**To deploy changes:**

```bash
# 1. Add/modify resources in appropriate directory
vim infra/homecraft/base/namespace.yaml

# 2. Commit and push
git add infra/
git commit -m "Update HomeCraft infrastructure"
git push

# 3. Flux automatically syncs (or force it)
flux reconcile kustomization flux-system --with-source
```

## Deployment

### Deploy to Kubernetes

1. Build and push Docker image:
```bash
make docker-build docker-push
```

2. Create API deployment in `infra/` (to be managed by Flux):

**Example Deployment manifest:**
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: homecraft-api
spec:
  replicas: 2
  selector:
    matchLabels:
      app: homecraft-api
  template:
    metadata:
      labels:
        app: homecraft-api
    spec:
      serviceAccountName: homecraft-api
      containers:
      - name: api
        image: homecraft/backend:latest
        ports:
        - containerPort: 8080
        env:
        - name: GIN_MODE
          value: "release"
---
apiVersion: v1
kind: Service
metadata:
  name: homecraft-api
spec:
  selector:
    app: homecraft-api
  ports:
  - port: 80
    targetPort: 8080
  type: ClusterIP
```

3. Create ServiceAccount with RBAC:
```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: homecraft-api
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: homecraft-api
rules:
- apiGroups: ["homecraft.io"]
  resources: ["minecraftservers"]
  verbs: ["get", "list", "create", "update", "delete"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: homecraft-api
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: homecraft-api
subjects:
- kind: ServiceAccount
  name: homecraft-api
  namespace: default
---
# Additional permissions for managing the minecraft-servers namespace
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: homecraft-namespace-manager
  namespace: minecraft-servers
rules:
- apiGroups: [""]
  resources: ["pods", "services", "persistentvolumeclaims"]
  verbs: ["get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: homecraft-namespace-manager
  namespace: minecraft-servers
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: homecraft-namespace-manager
subjects:
- kind: ServiceAccount
  name: homecraft-api
  namespace: default
```

## Cluster Resources

Check available cluster resources before creating servers:

```
GET /api/v1/cluster/resources
```

Response:
```json
{
  "totalMemory": "16.0 GiB",
  "allocatedMemory": "4.0 GiB",
  "availableMemory": "12.0 GiB"
}
```

## CRD Specification

The `MinecraftServer` CRD spec includes:

### Required Fields
- `eula` (bool) - Accept Minecraft EULA
- `memory` (string) - Memory allocation (e.g., "2Gi", "4Gi")
  - API checks cluster capacity before creation
  - Must be in format: `<number>Mi`, `<number>Gi`, or `<number>Ti`

### Optional Fields
- `storageSize` (string) - PVC size (default: "1Gi")
- `version` (string) - Minecraft version (default: "LATEST")
- `serverType` (string) - Server type: VANILLA, PAPER, FORGE (default: "VANILLA")
- `maxPlayers` (int) - Max players (default: 20)
- `difficulty` (string) - peaceful/easy/normal/hard (default: "normal")
- `gamemode` (string) - survival/creative/adventure/spectator (default: "survival")

### Auto-Generated Fields
- `sftpUsername` (string) - Automatically generated as `mc-<server-name>`
- `sftpPassword` (string) - Automatically generated 16-character secure password

**Important:** All MinecraftServer resources are created in the dedicated `minecraft-servers` namespace.

## Next Steps

This API creates `MinecraftServer` Custom Resources. To actually deploy the Minecraft servers, you need to implement a **Kubernetes Controller/Operator** that:

1. Watches `MinecraftServer` resources
2. Creates StatefulSets with 2 containers:
   - `itzg/minecraft-server` (game server)
   - `atmoz/sftp` (file access)
3. Creates Services and PVCs
4. Updates the `status` field with endpoints

See the `/operator` directory for the controller implementation.

## Troubleshooting

### Namespace or CRD Not Found
```bash
# Check Flux sync status
kubectl get kustomization flux-system -n flux-system

# Check if resources exist
kubectl get namespace minecraft-servers
kubectl get crd minecraftservers.homecraft.io

# Force Flux reconciliation (syncs everything in infra/)
flux reconcile kustomization flux-system --with-source

# Check Flux logs
kubectl logs -n flux-system deployment/kustomize-controller

# Verify kustomization builds locally
kubectl kustomize infra/
```

### Flux Not Syncing
```bash
# Verify Flux is running
kubectl get pods -n flux-system

# Check GitRepository source
kubectl get gitrepository flux-system -n flux-system

# Force reconciliation
flux reconcile source git flux-system
flux reconcile kustomization flux-system
```

### Connection Refused
Make sure your kubeconfig is properly configured:
```bash
kubectl cluster-info
```

### Permission Denied
Ensure your ServiceAccount has the correct RBAC permissions (see Deployment section).

## License

MIT

## Contributing

Contributions are welcome! Please open an issue or pull request.

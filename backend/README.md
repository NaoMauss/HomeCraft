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

### 1. Install the CRD

Before running the API, you must register the `MinecraftServer` CRD with your cluster:

```bash
make deploy-crd
```

Verify the CRD is installed:

```bash
kubectl get crd minecraftservers.homecraft.io
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
  "sftpUsers": [
    "player1:password123:1000:1000:minecraft"
  ],
  "storageSize": "2Gi",
  "version": "1.20.1",
  "serverType": "VANILLA",
  "maxPlayers": 20,
  "difficulty": "normal",
  "gamemode": "survival"
}
```

Response (201 Created):
```json
{
  "name": "my-server",
  "namespace": "default",
  "eula": true,
  "sftpUsers": ["player1:password123:1000:1000:minecraft"],
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
      "namespace": "default",
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
  "namespace": "default",
  "eula": true,
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

## Deployment

### Deploy to Kubernetes

1. Build and push Docker image:
```bash
make docker-build docker-push
```

2. Create a Deployment manifest (example):
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
```

## CRD Specification

The `MinecraftServer` CRD spec includes:

### Required Fields
- `eula` (bool) - Accept Minecraft EULA
- `sftpUsers` ([]string) - SFTP users in format: `user:pass:uid:gid:directory`
- `storageSize` (string) - PVC size (e.g., "1Gi", "5Gi")

### Optional Fields
- `version` (string) - Minecraft version (default: "LATEST")
- `serverType` (string) - Server type: VANILLA, PAPER, FORGE (default: "VANILLA")
- `maxPlayers` (int) - Max players (default: 20)
- `difficulty` (string) - peaceful/easy/normal/hard (default: "normal")
- `gamemode` (string) - survival/creative/adventure/spectator (default: "survival")

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

### CRD Not Found
```bash
# Check if CRD exists
kubectl get crd minecraftservers.homecraft.io

# Reinstall CRD
make deploy-crd
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

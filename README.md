# HomeCraft

A Kubernetes-native Minecraft server management platform that allows users to easily create, manage, and delete Minecraft servers through a modern web interface.

## Overview

HomeCraft uses the Kubernetes Operator Pattern to automate Minecraft server provisioning. When you create a server through the web UI or API, the system:

1. Creates a `MinecraftServer` Custom Resource in Kubernetes
2. The operator watches for these resources and automatically provisions:
   - StatefulSet with Minecraft server + SFTP sidecar containers
   - PersistentVolumeClaim for game data
   - LoadBalancer services for game and SFTP access
   - Secrets for SFTP credentials

## Features

- **Multi-Server Management** - Create and manage multiple Minecraft servers
- **Server Types** - Support for Vanilla, Paper, Forge, and Fabric
- **Configurable Resources** - Customize memory, storage, and player limits
- **SFTP Access** - Auto-generated credentials for file management
- **Modern Web UI** - Vue 3 interface with real-time status updates
- **GitOps Ready** - Flux CD integration for automated deployments
- **External Access** - Support for Cloudflare Tunnel and Playit.gg

## Tech Stack

| Component | Technology |
|-----------|------------|
| Backend | Go 1.25, Gin, client-go |
| Frontend | Vue 3, TypeScript, Vite, TailwindCSS |
| Operator | Go 1.25, controller-runtime |
| Infrastructure | Kubernetes (K3s), Flux CD, Helm |

## Project Structure

```
HomeCraft/
├── backend/          # Go REST API
│   ├── cmd/api/      # Entry point
│   ├── pkg/          # Application code
│   └── chart/        # Helm chart
├── front/            # Vue.js frontend
│   ├── src/          # Application source
│   └── chart/        # Helm chart
├── operator/         # Kubernetes Operator
│   ├── controllers/  # Reconciliation logic
│   └── chart/        # Helm chart
├── infra/            # Flux GitOps configuration
│   ├── flux-system/  # Flux components
│   ├── releases/     # HelmRelease definitions
│   └── homecraft/    # Base resources
└── .github/workflows/ # CI/CD pipelines
```

## Prerequisites

- Go 1.25.4+
- Node.js 18+
- kubectl configured for your cluster
- Kubernetes cluster (K3s recommended)
- Flux CD installed on the cluster
- Docker (for building images)

## Quick Start

### Local Development

**Backend:**
```bash
cd backend
make deps   # Download dependencies
make run    # Run on :8080
```

**Frontend:**
```bash
cd front
npm install
npm run dev # Start dev server
```

### Deploy to Kubernetes

The project uses Flux CD for GitOps-based deployments:

```bash
# Push changes to main branch
git add . && git commit -m "Deploy changes" && git push

# Force Flux reconciliation (optional)
flux reconcile kustomization flux-system --with-source
```

For detailed deployment instructions, see [DEPLOYMENT.md](./DEPLOYMENT.md).

## API Reference

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| POST | `/api/v1/servers` | Create a Minecraft server |
| GET | `/api/v1/servers` | List all servers |
| GET | `/api/v1/servers/:name` | Get server details |
| DELETE | `/api/v1/servers/:name` | Delete a server |
| GET | `/api/v1/cluster/resources` | Get cluster resources |

### Create Server Request

```json
{
  "name": "my-server",
  "eula": true,
  "memory": "2Gi",
  "storageSize": "10Gi",
  "version": "1.20.4",
  "serverType": "PAPER",
  "maxPlayers": 20,
  "difficulty": "normal",
  "gamemode": "survival"
}
```

## MinecraftServer CRD

The custom resource definition supports:

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| eula | boolean | Yes | - | Accept Minecraft EULA |
| memory | string | Yes | - | Memory allocation (e.g., "2Gi") |
| storageSize | string | Yes | - | PVC size (e.g., "10Gi") |
| version | string | No | "LATEST" | Minecraft version |
| serverType | string | No | "VANILLA" | VANILLA, PAPER, FORGE, FABRIC |
| maxPlayers | int | No | 20 | Maximum players |
| difficulty | string | No | "normal" | peaceful/easy/normal/hard |
| gamemode | string | No | "survival" | survival/creative/adventure/spectator |

## Environment Variables

**Backend:**
| Variable | Description | Default |
|----------|-------------|---------|
| PORT | API port | 8080 |
| GIN_MODE | debug/release | debug |
| KUBECONFIG | Path to kubeconfig | In-cluster |

**Frontend:**
| Variable | Description | Default |
|----------|-------------|---------|
| VITE_API_URL | Backend API URL | Same origin |

## Architecture

Each Minecraft server deployment creates:

- **StatefulSet** with 2 containers:
  - `itzg/minecraft-server` - Game server
  - `atmoz/sftp` - SFTP sidecar for file access
- **PersistentVolumeClaim** - Shared storage
- **Service (Minecraft)** - LoadBalancer on port 25565
- **Service (SFTP)** - LoadBalancer on port 22
- **Secret** - Auto-generated SFTP credentials

## Documentation

- [Backend README](./backend/README.md) - API documentation
- [Backend Architecture](./backend/ARCHITECTURE.md) - K8s client deep dive
- [Deployment Guide](./DEPLOYMENT.md) - Full deployment instructions

## License

See [LICENSE](./LICENSE) for details.

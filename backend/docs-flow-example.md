# Example Request Flow

## Creating a Minecraft Server: Step-by-Step

### 1. User sends HTTP request
```bash
curl -X POST http://localhost:8080/api/v1/servers \
  -H "Content-Type: application/json" \
  -d '{
    "name": "survival-world",
    "eula": true,
    "sftpUsers": ["player1:password:1000:1000:minecraft"],
    "storageSize": "5Gi"
  }'
```

### 2. Gin routes to handler (cmd/api/main.go:37)
```go
v1.POST("/servers", serverHandler.CreateServer)
```

### 3. Handler validates and creates CR (pkg/handlers/server_handler.go:47)
```go
server := &v1alpha1.MinecraftServer{
    ObjectMeta: metav1.ObjectMeta{
        Name: "survival-world",
        Namespace: "default",
    },
    Spec: v1alpha1.MinecraftServerSpec{
        EULA: true,
        SFTPUsers: ["player1:password:1000:1000:minecraft"],
        StorageSize: "5Gi",
    },
}

result, err := h.k8sClient.CreateMinecraftServer(ctx, "default", server)
```

### 4. K8s client makes HTTPS call (pkg/k8s/client.go:77-82)
```go
err := c.restClient.Post().
    Namespace("default").
    Resource("minecraftservers").
    Body(server).
    Do(ctx).
    Into(result)
```

### 5. Actual HTTPS request to K8s API
```http
POST https://10.43.0.1:443/apis/homecraft.io/v1alpha1/namespaces/default/minecraftservers
Authorization: Bearer <service-account-token>
Content-Type: application/json

{
  "apiVersion": "homecraft.io/v1alpha1",
  "kind": "MinecraftServer",
  "metadata": {
    "name": "survival-world",
    "namespace": "default"
  },
  "spec": {
    "eula": true,
    "sftpUsers": ["player1:password:1000:1000:minecraft"],
    "storageSize": "5Gi",
    "version": "LATEST",
    "serverType": "VANILLA",
    "maxPlayers": 20,
    "difficulty": "normal",
    "gamemode": "survival"
  }
}
```

### 6. K8s API Server responds
```json
{
  "apiVersion": "homecraft.io/v1alpha1",
  "kind": "MinecraftServer",
  "metadata": {
    "name": "survival-world",
    "namespace": "default",
    "uid": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "creationTimestamp": "2025-11-27T22:15:30Z",
    "resourceVersion": "123456"
  },
  "spec": {
    "eula": true,
    "sftpUsers": ["player1:password:1000:1000:minecraft"],
    "storageSize": "5Gi",
    "version": "LATEST",
    "serverType": "VANILLA",
    "maxPlayers": 20,
    "difficulty": "normal",
    "gamemode": "survival"
  },
  "status": {
    "phase": "Pending"
  }
}
```

### 7. Handler returns to user
```json
HTTP/1.1 201 Created
Content-Type: application/json

{
  "name": "survival-world",
  "namespace": "default",
  "eula": true,
  "sftpUsers": ["player1:password:1000:1000:minecraft"],
  "storageSize": "5Gi",
  "version": "LATEST",
  "serverType": "VANILLA",
  "maxPlayers": 20,
  "difficulty": "normal",
  "gamemode": "survival",
  "phase": "Pending",
  "createdAt": "2025-11-27T22:15:30Z"
}
```

### 8. Verify it exists in K8s
```bash
# Check if the CR was created
kubectl get minecraftservers.homecraft.io

# Output:
NAME             PHASE     ENDPOINT   AGE
survival-world   Pending              5s

# Get full details
kubectl get minecraftservers.homecraft.io survival-world -o yaml
```

### 9. What happens next?
At this point, the MinecraftServer resource exists in etcd, but **no pods are running**.

The **controller/operator** (not built yet) will:
1. Watch for new MinecraftServer resources
2. Create a StatefulSet with Minecraft + SFTP containers
3. Create a Service for network access
4. Create a PVC for persistent storage
5. Update the MinecraftServer status with the endpoint

Only then will the status change from `Pending` → `Running` and you'll get an endpoint!

---

## Key Takeaway

The **backend API** is just a friendly REST interface that:
- Accepts JSON from users
- Validates the input
- Creates Kubernetes Custom Resources
- **Does NOT create Pods directly**

The **controller** (future task) does the actual work:
- Watches Custom Resources
- Creates/manages workloads (StatefulSets, Services, etc.)
- Updates status fields

This separation is the **Operator Pattern**! 🎯

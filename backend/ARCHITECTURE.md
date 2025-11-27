# How HomeCraft Backend Works: Architecture Deep Dive

## 📡 Yes, the K8s Client Talks Directly to the K8s API Server!

You're absolutely correct. The Kubernetes client (`client-go`) communicates directly with the Kubernetes API server over HTTPS. Let me break down exactly how this works.

---

## 🔐 Step 1: Authentication & Connection

### In-Cluster Configuration (When Running in K3s)
When the backend API runs **inside** your K3s cluster as a Pod:

```go
config, err := rest.InClusterConfig()
```

This reads:
- **Service Account Token**: `/var/run/secrets/kubernetes.io/serviceaccount/token`
- **CA Certificate**: `/var/run/secrets/kubernetes.io/serviceaccount/ca.crt`
- **API Server Address**: From `KUBERNETES_SERVICE_HOST` and `KUBERNETES_SERVICE_PORT` environment variables

Example values:
```bash
KUBERNETES_SERVICE_HOST=10.43.0.1
KUBERNETES_SERVICE_PORT=443
TOKEN=/var/run/secrets/kubernetes.io/serviceaccount/token
```

The client uses this token to authenticate every request to the API server.

---

### Local Development Configuration (Kubeconfig)
When running locally (outside the cluster):

```go
loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
config, err = kubeConfig.ClientConfig()
```

This reads your `~/.kube/config` file:
```yaml
apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: LS0tLS1CRUdJTi...
    server: https://192.168.1.100:6443  # Your K3s API server
  name: k3s
contexts:
- context:
    cluster: k3s
    user: admin
  name: k3s
current-context: k3s
users:
- name: admin
  user:
    client-certificate-data: LS0tLS1CRUdJTi...
    client-key-data: LS0tLS1CRUdJTi...
```

---

## 🌐 Step 2: REST Client Setup for Custom Resources

Since `MinecraftServer` is a **Custom Resource** (not a built-in K8s type), we need a special REST client:

```go
// Lines 51-64 in client.go
crdConfig := *config
crdConfig.ContentConfig.GroupVersion = &schema.GroupVersion{
    Group:   "homecraft.io",      // Our custom API group
    Version: "v1alpha1",           // Our API version
}
crdConfig.APIPath = "/apis"        // Custom resources use /apis path
crdConfig.NegotiatedSerializer = serializer.NewCodecFactory(scheme)

restClient, err := rest.UnversionedRESTClientFor(&crdConfig)
```

This configures the client to talk to:
```
https://<k8s-api-server>/apis/homecraft.io/v1alpha1/namespaces/default/minecraftservers
```

---

## 🔄 Step 3: API Request Flow

Let's trace what happens when a user creates a server:

### 1. User Makes HTTP Request
```bash
curl -X POST http://localhost:8080/api/v1/servers \
  -d '{"name": "my-server", "eula": true, "sftpUsers": ["user:pass:1000:1000:minecraft"]}'
```

### 2. Gin Router → Handler
```go
// cmd/api/main.go:37
v1.POST("/servers", serverHandler.CreateServer)
```

### 3. Handler Validates & Transforms
```go
// pkg/handlers/server_handler.go:47
func (h *ServerHandler) CreateServer(c *gin.Context) {
    // Parse JSON request
    var req models.CreateServerRequest
    c.ShouldBindJSON(&req)
    
    // Create MinecraftServer CR object
    server := &v1alpha1.MinecraftServer{
        ObjectMeta: metav1.ObjectMeta{
            Name: req.Name,
            Namespace: "default",
        },
        Spec: v1alpha1.MinecraftServerSpec{
            EULA: req.EULA,
            SFTPUsers: req.SFTPUsers,
            // ... other fields
        },
    }
    
    // Send to K8s API
    result, err := h.k8sClient.CreateMinecraftServer(ctx, "default", server)
}
```

### 4. K8s Client Makes HTTPS Request
```go
// pkg/k8s/client.go:75-86
func (c *Client) CreateMinecraftServer(...) (*v1alpha1.MinecraftServer, error) {
    result := &v1alpha1.MinecraftServer{}
    err := c.restClient.Post().
        Namespace(namespace).           // /namespaces/default
        Resource("minecraftservers").   // /minecraftservers
        Body(server).                   // JSON body with spec
        Do(ctx).                        // Execute HTTP POST
        Into(result)                    // Decode response
    return result, nil
}
```

### 5. What Actually Happens on the Wire

The client-go library makes an **HTTPS POST** request:

```http
POST https://10.43.0.1:443/apis/homecraft.io/v1alpha1/namespaces/default/minecraftservers
Authorization: Bearer eyJhbGciOiJSUzI1NiIsImtpZCI6Ij...  (Service Account Token)
Content-Type: application/json

{
  "apiVersion": "homecraft.io/v1alpha1",
  "kind": "MinecraftServer",
  "metadata": {
    "name": "my-server",
    "namespace": "default"
  },
  "spec": {
    "eula": true,
    "sftpUsers": ["user:pass:1000:1000:minecraft"],
    "storageSize": "1Gi",
    "version": "LATEST",
    ...
  }
}
```

### 6. Kubernetes API Server Processes Request

The K8s API server:
1. **Authenticates** the token (checks if ServiceAccount is valid)
2. **Authorizes** the request (checks RBAC: does this account have permission to create MinecraftServers?)
3. **Validates** the payload against the CRD schema (the YAML we installed with `make deploy-crd`)
4. **Stores** the object in **etcd** (K8s database)
5. **Returns** the created object with additional metadata (UID, creation timestamp, etc.)

```json
{
  "apiVersion": "homecraft.io/v1alpha1",
  "kind": "MinecraftServer",
  "metadata": {
    "name": "my-server",
    "namespace": "default",
    "uid": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "creationTimestamp": "2025-11-27T22:00:00Z",
    "resourceVersion": "12345"
  },
  "spec": { ... },
  "status": {}
}
```

### 7. Handler Returns Response to User
```go
// pkg/handlers/server_handler.go:108
c.JSON(http.StatusCreated, convertToResponse(result))
```

---

## 🔍 What is etcd and Where is the Data?

**etcd** is Kubernetes' distributed key-value store (database). When you create a MinecraftServer:

```bash
# In etcd, the data is stored at:
/registry/homecraft.io/minecraftservers/default/my-server

# You can see it with:
kubectl get minecraftservers.homecraft.io my-server -o yaml
```

The object lives **only in etcd** until the controller/operator reconciles it.

---

## 🤖 What Happens Next? (The Controller's Job)

Right now, creating a `MinecraftServer` just stores data in etcd. **Nothing runs yet** because we haven't built the controller.

The **controller/operator** (future task) will:

```go
// Pseudo-code for the controller (not implemented yet)
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) {
    // 1. Watch for MinecraftServer resources
    server := &v1alpha1.MinecraftServer{}
    r.Get(ctx, req.NamespacedName, server)
    
    // 2. Create StatefulSet with 2 containers
    statefulset := &appsv1.StatefulSet{
        Spec: appsv1.StatefulSetSpec{
            Template: corev1.PodTemplateSpec{
                Spec: corev1.PodSpec{
                    Containers: []corev1.Container{
                        {
                            Name:  "minecraft",
                            Image: "itzg/minecraft-server",
                            Env: []corev1.EnvVar{
                                {Name: "EULA", Value: strconv.FormatBool(server.Spec.EULA)},
                                {Name: "VERSION", Value: server.Spec.Version},
                            },
                        },
                        {
                            Name:  "sftp",
                            Image: "atmoz/sftp",
                            Args:  server.Spec.SFTPUsers,
                        },
                    },
                },
            },
        },
    }
    r.Create(ctx, statefulset)
    
    // 3. Update MinecraftServer status
    server.Status.Phase = "Running"
    server.Status.Endpoint = "my-server.example.com:25565"
    r.Status().Update(ctx, server)
}
```

---

## 📊 Complete Data Flow Diagram

```
User                Backend API              K8s API Server         etcd            Controller
 │                      │                          │                  │                  │
 │ POST /servers        │                          │                  │                  │
 ├─────────────────────▶│                          │                  │                  │
 │                      │                          │                  │                  │
 │                      │ POST /apis/homecraft.io/ │                  │                  │
 │                      │   .../minecraftservers   │                  │                  │
 │                      ├─────────────────────────▶│                  │                  │
 │                      │                          │                  │                  │
 │                      │                          │ Validate CRD     │                  │
 │                      │                          │ Check RBAC       │                  │
 │                      │                          │                  │                  │
 │                      │                          │ STORE object     │                  │
 │                      │                          ├─────────────────▶│                  │
 │                      │                          │                  │                  │
 │                      │                          │                  │ WATCH event      │
 │                      │                          │                  ├─────────────────▶│
 │                      │                          │                  │                  │
 │                      │                          │                  │                  │ Create
 │                      │                          │                  │                  │ StatefulSet
 │                      │                          │                  │                  │ Service
 │                      │                          │                  │                  │ PVC
 │                      │                          │                  │                  │
 │                      │                          │ ◀────────────────┼──────────────────┤
 │                      │                          │ Update Status    │                  │
 │                      │                          │                  │                  │
 │                      │ ◀────────────────────────┤                  │                  │
 │                      │ Return MinecraftServer   │                  │                  │
 │                      │                          │                  │                  │
 │ ◀─────────────────────┤                          │                  │                  │
 │ 201 Created          │                          │                  │                  │
```

---

## 🔒 RBAC: Permissions Required

For the backend API to work, it needs a ServiceAccount with permissions:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: homecraft-api
rules:
- apiGroups: ["homecraft.io"]
  resources: ["minecraftservers"]
  verbs: ["get", "list", "create", "update", "delete"]
```

Without this, the API server will return `403 Forbidden`.

---

## 🎯 Summary

**Yes**, the K8s client in our backend talks **directly** to the Kubernetes API server via HTTPS:

1. **Authentication**: Service account token or kubeconfig credentials
2. **Authorization**: RBAC checks permissions
3. **Validation**: CRD schema validates the payload
4. **Storage**: etcd stores the MinecraftServer resource
5. **Watch**: Controller (future) watches for changes and creates workloads

The backend API is a **thin wrapper** that:
- Provides a friendly REST API for users
- Validates input
- Transforms JSON → MinecraftServer CRD
- Delegates actual workload creation to the controller

This is the **Operator Pattern** in action! 🚀

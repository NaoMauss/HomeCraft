# API Changes Summary

## What Changed

### 1. **Auto-Generated SFTP Credentials** ✅

**Before:**
Users had to provide SFTP credentials in the request:
```json
{
  "sftpUsers": ["player1:password:1000:1000:minecraft"]
}
```

**After:**
SFTP credentials are **automatically generated** by the API:
- Username format: `mc-<sanitized-server-name>`
- Password: Cryptographically secure 16-character random string
- Returned in the response for user to use

**Example:**
```bash
POST /api/v1/servers
{
  "name": "survival-world",
  "memory": "4Gi",
  ...
}

Response:
{
  "name": "survival-world",
  "sftpUsername": "mc-survival-world",
  "sftpPassword": "Xa9K2mP7nQ4vL8tY",
  ...
}
```

---

### 2. **Memory (RAM) Allocation** ✅

**New Required Field:**
Users must specify how much RAM the server needs:
```json
{
  "memory": "4Gi"  // Required! Can be "512Mi", "1Gi", "2Gi", "4Gi", etc.
}
```

**Features:**
- Validates format (must match pattern `^[0-9]+[MGT]i$`)
- Checks cluster capacity **before** creating the server
- Returns error if insufficient memory available

**Error Response Example:**
```json
{
  "error": "insufficient_capacity",
  "message": "insufficient memory: requested 4.0 GiB, available 2.5 GiB"
}
```

---

### 3. **Cluster Resource Endpoint** ✅

**New Endpoint:**
```
GET /api/v1/cluster/resources
```

**Purpose:**
Frontend can fetch available cluster resources to show users how much RAM they can allocate.

**Response:**
```json
{
  "totalMemory": "16.0 GiB",
  "allocatedMemory": "8.5 GiB",
  "availableMemory": "7.5 GiB",
  "totalNodes": 2,
  "nodes": [
    {
      "name": "node-1",
      "totalMemory": "8.0 GiB",
      "allocatedMemory": "4.2 GiB",
      "availableMemory": "3.8 GiB"
    },
    {
      "name": "node-2",
      "totalMemory": "8.0 GiB",
      "allocatedMemory": "4.3 GiB",
      "availableMemory": "3.7 GiB"
    }
  ]
}
```

---

## Updated API Endpoints

### POST /api/v1/servers

**New Request Body:**
```json
{
  "name": "my-server",          // Required
  "eula": true,                 // Required (must be true)
  "memory": "4Gi",              // Required (NEW!)
  "storageSize": "5Gi",         // Optional (default: "1Gi")
  "version": "1.20.1",          // Optional (default: "LATEST")
  "serverType": "VANILLA",      // Optional (default: "VANILLA")
  "maxPlayers": 20,             // Optional (default: 20)
  "difficulty": "normal",       // Optional (default: "normal")
  "gamemode": "survival"        // Optional (default: "survival")
}
```

**New Response Fields:**
```json
{
  "name": "my-server",
  "namespace": "default",
  "eula": true,
  "memory": "4Gi",              // Requested memory
  "storageSize": "5Gi",
  "sftpUsername": "mc-my-server",    // NEW! Auto-generated
  "sftpPassword": "Xa9K2mP7nQ4vL8tY", // NEW! Auto-generated
  "allocatedMemory": "4Gi",          // NEW! Actual allocated (from status)
  "phase": "Pending",
  "createdAt": "2025-11-27T22:30:00Z"
}
```

### GET /api/v1/cluster/resources (NEW!)

Returns cluster resource information for frontend.

---

## Implementation Details

### Auto-Generated Credentials (`pkg/utils/credentials.go`)

```go
username, password, err := utils.GenerateSFTPCredentials("my-server")
// username: "mc-my-server"
// password: "Xa9K2mP7nQ4vL8tY" (random)
```

Features:
- Sanitizes server name (removes special chars, lowercase)
- Generates cryptographically secure passwords
- No special characters that might cause shell issues

### Capacity Checking (`pkg/k8s/client.go`)

```go
hasCapacity, message, err := k8sClient.CheckMemoryAvailability(ctx, requestedMemory)
if !hasCapacity {
    return "insufficient memory: requested 4.0 GiB, available 2.5 GiB"
}
```

How it works:
1. Queries all nodes for allocatable memory
2. Sums up memory requests from all running pods
3. Calculates: `available = total - allocated`
4. Checks if `requested <= available`

---

## CRD Changes

### MinecraftServerSpec (New Fields)

```yaml
spec:
  sftpUsername: "mc-my-server"    # Auto-generated
  sftpPassword: "Xa9K2mP7nQ4vL8tY" # Auto-generated
  memory: "4Gi"                   # Required
  storageSize: "5Gi"
```

### MinecraftServerStatus (New Fields)

```yaml
status:
  sftpUsername: "mc-my-server"    # Copied from spec
  sftpPassword: "Xa9K2mP7nQ4vL8tY" # Copied from spec
  allocatedMemory: "4Gi"          # Set by controller
```

---

## Testing Examples

### 1. Create Server with Memory

```bash
curl -X POST http://localhost:8080/api/v1/servers \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-server",
    "eula": true,
    "memory": "2Gi"
  }'
```

**Success Response:**
```json
{
  "name": "test-server",
  "memory": "2Gi",
  "sftpUsername": "mc-test-server",
  "sftpPassword": "Xa9K2mP7nQ4vL8tY",
  "phase": "Pending",
  ...
}
```

### 2. Check Cluster Resources

```bash
curl http://localhost:8080/api/v1/cluster/resources
```

**Response:**
```json
{
  "totalMemory": "16.0 GiB",
  "allocatedMemory": "8.5 GiB",
  "availableMemory": "7.5 GiB",
  "totalNodes": 2,
  "nodes": [...]
}
```

### 3. Insufficient Memory Error

```bash
curl -X POST http://localhost:8080/api/v1/servers \
  -H "Content-Type: application/json" \
  -d '{
    "name": "big-server",
    "eula": true,
    "memory": "100Gi"
  }'
```

**Error Response:**
```json
{
  "error": "insufficient_capacity",
  "message": "insufficient memory: requested 100.0 GiB, available 7.5 GiB"
}
```

---

## Frontend Integration

### Recommended UX Flow

1. **On page load:**
   ```javascript
   fetch('/api/v1/cluster/resources')
     .then(r => r.json())
     .then(data => {
       // Show user: "Available RAM: 7.5 GiB"
       // Populate dropdown: ["512Mi", "1Gi", "2Gi", "4Gi", "8Gi"]
       // Disable options > availableMemory
     });
   ```

2. **On form submit:**
   ```javascript
   fetch('/api/v1/servers', {
     method: 'POST',
     body: JSON.stringify({
       name: "my-server",
       eula: true,
       memory: selectedMemory  // from dropdown
     })
   })
   .then(r => r.json())
   .then(server => {
     // Show user their SFTP credentials:
     // Username: server.sftpUsername
     // Password: server.sftpPassword
     // WARNING: Save these! They won't be shown again
   });
   ```

3. **Display SFTP credentials:**
   ```
   ╔═══════════════════════════════════════╗
   ║ Server Created Successfully!          ║
   ║                                       ║
   ║ SFTP Access:                          ║
   ║ Host:     mc-my-server.example.com    ║
   ║ Username: mc-my-server                ║
   ║ Password: Xa9K2mP7nQ4vL8tY            ║
   ║                                       ║
   ║ ⚠️  Save these credentials!           ║
   ║     You won't see them again          ║
   ╚═══════════════════════════════════════╝
   ```

---

## Migration Notes

### Breaking Changes ⚠️

1. **`sftpUsers` field removed** - No longer accepted in requests
2. **`memory` field now required** - All create requests must specify RAM

### If you have existing MinecraftServer resources:

The CRD is backward compatible, but old resources won't have:
- `sftpUsername` / `sftpPassword` (will be empty)
- `memory` (will be empty - controller should handle gracefully)

Recommended: Delete and recreate servers after updating the CRD.

---

## Security Considerations

### SFTP Credentials

- Passwords are **not hashed** in the CRD (stored in plaintext in etcd)
- This is intentional: controller needs plaintext to configure SFTP container
- **Alternative:** Use Kubernetes Secrets (future improvement)

**Recommendation:**
```yaml
# Future: Store credentials in a Secret
apiVersion: v1
kind: Secret
metadata:
  name: mc-my-server-sftp
type: Opaque
stringData:
  username: mc-my-server
  password: Xa9K2mP7nQ4vL8tY
```

Then reference in the MinecraftServer:
```yaml
spec:
  sftpCredentialsSecretRef:
    name: mc-my-server-sftp
```

---

## Files Changed

```
backend/
├── pkg/
│   ├── apis/homecraft/v1alpha1/
│   │   ├── types.go                      # Updated: Added memory, sftp fields
│   │   └── zz_generated.deepcopy.go      # Updated: Removed SFTPUsers copy logic
│   ├── handlers/
│   │   └── server_handler.go             # Updated: Auto-gen SFTP, capacity check, new endpoint
│   ├── k8s/
│   │   └── client.go                     # Updated: Added capacity checking methods
│   ├── models/
│   │   └── request.go                    # Updated: Removed sftpUsers, added memory & ClusterResourcesResponse
│   └── utils/
│       └── credentials.go                # NEW: SFTP credential generation
├── cmd/api/
│   └── main.go                           # Updated: Added /cluster/resources route
├── config/crd/
│   └── minecraftserver-crd.yaml          # Updated: New fields in spec/status
└── examples/
    └── create-server-request.json        # Updated: Removed sftpUsers, added memory
```

---

## Next Steps

1. **Update the CRD in your cluster:**
   ```bash
   kubectl apply -f config/crd/minecraftserver-crd.yaml
   ```

2. **Rebuild and restart the API:**
   ```bash
   make build && make run
   ```

3. **Update your frontend** to:
   - Remove SFTP user input fields
   - Add RAM selection dropdown
   - Fetch cluster resources on load
   - Display generated SFTP credentials after creation

4. **Update the controller** to:
   - Read `sftpUsername` and `sftpPassword` from spec
   - Use `memory` field for container resource requests
   - Update `status.allocatedMemory` after pod creation

---

## Summary

✅ **SFTP credentials are now auto-generated** - users don't need to provide them  
✅ **RAM allocation is now required** - ensures proper resource management  
✅ **Capacity checking prevents over-allocation** - API rejects requests if insufficient resources  
✅ **New cluster resources endpoint** - frontend can show available RAM to users  

The API is now smarter, safer, and easier to use! 🚀

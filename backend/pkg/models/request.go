package models

// CreateServerRequest represents the request to create a new Minecraft server
type CreateServerRequest struct {
	Name        string `json:"name" binding:"required"`
	EULA        bool   `json:"eula"`
	Memory      string `json:"memory" binding:"required"` // Required: RAM allocation (e.g., "2Gi", "4Gi")
	StorageSize string `json:"storageSize"`
	Version     string `json:"version"`
	ServerType  string `json:"serverType"`
	MaxPlayers  int    `json:"maxPlayers"`
	Difficulty  string `json:"difficulty"`
	Gamemode    string `json:"gamemode"`
}

// ServerResponse represents a Minecraft server in API responses
type ServerResponse struct {
	Name            string `json:"name"`
	Namespace       string `json:"namespace"`
	EULA            bool   `json:"eula"`
	Memory          string `json:"memory"`
	StorageSize     string `json:"storageSize"`
	Version         string `json:"version"`
	ServerType      string `json:"serverType"`
	MaxPlayers      int    `json:"maxPlayers"`
	Difficulty      string `json:"difficulty"`
	Gamemode        string `json:"gamemode"`
	Phase           string `json:"phase,omitempty"`
	Endpoint        string `json:"endpoint,omitempty"`
	SFTPEndpoint    string `json:"sftpEndpoint,omitempty"`
	SFTPUsername    string `json:"sftpUsername,omitempty"`
	SFTPPassword    string `json:"sftpPassword,omitempty"`
	AllocatedMemory string `json:"allocatedMemory,omitempty"`
	CreatedAt       string `json:"createdAt,omitempty"`
}

// ClusterResourcesResponse represents available cluster resources
type ClusterResourcesResponse struct {
	TotalMemory     string `json:"totalMemory"`     // Total RAM in cluster
	AllocatedMemory string `json:"allocatedMemory"` // RAM used by all servers
	AvailableMemory string `json:"availableMemory"` // RAM available for new servers
	TotalNodes      int    `json:"totalNodes"`      // Number of nodes
	Nodes           []Node `json:"nodes"`           // Per-node resource info
}

// Node represents a single node's resources
type Node struct {
	Name            string `json:"name"`
	TotalMemory     string `json:"totalMemory"`
	AllocatedMemory string `json:"allocatedMemory"`
	AvailableMemory string `json:"availableMemory"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

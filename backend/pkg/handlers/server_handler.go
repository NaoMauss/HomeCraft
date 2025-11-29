package handlers

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/homecraft/backend/pkg/apis/homecraft/v1alpha1"
	"github.com/homecraft/backend/pkg/k8s"
	"github.com/homecraft/backend/pkg/models"
	"github.com/homecraft/backend/pkg/utils"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// MinecraftNamespace is the dedicated namespace for all Minecraft servers
	MinecraftNamespace = "minecraft-servers"
)

// ServerHandler handles HTTP requests for Minecraft servers
type ServerHandler struct {
	k8sClient *k8s.Client
}

// NewServerHandler creates a new ServerHandler
func NewServerHandler(k8sClient *k8s.Client) *ServerHandler {
	return &ServerHandler{
		k8sClient: k8sClient,
	}
}

// HealthCheck handles GET /health
func (h *ServerHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, models.HealthResponse{
		Status:  "healthy",
		Message: "HomeCraft API is running",
	})
}

// CreateServer handles POST /servers
func (h *ServerHandler) CreateServer(c *gin.Context) {
	var req models.CreateServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	// Validate memory format
	if !isValidMemoryFormat(req.Memory) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_memory",
			Message: "Memory must be in format like '2Gi', '4Gi', '512Mi'",
		})
		return
	}

	// Parse requested memory to bytes for capacity check
	requestedMemory, err := parseMemoryToBytes(req.Memory)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_memory",
			Message: fmt.Sprintf("Failed to parse memory: %v", err),
		})
		return
	}

	// Check cluster capacity
	hasCapacity, message, err := h.k8sClient.CheckMemoryAvailability(c.Request.Context(), requestedMemory)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "capacity_check_failed",
			Message: fmt.Sprintf("Failed to check cluster capacity: %v", err),
		})
		return
	}

	if !hasCapacity {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "insufficient_capacity",
			Message: message,
		})
		return
	}

	// Generate SFTP credentials
	sftpUsername, sftpPassword, err := utils.GenerateSFTPCredentials(req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "credential_generation_failed",
			Message: fmt.Sprintf("Failed to generate SFTP credentials: %v", err),
		})
		return
	}

	// Set defaults
	if req.StorageSize == "" {
		req.StorageSize = "1Gi"
	}
	if req.Version == "" {
		req.Version = "LATEST"
	}
	if req.ServerType == "" {
		req.ServerType = "VANILLA"
	}
	if req.MaxPlayers == 0 {
		req.MaxPlayers = 20
	}
	if req.Difficulty == "" {
		req.Difficulty = "normal"
	}
	if req.Gamemode == "" {
		req.Gamemode = "survival"
	}

	// Create MinecraftServer CR
	server := &v1alpha1.MinecraftServer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: MinecraftNamespace,
		},
		Spec: v1alpha1.MinecraftServerSpec{
			EULA:           req.EULA,
			SFTPUsername:   sftpUsername,
			SFTPPassword:   sftpPassword,
			Memory:         req.Memory,
			StorageSize:    req.StorageSize,
			Version:        req.Version,
			ServerType:     req.ServerType,
			MaxPlayers:     req.MaxPlayers,
			Difficulty:     req.Difficulty,
			Gamemode:       req.Gamemode,
			PublicEndpoint: req.PublicEndpoint,
		},
	}

	result, err := h.k8sClient.CreateMinecraftServer(c.Request.Context(), MinecraftNamespace, server)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "creation_failed",
			Message: fmt.Sprintf("Failed to create server: %v", err),
		})
		return
	}

	c.JSON(http.StatusCreated, convertToResponse(result))
}

// ListServers handles GET /servers
func (h *ServerHandler) ListServers(c *gin.Context) {
	list, err := h.k8sClient.ListMinecraftServers(c.Request.Context(), MinecraftNamespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "list_failed",
			Message: fmt.Sprintf("Failed to list servers: %v", err),
		})
		return
	}

	responses := make([]models.ServerResponse, len(list.Items))
	for i, item := range list.Items {
		responses[i] = convertToResponse(&item)
	}

	c.JSON(http.StatusOK, gin.H{
		"items": responses,
		"count": len(responses),
	})
}

// GetServer handles GET /servers/:name
func (h *ServerHandler) GetServer(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Server name is required",
		})
		return
	}

	server, err := h.k8sClient.GetMinecraftServer(c.Request.Context(), MinecraftNamespace, name)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "not_found",
			Message: fmt.Sprintf("Server not found: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, convertToResponse(server))
}

// DeleteServer handles DELETE /servers/:name
func (h *ServerHandler) DeleteServer(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Server name is required",
		})
		return
	}

	err := h.k8sClient.DeleteMinecraftServer(c.Request.Context(), MinecraftNamespace, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "deletion_failed",
			Message: fmt.Sprintf("Failed to delete server: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Server deleted successfully",
		"name":    name,
	})
}

// GetClusterResources handles GET /cluster/resources
func (h *ServerHandler) GetClusterResources(c *gin.Context) {
	total, allocated, available, err := h.k8sClient.GetClusterMemoryResources(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "resource_fetch_failed",
			Message: fmt.Sprintf("Failed to fetch cluster resources: %v", err),
		})
		return
	}

	// Get per-node information
	nodes, err := h.k8sClient.GetClientset().CoreV1().Nodes().List(c.Request.Context(), metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "node_fetch_failed",
			Message: fmt.Sprintf("Failed to fetch node information: %v", err),
		})
		return
	}

	nodeResources := make([]models.Node, 0, len(nodes.Items))
	for _, node := range nodes.Items {
		nodeTotal := int64(0)
		if memory, ok := node.Status.Allocatable["memory"]; ok {
			nodeTotal = memory.Value()
		}

		// Calculate allocated for this node (simplified - just divide evenly)
		nodeAllocated := allocated / int64(len(nodes.Items))
		nodeAvailable := nodeTotal - nodeAllocated

		nodeResources = append(nodeResources, models.Node{
			Name:            node.Name,
			TotalMemory:     bytesToHumanReadable(nodeTotal),
			AllocatedMemory: bytesToHumanReadable(nodeAllocated),
			AvailableMemory: bytesToHumanReadable(nodeAvailable),
		})
	}

	response := models.ClusterResourcesResponse{
		TotalMemory:     bytesToHumanReadable(total),
		AllocatedMemory: bytesToHumanReadable(allocated),
		AvailableMemory: bytesToHumanReadable(available),
		TotalNodes:      len(nodes.Items),
		Nodes:           nodeResources,
	}

	c.JSON(http.StatusOK, response)
}

// Helper functions

func convertToResponse(server *v1alpha1.MinecraftServer) models.ServerResponse {
	// Use publicEndpoint from status if available, otherwise fall back to spec
	publicEndpoint := server.Status.PublicEndpoint
	if publicEndpoint == "" {
		publicEndpoint = server.Spec.PublicEndpoint
	}

	return models.ServerResponse{
		Name:            server.Name,
		Namespace:       server.Namespace,
		EULA:            server.Spec.EULA,
		Memory:          server.Spec.Memory,
		StorageSize:     server.Spec.StorageSize,
		Version:         server.Spec.Version,
		ServerType:      server.Spec.ServerType,
		MaxPlayers:      server.Spec.MaxPlayers,
		Difficulty:      server.Spec.Difficulty,
		Gamemode:        server.Spec.Gamemode,
		Phase:           server.Status.Phase,
		Endpoint:        server.Status.Endpoint,
		PublicEndpoint:  publicEndpoint,
		SFTPEndpoint:    server.Status.SFTPEndpoint,
		SFTPUsername:    server.Status.SFTPUsername,
		SFTPPassword:    server.Status.SFTPPassword,
		AllocatedMemory: server.Status.AllocatedMemory,
		CreatedAt:       server.CreationTimestamp.Format("2006-01-02T15:04:05Z"),
	}
}

func isValidMemoryFormat(memory string) bool {
	// Match patterns like "512Mi", "1Gi", "2Gi", etc.
	matched, _ := regexp.MatchString(`^[0-9]+[MGT]i$`, memory)
	return matched
}

func parseMemoryToBytes(memory string) (int64, error) {
	quantity, err := resource.ParseQuantity(memory)
	if err != nil {
		return 0, err
	}
	return quantity.Value(), nil
}

func bytesToHumanReadable(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// parseHumanReadableToBytes converts human readable memory to bytes (kept for compatibility)
func parseHumanReadableToBytes(memory string) (int64, error) {
	re := regexp.MustCompile(`^(\d+)([KMGT]?)i?B?$`)
	matches := re.FindStringSubmatch(memory)
	if len(matches) < 3 {
		return 0, fmt.Errorf("invalid memory format: %s", memory)
	}

	value, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return 0, err
	}

	multipliers := map[string]int64{
		"":  1,
		"K": 1024,
		"M": 1024 * 1024,
		"G": 1024 * 1024 * 1024,
		"T": 1024 * 1024 * 1024 * 1024,
	}

	multiplier, ok := multipliers[matches[2]]
	if !ok {
		return 0, fmt.Errorf("invalid memory unit: %s", matches[2])
	}

	return value * multiplier, nil
}

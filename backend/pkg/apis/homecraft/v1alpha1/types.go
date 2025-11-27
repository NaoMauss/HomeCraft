package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MinecraftServerSpec defines the desired state of MinecraftServer
type MinecraftServerSpec struct {
	// EULA indicates whether the user accepts the Minecraft EULA
	// +kubebuilder:default=true
	EULA bool `json:"eula"`

	// SFTPUsername is the auto-generated SFTP username for file access
	// +optional
	SFTPUsername string `json:"sftpUsername,omitempty"`

	// SFTPPassword is the auto-generated SFTP password for file access
	// +optional
	SFTPPassword string `json:"sftpPassword,omitempty"`

	// Memory is the amount of RAM allocated to the server (e.g., "2Gi", "4Gi")
	// +kubebuilder:default="2Gi"
	// +kubebuilder:validation:Pattern=`^[0-9]+[MGT]i$`
	Memory string `json:"memory"`

	// StorageSize is the size of the persistent volume claim
	// +kubebuilder:default="1Gi"
	// +kubebuilder:validation:Pattern=`^[0-9]+[MGT]i$`
	StorageSize string `json:"storageSize"`

	// Version is the Minecraft server version (e.g., "1.20.1", "LATEST")
	// +kubebuilder:default="LATEST"
	// +optional
	Version string `json:"version,omitempty"`

	// ServerType is the Minecraft server type (VANILLA, PAPER, FORGE, etc.)
	// +kubebuilder:default="VANILLA"
	// +optional
	ServerType string `json:"serverType,omitempty"`

	// MaxPlayers is the maximum number of players
	// +kubebuilder:default=20
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=1000
	// +optional
	MaxPlayers int `json:"maxPlayers,omitempty"`

	// Difficulty is the game difficulty (peaceful, easy, normal, hard)
	// +kubebuilder:default="normal"
	// +optional
	Difficulty string `json:"difficulty,omitempty"`

	// Gamemode is the default game mode (survival, creative, adventure, spectator)
	// +kubebuilder:default="survival"
	// +optional
	Gamemode string `json:"gamemode,omitempty"`
}

// MinecraftServerStatus defines the observed state of MinecraftServer
type MinecraftServerStatus struct {
	// Phase represents the current phase of the server (Pending, Running, Failed)
	Phase string `json:"phase,omitempty"`

	// Endpoint is the service endpoint to connect to the server
	Endpoint string `json:"endpoint,omitempty"`

	// SFTPEndpoint is the SFTP endpoint for file access
	SFTPEndpoint string `json:"sftpEndpoint,omitempty"`

	// SFTPUsername is the generated SFTP username (populated by controller)
	SFTPUsername string `json:"sftpUsername,omitempty"`

	// SFTPPassword is the generated SFTP password (populated by controller)
	SFTPPassword string `json:"sftpPassword,omitempty"`

	// AllocatedMemory is the actual memory allocated to the server
	AllocatedMemory string `json:"allocatedMemory,omitempty"`

	// LastUpdated is the timestamp of the last status update
	LastUpdated metav1.Time `json:"lastUpdated,omitempty"`

	// Message provides additional information about the current state
	Message string `json:"message,omitempty"`

	// Conditions represent the latest available observations of the server's state
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,shortName=mcs
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Endpoint",type=string,JSONPath=`.status.endpoint`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// MinecraftServer is the Schema for the minecraftservers API
type MinecraftServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MinecraftServerSpec   `json:"spec,omitempty"`
	Status MinecraftServerStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

// MinecraftServerList contains a list of MinecraftServer
type MinecraftServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MinecraftServer `json:"items"`
}

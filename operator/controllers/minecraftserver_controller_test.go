package controllers

import (
	"context"
	"testing"
	"time"

	homecraftv1alpha1 "github.com/homecraft/backend/pkg/apis/homecraft/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestSecretForMinecraftServer(t *testing.T) {
	// Setup scheme
	s := runtime.NewScheme()
	_ = scheme.AddToScheme(s)
	_ = homecraftv1alpha1.AddToScheme(s)

	reconciler := &MinecraftServerReconciler{
		Client: fake.NewClientBuilder().WithScheme(s).Build(),
		Log:    zap.New(zap.UseDevMode(true)),
		Scheme: s,
	}

	minecraftServer := &homecraftv1alpha1.MinecraftServer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-server",
			Namespace: "default",
		},
		Spec: homecraftv1alpha1.MinecraftServerSpec{
			EULA:         true,
			SFTPUsername: "test-user",
			SFTPPassword: "test-password",
			Memory:       "2Gi",
			StorageSize:  "5Gi",
		},
	}

	secret := reconciler.secretForMinecraftServer(minecraftServer)

	// Verify secret metadata
	if secret.Name != "test-server-sftp" {
		t.Errorf("Expected secret name 'test-server-sftp', got %s", secret.Name)
	}
	if secret.Namespace != "default" {
		t.Errorf("Expected namespace 'default', got %s", secret.Namespace)
	}

	// Verify secret data
	if secret.StringData["username"] != "test-user" {
		t.Errorf("Expected username 'test-user', got %s", secret.StringData["username"])
	}
	if secret.StringData["password"] != "test-password" {
		t.Errorf("Expected password 'test-password', got %s", secret.StringData["password"])
	}
}

func TestPVCForMinecraftServer(t *testing.T) {
	s := runtime.NewScheme()
	_ = scheme.AddToScheme(s)
	_ = homecraftv1alpha1.AddToScheme(s)

	reconciler := &MinecraftServerReconciler{
		Client: fake.NewClientBuilder().WithScheme(s).Build(),
		Log:    zap.New(zap.UseDevMode(true)),
		Scheme: s,
	}

	tests := []struct {
		name        string
		server      *homecraftv1alpha1.MinecraftServer
		wantSize    string
		wantPVCName string
	}{
		{
			name: "default storage size",
			server: &homecraftv1alpha1.MinecraftServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-server",
					Namespace: "default",
				},
				Spec: homecraftv1alpha1.MinecraftServerSpec{
					StorageSize: "5Gi",
				},
			},
			wantSize:    "5Gi",
			wantPVCName: "test-server-data",
		},
		{
			name: "custom storage size",
			server: &homecraftv1alpha1.MinecraftServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "custom-server",
					Namespace: "minecraft",
				},
				Spec: homecraftv1alpha1.MinecraftServerSpec{
					StorageSize: "10Gi",
				},
			},
			wantSize:    "10Gi",
			wantPVCName: "custom-server-data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pvc := reconciler.pvcForMinecraftServer(tt.server)

			if pvc.Name != tt.wantPVCName {
				t.Errorf("Expected PVC name %s, got %s", tt.wantPVCName, pvc.Name)
			}

			storageRequest := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
			if storageRequest.String() != tt.wantSize {
				t.Errorf("Expected storage size %s, got %s", tt.wantSize, storageRequest.String())
			}

			if len(pvc.Spec.AccessModes) != 1 || pvc.Spec.AccessModes[0] != corev1.ReadWriteOnce {
				t.Errorf("Expected AccessMode ReadWriteOnce")
			}
		})
	}
}

func TestStatefulSetForMinecraftServer(t *testing.T) {
	s := runtime.NewScheme()
	_ = scheme.AddToScheme(s)
	_ = homecraftv1alpha1.AddToScheme(s)

	reconciler := &MinecraftServerReconciler{
		Client: fake.NewClientBuilder().WithScheme(s).Build(),
		Log:    zap.New(zap.UseDevMode(true)),
		Scheme: s,
	}

	tests := []struct {
		name           string
		server         *homecraftv1alpha1.MinecraftServer
		wantReplicas   int32
		wantMemory     string
		wantVersion    string
		wantType       string
		wantContainers int
		wantMaxPlayers string
		wantDifficulty string
		wantGamemode   string
	}{
		{
			name: "default configuration",
			server: &homecraftv1alpha1.MinecraftServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-server",
					Namespace: "default",
				},
				Spec: homecraftv1alpha1.MinecraftServerSpec{
					EULA:         true,
					SFTPUsername: "test-user",
					SFTPPassword: "test-pass",
					Memory:       "2Gi",
					StorageSize:  "5Gi",
				},
			},
			wantReplicas:   1,
			wantMemory:     "2Gi",
			wantVersion:    "LATEST",
			wantType:       "VANILLA",
			wantContainers: 2,
		},
		{
			name: "custom configuration",
			server: &homecraftv1alpha1.MinecraftServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "custom-server",
					Namespace: "minecraft",
				},
				Spec: homecraftv1alpha1.MinecraftServerSpec{
					EULA:         true,
					SFTPUsername: "custom-user",
					SFTPPassword: "custom-pass",
					Memory:       "4Gi",
					StorageSize:  "10Gi",
					Version:      "1.19.4",
					ServerType:   "PAPER",
					MaxPlayers:   50,
					Difficulty:   "hard",
					Gamemode:     "creative",
				},
			},
			wantReplicas:   1,
			wantMemory:     "4Gi",
			wantVersion:    "1.19.4",
			wantType:       "PAPER",
			wantContainers: 2,
			wantMaxPlayers: "50",
			wantDifficulty: "hard",
			wantGamemode:   "creative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sts := reconciler.statefulSetForMinecraftServer(tt.server)

			// Check replicas
			if *sts.Spec.Replicas != tt.wantReplicas {
				t.Errorf("Expected %d replicas, got %d", tt.wantReplicas, *sts.Spec.Replicas)
			}

			// Check containers
			if len(sts.Spec.Template.Spec.Containers) != tt.wantContainers {
				t.Errorf("Expected %d containers, got %d", tt.wantContainers, len(sts.Spec.Template.Spec.Containers))
			}

			// Check minecraft container
			minecraftContainer := sts.Spec.Template.Spec.Containers[0]
			if minecraftContainer.Name != "minecraft" {
				t.Errorf("Expected container name 'minecraft', got %s", minecraftContainer.Name)
			}

			// Check memory resources
			memoryLimit := minecraftContainer.Resources.Limits[corev1.ResourceMemory]
			if memoryLimit.String() != tt.wantMemory {
				t.Errorf("Expected memory limit %s, got %s", tt.wantMemory, memoryLimit.String())
			}

			// Check environment variables
			envMap := make(map[string]string)
			for _, env := range minecraftContainer.Env {
				envMap[env.Name] = env.Value
			}

			if envMap["VERSION"] != tt.wantVersion {
				t.Errorf("Expected VERSION %s, got %s", tt.wantVersion, envMap["VERSION"])
			}
			if envMap["TYPE"] != tt.wantType {
				t.Errorf("Expected TYPE %s, got %s", tt.wantType, envMap["TYPE"])
			}
			if envMap["MEMORY"] != tt.wantMemory {
				t.Errorf("Expected MEMORY %s, got %s", tt.wantMemory, envMap["MEMORY"])
			}

			// Check optional env vars if set
			if tt.wantMaxPlayers != "" && envMap["MAX_PLAYERS"] != tt.wantMaxPlayers {
				t.Errorf("Expected MAX_PLAYERS %s, got %s", tt.wantMaxPlayers, envMap["MAX_PLAYERS"])
			}
			if tt.wantDifficulty != "" && envMap["DIFFICULTY"] != tt.wantDifficulty {
				t.Errorf("Expected DIFFICULTY %s, got %s", tt.wantDifficulty, envMap["DIFFICULTY"])
			}
			if tt.wantGamemode != "" && envMap["MODE"] != tt.wantGamemode {
				t.Errorf("Expected MODE %s, got %s", tt.wantGamemode, envMap["MODE"])
			}

			// Check SFTP container
			sftpContainer := sts.Spec.Template.Spec.Containers[1]
			if sftpContainer.Name != "sftp" {
				t.Errorf("Expected container name 'sftp', got %s", sftpContainer.Name)
			}

			// Check volumes
			if len(sts.Spec.Template.Spec.Volumes) != 1 {
				t.Errorf("Expected 1 volume, got %d", len(sts.Spec.Template.Spec.Volumes))
			}
		})
	}
}

func TestServiceForMinecraft(t *testing.T) {
	s := runtime.NewScheme()
	_ = scheme.AddToScheme(s)
	_ = homecraftv1alpha1.AddToScheme(s)

	reconciler := &MinecraftServerReconciler{
		Client: fake.NewClientBuilder().WithScheme(s).Build(),
		Log:    zap.New(zap.UseDevMode(true)),
		Scheme: s,
	}

	minecraftServer := &homecraftv1alpha1.MinecraftServer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-server",
			Namespace: "default",
		},
		Spec: homecraftv1alpha1.MinecraftServerSpec{},
	}

	svc := reconciler.serviceForMinecraft(minecraftServer)

	// Verify service metadata
	if svc.Name != "test-server-minecraft" {
		t.Errorf("Expected service name 'test-server-minecraft', got %s", svc.Name)
	}

	// Verify service type
	if svc.Spec.Type != corev1.ServiceTypeNodePort {
		t.Errorf("Expected service type NodePort, got %s", svc.Spec.Type)
	}

	// Verify ports
	if len(svc.Spec.Ports) != 1 {
		t.Fatalf("Expected 1 port, got %d", len(svc.Spec.Ports))
	}
	if svc.Spec.Ports[0].Port != 25565 {
		t.Errorf("Expected port 25565, got %d", svc.Spec.Ports[0].Port)
	}
}

func TestServiceForSFTP(t *testing.T) {
	s := runtime.NewScheme()
	_ = scheme.AddToScheme(s)
	_ = homecraftv1alpha1.AddToScheme(s)

	reconciler := &MinecraftServerReconciler{
		Client: fake.NewClientBuilder().WithScheme(s).Build(),
		Log:    zap.New(zap.UseDevMode(true)),
		Scheme: s,
	}

	minecraftServer := &homecraftv1alpha1.MinecraftServer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-server",
			Namespace: "default",
		},
		Spec: homecraftv1alpha1.MinecraftServerSpec{},
	}

	svc := reconciler.serviceForSFTP(minecraftServer)

	// Verify service metadata
	if svc.Name != "test-server-sftp" {
		t.Errorf("Expected service name 'test-server-sftp', got %s", svc.Name)
	}

	// Verify service type
	if svc.Spec.Type != corev1.ServiceTypeNodePort {
		t.Errorf("Expected service type NodePort, got %s", svc.Spec.Type)
	}

	// Verify ports
	if len(svc.Spec.Ports) != 1 {
		t.Fatalf("Expected 1 port, got %d", len(svc.Spec.Ports))
	}
	if svc.Spec.Ports[0].Port != 22 {
		t.Errorf("Expected port 22, got %d", svc.Spec.Ports[0].Port)
	}
}

func TestReconcile_CreateResources(t *testing.T) {
	// Setup scheme
	s := runtime.NewScheme()
	_ = scheme.AddToScheme(s)
	_ = homecraftv1alpha1.AddToScheme(s)

	// Create test MinecraftServer
	minecraftServer := &homecraftv1alpha1.MinecraftServer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-server",
			Namespace: "default",
		},
		Spec: homecraftv1alpha1.MinecraftServerSpec{
			EULA:         true,
			SFTPUsername: "test-user",
			SFTPPassword: "test-pass",
			Memory:       "2Gi",
			StorageSize:  "5Gi",
		},
	}

	// Create fake client with the MinecraftServer
	fakeClient := fake.NewClientBuilder().
		WithScheme(s).
		WithObjects(minecraftServer).
		WithStatusSubresource(minecraftServer).
		Build()

	reconciler := &MinecraftServerReconciler{
		Client: fakeClient,
		Log:    zap.New(zap.UseDevMode(true)),
		Scheme: s,
	}

	// Reconcile
	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-server",
			Namespace: "default",
		},
	}

	ctx := context.Background()
	result, err := reconciler.Reconcile(ctx, req)

	if err != nil {
		t.Fatalf("Reconcile failed: %v", err)
	}

	if result.RequeueAfter != 30*time.Second {
		t.Errorf("Expected requeue after 30s, got %v", result.RequeueAfter)
	}

	// Verify Secret was created
	secret := &corev1.Secret{}
	err = fakeClient.Get(ctx, types.NamespacedName{Name: "test-server-sftp", Namespace: "default"}, secret)
	if err != nil {
		t.Errorf("Failed to get Secret: %v", err)
	}

	// Verify PVC was created
	pvc := &corev1.PersistentVolumeClaim{}
	err = fakeClient.Get(ctx, types.NamespacedName{Name: "test-server-data", Namespace: "default"}, pvc)
	if err != nil {
		t.Errorf("Failed to get PVC: %v", err)
	}

	// Verify StatefulSet was created
	sts := &appsv1.StatefulSet{}
	err = fakeClient.Get(ctx, types.NamespacedName{Name: "test-server", Namespace: "default"}, sts)
	if err != nil {
		t.Errorf("Failed to get StatefulSet: %v", err)
	}

	// Verify Minecraft Service was created
	minecraftSvc := &corev1.Service{}
	err = fakeClient.Get(ctx, types.NamespacedName{Name: "test-server-minecraft", Namespace: "default"}, minecraftSvc)
	if err != nil {
		t.Errorf("Failed to get Minecraft Service: %v", err)
	}

	// Verify SFTP Service was created
	sftpSvc := &corev1.Service{}
	err = fakeClient.Get(ctx, types.NamespacedName{Name: "test-server-sftp", Namespace: "default"}, sftpSvc)
	if err != nil {
		t.Errorf("Failed to get SFTP Service: %v", err)
	}
}

func TestReconcile_NotFound(t *testing.T) {
	s := runtime.NewScheme()
	_ = scheme.AddToScheme(s)
	_ = homecraftv1alpha1.AddToScheme(s)

	fakeClient := fake.NewClientBuilder().WithScheme(s).Build()

	reconciler := &MinecraftServerReconciler{
		Client: fakeClient,
		Log:    zap.New(zap.UseDevMode(true)),
		Scheme: s,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "non-existent",
			Namespace: "default",
		},
	}

	result, err := reconciler.Reconcile(context.Background(), req)

	if err != nil {
		t.Errorf("Expected no error for non-existent resource, got %v", err)
	}

	if result.Requeue {
		t.Error("Expected no requeue for non-existent resource")
	}
}

func TestUpdateStatus(t *testing.T) {
	s := runtime.NewScheme()
	_ = scheme.AddToScheme(s)
	_ = homecraftv1alpha1.AddToScheme(s)

	minecraftServer := &homecraftv1alpha1.MinecraftServer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-server",
			Namespace: "default",
		},
		Spec: homecraftv1alpha1.MinecraftServerSpec{
			EULA:         true,
			SFTPUsername: "test-user",
			SFTPPassword: "test-pass",
			Memory:       "2Gi",
		},
	}

	// Create StatefulSet with ready replicas
	readyReplicas := int32(1)
	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-server",
			Namespace: "default",
		},
		Status: appsv1.StatefulSetStatus{
			ReadyReplicas: readyReplicas,
			Replicas:      1,
		},
	}

	// Create services with NodePorts
	minecraftSvc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-server-minecraft",
			Namespace: "default",
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port:     25565,
					NodePort: 30000,
				},
			},
		},
	}

	sftpSvc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-server-sftp",
			Namespace: "default",
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port:     22,
					NodePort: 30001,
				},
			},
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(s).
		WithObjects(minecraftServer, sts, minecraftSvc, sftpSvc).
		WithStatusSubresource(minecraftServer).
		Build()

	reconciler := &MinecraftServerReconciler{
		Client: fakeClient,
		Log:    zap.New(zap.UseDevMode(true)),
		Scheme: s,
	}

	err := reconciler.updateStatus(context.Background(), minecraftServer, sts, minecraftSvc, sftpSvc)
	if err != nil {
		t.Fatalf("updateStatus failed: %v", err)
	}

	// Verify status was updated
	if minecraftServer.Status.Phase != "Running" {
		t.Errorf("Expected phase 'Running', got %s", minecraftServer.Status.Phase)
	}
	if minecraftServer.Status.Endpoint != "<node-ip>:30000" {
		t.Errorf("Expected endpoint '<node-ip>:30000', got %s", minecraftServer.Status.Endpoint)
	}
	if minecraftServer.Status.SFTPEndpoint != "<node-ip>:30001" {
		t.Errorf("Expected SFTP endpoint '<node-ip>:30001', got %s", minecraftServer.Status.SFTPEndpoint)
	}
	if minecraftServer.Status.AllocatedMemory != "2Gi" {
		t.Errorf("Expected allocated memory '2Gi', got %s", minecraftServer.Status.AllocatedMemory)
	}
}

func BenchmarkStatefulSetForMinecraftServer(b *testing.B) {
	s := runtime.NewScheme()
	_ = scheme.AddToScheme(s)
	_ = homecraftv1alpha1.AddToScheme(s)

	reconciler := &MinecraftServerReconciler{
		Client: fake.NewClientBuilder().WithScheme(s).Build(),
		Log:    zap.New(zap.UseDevMode(true)),
		Scheme: s,
	}

	minecraftServer := &homecraftv1alpha1.MinecraftServer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-server",
			Namespace: "default",
		},
		Spec: homecraftv1alpha1.MinecraftServerSpec{
			EULA:         true,
			SFTPUsername: "test-user",
			SFTPPassword: "test-pass",
			Memory:       "2Gi",
			StorageSize:  "5Gi",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = reconciler.statefulSetForMinecraftServer(minecraftServer)
	}
}

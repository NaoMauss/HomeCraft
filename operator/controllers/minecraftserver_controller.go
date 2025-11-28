package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	homecraftv1alpha1 "github.com/homecraft/backend/pkg/apis/homecraft/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	finalizerName = "minecraftserver.homecraft.io/finalizer"
)

// MinecraftServerReconciler reconciles a MinecraftServer object
type MinecraftServerReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=homecraft.io,resources=minecraftservers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=homecraft.io,resources=minecraftservers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=homecraft.io,resources=minecraftservers/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete

func (r *MinecraftServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("minecraftserver", req.NamespacedName)

	// Fetch the MinecraftServer instance
	minecraftServer := &homecraftv1alpha1.MinecraftServer{}
	err := r.Get(ctx, req.NamespacedName, minecraftServer)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("MinecraftServer resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get MinecraftServer")
		return ctrl.Result{}, err
	}

	// Handle deletion
	if !minecraftServer.ObjectMeta.DeletionTimestamp.IsZero() {
		if controllerutil.ContainsFinalizer(minecraftServer, finalizerName) {
			// Cleanup logic here if needed
			log.Info("Cleaning up resources for MinecraftServer")

			controllerutil.RemoveFinalizer(minecraftServer, finalizerName)
			err := r.Update(ctx, minecraftServer)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(minecraftServer, finalizerName) {
		controllerutil.AddFinalizer(minecraftServer, finalizerName)
		err = r.Update(ctx, minecraftServer)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	// Create or update Secret for SFTP credentials
	secret := r.secretForMinecraftServer(minecraftServer)
	if err := r.createOrUpdateResource(ctx, secret, minecraftServer); err != nil {
		return ctrl.Result{}, err
	}

	// Create or update PVC
	pvc := r.pvcForMinecraftServer(minecraftServer)
	if err := r.createOrUpdateResource(ctx, pvc, minecraftServer); err != nil {
		return ctrl.Result{}, err
	}

	// Create or update StatefulSet
	statefulSet := r.statefulSetForMinecraftServer(minecraftServer)
	if err := r.createOrUpdateResource(ctx, statefulSet, minecraftServer); err != nil {
		return ctrl.Result{}, err
	}

	// Create or update Service for Minecraft (game port)
	minecraftSvc := r.serviceForMinecraft(minecraftServer)
	if err := r.createOrUpdateResource(ctx, minecraftSvc, minecraftServer); err != nil {
		return ctrl.Result{}, err
	}

	// Create or update Service for SFTP
	sftpSvc := r.serviceForSFTP(minecraftServer)
	if err := r.createOrUpdateResource(ctx, sftpSvc, minecraftServer); err != nil {
		return ctrl.Result{}, err
	}

	// Update status
	if err := r.updateStatus(ctx, minecraftServer, statefulSet, minecraftSvc, sftpSvc); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

func (r *MinecraftServerReconciler) createOrUpdateResource(ctx context.Context, obj client.Object, owner *homecraftv1alpha1.MinecraftServer) error {
	// Set owner reference for garbage collection
	if err := controllerutil.SetControllerReference(owner, obj, r.Scheme); err != nil {
		return err
	}

	// Try to get the resource
	key := types.NamespacedName{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
	}

	existing := obj.DeepCopyObject().(client.Object)
	err := r.Get(ctx, key, existing)

	if err != nil && errors.IsNotFound(err) {
		// Create the resource
		r.Log.Info("Creating resource", "kind", obj.GetObjectKind().GroupVersionKind().Kind, "name", obj.GetName())
		return r.Create(ctx, obj)
	} else if err != nil {
		return err
	}

	r.Log.Info("Resource already exists", "kind", obj.GetObjectKind().GroupVersionKind().Kind, "name", obj.GetName())
	return nil
}

func (r *MinecraftServerReconciler) secretForMinecraftServer(m *homecraftv1alpha1.MinecraftServer) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name + "-sftp",
			Namespace: m.Namespace,
		},
		StringData: map[string]string{
			"username": m.Spec.SFTPUsername,
			"password": m.Spec.SFTPPassword,
		},
	}
}

func (r *MinecraftServerReconciler) pvcForMinecraftServer(m *homecraftv1alpha1.MinecraftServer) *corev1.PersistentVolumeClaim {
	storageQuantity := resource.MustParse(m.Spec.StorageSize)

	return &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name + "-data",
			Namespace: m.Namespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: storageQuantity,
				},
			},
		},
	}
}

func (r *MinecraftServerReconciler) statefulSetForMinecraftServer(m *homecraftv1alpha1.MinecraftServer) *appsv1.StatefulSet {
	replicas := int32(1)
	memoryQuantity := resource.MustParse(m.Spec.Memory)

	// Default values
	version := m.Spec.Version
	if version == "" {
		version = "LATEST"
	}
	serverType := m.Spec.ServerType
	if serverType == "" {
		serverType = "VANILLA"
	}

	labels := map[string]string{
		"app":                          "minecraft",
		"minecraftserver":              m.Name,
		"app.kubernetes.io/name":       "minecraft",
		"app.kubernetes.io/instance":   m.Name,
		"app.kubernetes.io/managed-by": "homecraft-operator",
	}

	// Minecraft container environment variables
	minecraftEnv := []corev1.EnvVar{
		{Name: "EULA", Value: fmt.Sprintf("%t", m.Spec.EULA)},
		{Name: "VERSION", Value: version},
		{Name: "TYPE", Value: serverType},
		{Name: "MEMORY", Value: m.Spec.Memory},
	}

	if m.Spec.MaxPlayers > 0 {
		minecraftEnv = append(minecraftEnv, corev1.EnvVar{
			Name:  "MAX_PLAYERS",
			Value: fmt.Sprintf("%d", m.Spec.MaxPlayers),
		})
	}
	if m.Spec.Difficulty != "" {
		minecraftEnv = append(minecraftEnv, corev1.EnvVar{
			Name:  "DIFFICULTY",
			Value: m.Spec.Difficulty,
		})
	}
	if m.Spec.Gamemode != "" {
		minecraftEnv = append(minecraftEnv, corev1.EnvVar{
			Name:  "MODE",
			Value: m.Spec.Gamemode,
		})
	}

	// SFTP user format: username:password:uid:gid:dir
	sftpUser := fmt.Sprintf("%s:%s:1000:1000:/data", m.Spec.SFTPUsername, m.Spec.SFTPPassword)

	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas:    &replicas,
			ServiceName: m.Name,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "minecraft",
							Image: "itzg/minecraft-server:latest",
							Ports: []corev1.ContainerPort{
								{
									Name:          "minecraft",
									ContainerPort: 25565,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							Env: minecraftEnv,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "data",
									MountPath: "/data",
								},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceMemory: memoryQuantity,
								},
								Limits: corev1.ResourceList{
									corev1.ResourceMemory: memoryQuantity,
								},
							},
						},
						{
							Name:  "sftp",
							Image: "atmoz/sftp:latest",
							Args:  []string{sftpUser},
							Ports: []corev1.ContainerPort{
								{
									Name:          "sftp",
									ContainerPort: 22,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "data",
									MountPath: "/home/" + m.Spec.SFTPUsername + "/data",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "data",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: m.Name + "-data",
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *MinecraftServerReconciler) serviceForMinecraft(m *homecraftv1alpha1.MinecraftServer) *corev1.Service {
	labels := map[string]string{
		"app":             "minecraft",
		"minecraftserver": m.Name,
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name + "-minecraft",
			Namespace: m.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeLoadBalancer,
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Name:       "minecraft",
					Port:       25565,
					TargetPort: intstr.FromInt(25565),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}
}

func (r *MinecraftServerReconciler) serviceForSFTP(m *homecraftv1alpha1.MinecraftServer) *corev1.Service {
	labels := map[string]string{
		"app":             "minecraft",
		"minecraftserver": m.Name,
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name + "-sftp",
			Namespace: m.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeLoadBalancer,
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Name:       "sftp",
					Port:       22,
					TargetPort: intstr.FromInt(22),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}
}

func (r *MinecraftServerReconciler) updateStatus(ctx context.Context, m *homecraftv1alpha1.MinecraftServer,
	sts *appsv1.StatefulSet, minecraftSvc *corev1.Service, sftpSvc *corev1.Service) error {

	// Get the actual StatefulSet to check status
	actualSts := &appsv1.StatefulSet{}
	err := r.Get(ctx, types.NamespacedName{Name: sts.Name, Namespace: sts.Namespace}, actualSts)
	if err != nil {
		return err
	}

	// Get services to get NodePort information
	actualMinecraftSvc := &corev1.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: minecraftSvc.Name, Namespace: minecraftSvc.Namespace}, actualMinecraftSvc)
	if err != nil {
		return err
	}

	actualSftpSvc := &corev1.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: sftpSvc.Name, Namespace: sftpSvc.Namespace}, actualSftpSvc)
	if err != nil {
		return err
	}

	// Determine phase
	phase := "Pending"
	message := "Creating resources"

	if actualSts.Status.ReadyReplicas > 0 {
		phase = "Running"
		message = "Server is running"
	} else if actualSts.Status.Replicas > 0 {
		phase = "Starting"
		message = "Server is starting"
	}

	// Build endpoints using LoadBalancer IPs
	minecraftEndpoint := ""
	sftpEndpoint := ""

	// Get LoadBalancer IP for Minecraft service
	if len(actualMinecraftSvc.Status.LoadBalancer.Ingress) > 0 {
		lbIP := actualMinecraftSvc.Status.LoadBalancer.Ingress[0].IP
		if lbIP != "" && len(actualMinecraftSvc.Spec.Ports) > 0 {
			minecraftEndpoint = fmt.Sprintf("%s:%d", lbIP, actualMinecraftSvc.Spec.Ports[0].Port)
		}
	}

	// Get LoadBalancer IP for SFTP service
	if len(actualSftpSvc.Status.LoadBalancer.Ingress) > 0 {
		lbIP := actualSftpSvc.Status.LoadBalancer.Ingress[0].IP
		if lbIP != "" && len(actualSftpSvc.Spec.Ports) > 0 {
			sftpEndpoint = fmt.Sprintf("%s:%d", lbIP, actualSftpSvc.Spec.Ports[0].Port)
		}
	}

	// Update status
	m.Status.Phase = phase
	m.Status.Message = message
	m.Status.Endpoint = minecraftEndpoint
	m.Status.SFTPEndpoint = sftpEndpoint
	m.Status.SFTPUsername = m.Spec.SFTPUsername
	m.Status.SFTPPassword = m.Spec.SFTPPassword
	m.Status.AllocatedMemory = m.Spec.Memory
	m.Status.LastUpdated = metav1.Now()

	return r.Status().Update(ctx, m)
}

// SetupWithManager sets up the controller with the Manager.
func (r *MinecraftServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&homecraftv1alpha1.MinecraftServer{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Complete(r)
}

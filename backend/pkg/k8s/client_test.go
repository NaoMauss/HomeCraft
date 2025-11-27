package k8s

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestGetClusterMemoryResources(t *testing.T) {
	tests := []struct {
		name          string
		nodes         []corev1.Node
		pods          []corev1.Pod
		wantTotal     int64
		wantAllocated int64
		wantAvailable int64
		wantErr       bool
	}{
		{
			name: "single node with no pods",
			nodes: []corev1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node-1",
					},
					Status: corev1.NodeStatus{
						Allocatable: corev1.ResourceList{
							"memory": resource.MustParse("8Gi"),
						},
					},
				},
			},
			pods:          []corev1.Pod{},
			wantTotal:     8589934592, // 8 GiB
			wantAllocated: 0,
			wantAvailable: 8589934592,
			wantErr:       false,
		},
		{
			name: "single node with pods",
			nodes: []corev1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node-1",
					},
					Status: corev1.NodeStatus{
						Allocatable: corev1.ResourceList{
							"memory": resource.MustParse("16Gi"),
						},
					},
				},
			},
			pods: []corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod-1",
						Namespace: "default",
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name: "container-1",
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
										"memory": resource.MustParse("2Gi"),
									},
								},
							},
						},
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodRunning,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod-2",
						Namespace: "default",
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name: "container-1",
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
										"memory": resource.MustParse("4Gi"),
									},
								},
							},
						},
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodRunning,
					},
				},
			},
			wantTotal:     17179869184, // 16 GiB
			wantAllocated: 6442450944,  // 6 GiB
			wantAvailable: 10737418240, // 10 GiB
			wantErr:       false,
		},
		{
			name: "multiple nodes with pods",
			nodes: []corev1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node-1",
					},
					Status: corev1.NodeStatus{
						Allocatable: corev1.ResourceList{
							"memory": resource.MustParse("8Gi"),
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node-2",
					},
					Status: corev1.NodeStatus{
						Allocatable: corev1.ResourceList{
							"memory": resource.MustParse("8Gi"),
						},
					},
				},
			},
			pods: []corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod-1",
						Namespace: "default",
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name: "container-1",
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
										"memory": resource.MustParse("4Gi"),
									},
								},
							},
						},
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodRunning,
					},
				},
			},
			wantTotal:     17179869184, // 16 GiB
			wantAllocated: 4294967296,  // 4 GiB
			wantAvailable: 12884901888, // 12 GiB
			wantErr:       false,
		},
		{
			name: "skip completed and failed pods",
			nodes: []corev1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node-1",
					},
					Status: corev1.NodeStatus{
						Allocatable: corev1.ResourceList{
							"memory": resource.MustParse("8Gi"),
						},
					},
				},
			},
			pods: []corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod-running",
						Namespace: "default",
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name: "container-1",
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
										"memory": resource.MustParse("2Gi"),
									},
								},
							},
						},
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodRunning,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod-succeeded",
						Namespace: "default",
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name: "container-1",
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
										"memory": resource.MustParse("4Gi"),
									},
								},
							},
						},
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodSucceeded,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod-failed",
						Namespace: "default",
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name: "container-1",
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
										"memory": resource.MustParse("4Gi"),
									},
								},
							},
						},
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodFailed,
					},
				},
			},
			wantTotal:     8589934592, // 8 GiB
			wantAllocated: 2147483648, // 2 GiB (only running pod)
			wantAvailable: 6442450944, // 6 GiB
			wantErr:       false,
		},
		{
			name: "pods with multiple containers",
			nodes: []corev1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node-1",
					},
					Status: corev1.NodeStatus{
						Allocatable: corev1.ResourceList{
							"memory": resource.MustParse("8Gi"),
						},
					},
				},
			},
			pods: []corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod-1",
						Namespace: "default",
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name: "container-1",
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
										"memory": resource.MustParse("1Gi"),
									},
								},
							},
							{
								Name: "container-2",
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
										"memory": resource.MustParse("2Gi"),
									},
								},
							},
						},
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodRunning,
					},
				},
			},
			wantTotal:     8589934592, // 8 GiB
			wantAllocated: 3221225472, // 3 GiB (1 + 2)
			wantAvailable: 5368709120, // 5 GiB
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fake clientset with nodes and pods
			fakeClientset := fake.NewSimpleClientset()

			// Add nodes
			for _, node := range tt.nodes {
				_, err := fakeClientset.CoreV1().Nodes().Create(context.Background(), &node, metav1.CreateOptions{})
				if err != nil {
					t.Fatalf("failed to create node: %v", err)
				}
			}

			// Add pods
			for _, pod := range tt.pods {
				_, err := fakeClientset.CoreV1().Pods(pod.Namespace).Create(context.Background(), &pod, metav1.CreateOptions{})
				if err != nil {
					t.Fatalf("failed to create pod: %v", err)
				}
			}

			// Create client with fake clientset
			client := &Client{
				clientset: fakeClientset,
			}

			// Test GetClusterMemoryResources
			total, allocated, available, err := client.GetClusterMemoryResources(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("GetClusterMemoryResources() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if total != tt.wantTotal {
				t.Errorf("GetClusterMemoryResources() total = %d (%s), want %d (%s)",
					total, bytesToHumanReadable(total), tt.wantTotal, bytesToHumanReadable(tt.wantTotal))
			}

			if allocated != tt.wantAllocated {
				t.Errorf("GetClusterMemoryResources() allocated = %d (%s), want %d (%s)",
					allocated, bytesToHumanReadable(allocated), tt.wantAllocated, bytesToHumanReadable(tt.wantAllocated))
			}

			if available != tt.wantAvailable {
				t.Errorf("GetClusterMemoryResources() available = %d (%s), want %d (%s)",
					available, bytesToHumanReadable(available), tt.wantAvailable, bytesToHumanReadable(tt.wantAvailable))
			}
		})
	}
}

func TestCheckMemoryAvailability(t *testing.T) {
	tests := []struct {
		name            string
		nodes           []corev1.Node
		pods            []corev1.Pod
		requestedMemory int64
		wantAvailable   bool
		wantMessage     string
		wantErr         bool
	}{
		{
			name: "sufficient memory available",
			nodes: []corev1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node-1",
					},
					Status: corev1.NodeStatus{
						Allocatable: corev1.ResourceList{
							"memory": resource.MustParse("16Gi"),
						},
					},
				},
			},
			pods: []corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod-1",
						Namespace: "default",
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name: "container-1",
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
										"memory": resource.MustParse("4Gi"),
									},
								},
							},
						},
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodRunning,
					},
				},
			},
			requestedMemory: 2147483648, // 2 GiB
			wantAvailable:   true,
			wantMessage:     "",
			wantErr:         false,
		},
		{
			name: "insufficient memory available",
			nodes: []corev1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node-1",
					},
					Status: corev1.NodeStatus{
						Allocatable: corev1.ResourceList{
							"memory": resource.MustParse("8Gi"),
						},
					},
				},
			},
			pods: []corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod-1",
						Namespace: "default",
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name: "container-1",
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
										"memory": resource.MustParse("6Gi"),
									},
								},
							},
						},
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodRunning,
					},
				},
			},
			requestedMemory: 4294967296, // 4 GiB (only 2 GiB available)
			wantAvailable:   false,
			wantMessage:     "insufficient memory: requested 4.0 GiB, available 2.0 GiB",
			wantErr:         false,
		},
		{
			name: "exact memory available",
			nodes: []corev1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node-1",
					},
					Status: corev1.NodeStatus{
						Allocatable: corev1.ResourceList{
							"memory": resource.MustParse("8Gi"),
						},
					},
				},
			},
			pods:            []corev1.Pod{},
			requestedMemory: 8589934592, // 8 GiB (exact match)
			wantAvailable:   true,
			wantMessage:     "",
			wantErr:         false,
		},
		{
			name: "one byte over available",
			nodes: []corev1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node-1",
					},
					Status: corev1.NodeStatus{
						Allocatable: corev1.ResourceList{
							"memory": resource.MustParse("8Gi"),
						},
					},
				},
			},
			pods:            []corev1.Pod{},
			requestedMemory: 8589934593, // 8 GiB + 1 byte
			wantAvailable:   false,
			wantMessage:     "insufficient memory: requested 8.0 GiB, available 8.0 GiB",
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fake clientset with nodes and pods
			fakeClientset := fake.NewSimpleClientset()

			// Add nodes
			for _, node := range tt.nodes {
				_, err := fakeClientset.CoreV1().Nodes().Create(context.Background(), &node, metav1.CreateOptions{})
				if err != nil {
					t.Fatalf("failed to create node: %v", err)
				}
			}

			// Add pods
			for _, pod := range tt.pods {
				_, err := fakeClientset.CoreV1().Pods(pod.Namespace).Create(context.Background(), &pod, metav1.CreateOptions{})
				if err != nil {
					t.Fatalf("failed to create pod: %v", err)
				}
			}

			// Create client with fake clientset
			client := &Client{
				clientset: fakeClientset,
			}

			// Test CheckMemoryAvailability
			available, message, err := client.CheckMemoryAvailability(context.Background(), tt.requestedMemory)

			if (err != nil) != tt.wantErr {
				t.Errorf("CheckMemoryAvailability() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if available != tt.wantAvailable {
				t.Errorf("CheckMemoryAvailability() available = %v, want %v", available, tt.wantAvailable)
			}

			if message != tt.wantMessage {
				t.Errorf("CheckMemoryAvailability() message = %q, want %q", message, tt.wantMessage)
			}
		})
	}
}

func TestBytesToHumanReadable(t *testing.T) {
	tests := []struct {
		name  string
		bytes int64
		want  string
	}{
		{
			name:  "zero bytes",
			bytes: 0,
			want:  "0 B",
		},
		{
			name:  "bytes",
			bytes: 512,
			want:  "512 B",
		},
		{
			name:  "1 KiB",
			bytes: 1024,
			want:  "1.0 KiB",
		},
		{
			name:  "1 MiB",
			bytes: 1048576, // 1024 * 1024
			want:  "1.0 MiB",
		},
		{
			name:  "512 MiB",
			bytes: 536870912, // 512 * 1024 * 1024
			want:  "512.0 MiB",
		},
		{
			name:  "1 GiB",
			bytes: 1073741824, // 1 * 1024 * 1024 * 1024
			want:  "1.0 GiB",
		},
		{
			name:  "4 GiB",
			bytes: 4294967296, // 4 * 1024 * 1024 * 1024
			want:  "4.0 GiB",
		},
		{
			name:  "1 TiB",
			bytes: 1099511627776, // 1 * 1024 * 1024 * 1024 * 1024
			want:  "1.0 TiB",
		},
		{
			name:  "1.5 GiB",
			bytes: 1610612736, // 1.5 * 1024 * 1024 * 1024
			want:  "1.5 GiB",
		},
		{
			name:  "2.25 GiB",
			bytes: 2415919104, // 2.25 * 1024 * 1024 * 1024
			want:  "2.2 GiB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := bytesToHumanReadable(tt.bytes)
			if got != tt.want {
				t.Errorf("bytesToHumanReadable(%d) = %q, want %q", tt.bytes, got, tt.want)
			}
		})
	}
}

func TestBytesToHumanReadable_Boundaries(t *testing.T) {
	// Test boundary values
	boundaries := []struct {
		name  string
		bytes int64
	}{
		{"1023 bytes", 1023},
		{"1024 bytes (1 KiB)", 1024},
		{"1 MiB - 1", 1048575},
		{"1 MiB", 1048576},
		{"1 GiB - 1", 1073741823},
		{"1 GiB", 1073741824},
	}

	for _, tt := range boundaries {
		t.Run(tt.name, func(t *testing.T) {
			result := bytesToHumanReadable(tt.bytes)
			// Just verify it doesn't panic and returns a string
			if result == "" {
				t.Errorf("bytesToHumanReadable(%d) returned empty string", tt.bytes)
			}
		})
	}
}

func BenchmarkBytesToHumanReadable(b *testing.B) {
	testBytes := int64(4294967296) // 4 GiB
	for i := 0; i < b.N; i++ {
		_ = bytesToHumanReadable(testBytes)
	}
}

func BenchmarkBytesToHumanReadable_Small(b *testing.B) {
	testBytes := int64(512)
	for i := 0; i < b.N; i++ {
		_ = bytesToHumanReadable(testBytes)
	}
}

func BenchmarkBytesToHumanReadable_Large(b *testing.B) {
	testBytes := int64(1099511627776) // 1 TiB
	for i := 0; i < b.N; i++ {
		_ = bytesToHumanReadable(testBytes)
	}
}

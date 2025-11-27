package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/homecraft/backend/pkg/models"
)

func TestIsValidMemoryFormat(t *testing.T) {
	tests := []struct {
		name   string
		memory string
		want   bool
	}{
		{
			name:   "valid Mi format",
			memory: "512Mi",
			want:   true,
		},
		{
			name:   "valid Gi format",
			memory: "4Gi",
			want:   true,
		},
		{
			name:   "valid Ti format",
			memory: "2Ti",
			want:   true,
		},
		{
			name:   "invalid - lowercase",
			memory: "4gi",
			want:   false,
		},
		{
			name:   "invalid - no unit",
			memory: "4",
			want:   false,
		},
		{
			name:   "invalid - wrong unit",
			memory: "4GB",
			want:   false,
		},
		{
			name:   "invalid - decimal",
			memory: "4.5Gi",
			want:   false,
		},
		{
			name:   "invalid - negative",
			memory: "-4Gi",
			want:   false,
		},
		{
			name:   "valid - large number",
			memory: "1024Mi",
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidMemoryFormat(tt.memory); got != tt.want {
				t.Errorf("isValidMemoryFormat(%q) = %v, want %v", tt.memory, got, tt.want)
			}
		})
	}
}

func TestParseMemoryToBytes(t *testing.T) {
	tests := []struct {
		name    string
		memory  string
		want    int64
		wantErr bool
	}{
		{
			name:    "512Mi",
			memory:  "512Mi",
			want:    536870912, // 512 * 1024 * 1024
			wantErr: false,
		},
		{
			name:    "1Gi",
			memory:  "1Gi",
			want:    1073741824, // 1 * 1024 * 1024 * 1024
			wantErr: false,
		},
		{
			name:    "2Gi",
			memory:  "2Gi",
			want:    2147483648, // 2 * 1024 * 1024 * 1024
			wantErr: false,
		},
		{
			name:    "4Gi",
			memory:  "4Gi",
			want:    4294967296, // 4 * 1024 * 1024 * 1024
			wantErr: false,
		},
		{
			name:    "1Ti",
			memory:  "1Ti",
			want:    1099511627776, // 1 * 1024 * 1024 * 1024 * 1024
			wantErr: false,
		},
		{
			name:    "invalid format",
			memory:  "invalid",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseMemoryToBytes(tt.memory)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseMemoryToBytes(%q) error = %v, wantErr %v", tt.memory, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseMemoryToBytes(%q) = %v, want %v", tt.memory, got, tt.want)
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
			name:  "bytes",
			bytes: 512,
			want:  "512 B",
		},
		{
			name:  "kilobytes",
			bytes: 2048,
			want:  "2.0 KiB",
		},
		{
			name:  "megabytes",
			bytes: 536870912, // 512 * 1024 * 1024
			want:  "512.0 MiB",
		},
		{
			name:  "gigabytes",
			bytes: 4294967296, // 4 * 1024 * 1024 * 1024
			want:  "4.0 GiB",
		},
		{
			name:  "terabytes",
			bytes: 1099511627776, // 1 * 1024 * 1024 * 1024 * 1024
			want:  "1.0 TiB",
		},
		{
			name:  "zero",
			bytes: 0,
			want:  "0 B",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := bytesToHumanReadable(tt.bytes); got != tt.want {
				t.Errorf("bytesToHumanReadable(%d) = %v, want %v", tt.bytes, got, tt.want)
			}
		})
	}
}

func TestHealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a test router
	router := gin.New()
	handler := &ServerHandler{}
	router.GET("/health", handler.HealthCheck)

	// Create a test request
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("HealthCheck() status = %v, want %v", w.Code, http.StatusOK)
	}

	// Parse response
	var response models.HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.Status != "healthy" {
		t.Errorf("HealthCheck() status = %v, want 'healthy'", response.Status)
	}
}

func TestCreateServer_InvalidMemoryFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a test router
	router := gin.New()
	handler := &ServerHandler{}
	router.POST("/servers", handler.CreateServer)

	tests := []struct {
		name           string
		requestBody    models.CreateServerRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "invalid memory - no unit",
			requestBody: models.CreateServerRequest{
				Name:   "test-server",
				EULA:   true,
				Memory: "4",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid_memory",
		},
		{
			name: "invalid memory - wrong unit",
			requestBody: models.CreateServerRequest{
				Name:   "test-server",
				EULA:   true,
				Memory: "4GB",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid_memory",
		},
		{
			name: "invalid memory - lowercase",
			requestBody: models.CreateServerRequest{
				Name:   "test-server",
				EULA:   true,
				Memory: "4gi",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid_memory",
		},
		{
			name: "missing memory",
			requestBody: models.CreateServerRequest{
				Name: "test-server",
				EULA: true,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid_request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/servers", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Perform the request
			router.ServeHTTP(w, req)

			// Check response
			if w.Code != tt.expectedStatus {
				t.Errorf("CreateServer() status = %v, want %v", w.Code, tt.expectedStatus)
			}

			// Parse error response
			var response models.ErrorResponse
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to parse error response: %v", err)
			}

			if response.Error != tt.expectedError {
				t.Errorf("CreateServer() error = %v, want %v", response.Error, tt.expectedError)
			}
		})
	}
}

func BenchmarkParseMemoryToBytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = parseMemoryToBytes("4Gi")
	}
}

func BenchmarkBytesToHumanReadable(b *testing.B) {
	bytes := int64(4294967296) // 4Gi
	for i := 0; i < b.N; i++ {
		_ = bytesToHumanReadable(bytes)
	}
}

# Build stage
FROM golang:1.25.4-alpine AS builder

WORKDIR /workspace

# Copy go mod files for both operator and backend (needed for shared types)
COPY operator/go.mod operator/go.sum ./
COPY backend/go.mod backend/go.sum ../backend/

# Download dependencies
RUN go mod download

# Copy the backend API types (shared dependency)
COPY backend/pkg/apis/ ../backend/pkg/apis/

# Copy operator source code
COPY operator/cmd/ cmd/
COPY operator/controllers/ controllers/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager cmd/main.go

# Final stage
FROM gcr.io/distroless/static:nonroot

WORKDIR /

COPY --from=builder /workspace/manager .

USER 65532:65532

ENTRYPOINT ["/manager"]

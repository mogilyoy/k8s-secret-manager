# Build the manager and server binary
FROM golang:1.24 AS builder

WORKDIR /workspace

# Copy go mod and sum files
COPY go.mod go.sum ./
# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are unchanged
RUN go mod download

# Copy the source code
COPY . .

# Build the manager (Controller) binary
# Ensure CGO is disabled for static linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -o manager cmd/controller/main.go

# Build the server (REST API) binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -o server cmd/server/main.go

# Final stage: Use a minimal base image
FROM gcr.io/distroless/static:nonroot

# Copy the license file from the build stage (best practice)
# COPY --from=builder /workspace/hack/boilerplate.go.txt /licenses/

# Copy the built binaries
COPY --from=builder /workspace/manager /manager
COPY --from=builder /workspace/server /server

# Set the entrypoint to the controller (default for kubebuilder deployment)
USER 65532:65532
ENTRYPOINT ["/manager"]
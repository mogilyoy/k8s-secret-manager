# ---------- build stage ----------
FROM golang:1.24 AS builder

WORKDIR /workspace

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# controller
RUN CGO_ENABLED=0 GOOS=linux go build -a -o manager ./cmd/controller/main.go
# api server
RUN CGO_ENABLED=0 GOOS=linux go build -a -o server ./cmd/server/main.go

# ---------- controller image ----------
FROM gcr.io/distroless/static:nonroot AS controller

COPY --from=builder /workspace/manager /manager

USER 65532:65532
ENTRYPOINT ["/manager"]

# ---------- api server image ----------
FROM gcr.io/distroless/static:nonroot AS api-server

COPY --from=builder /workspace/server /server

USER 65532:65532
ENTRYPOINT ["/server"]
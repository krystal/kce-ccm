# syntax = docker/dockerfile:1.0-experimental
# Build the manager binary
FROM golang:1.16 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.sum Makefile ./

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN make deps

# Copy the go source
COPY . .
# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o cloud-controller-manager ./cmd/cloud-controller-manager

# Create a final image
FROM alpine:3.13.5

WORKDIR /
RUN addgroup --gid 1000 -S cloud-controller-manager && adduser -S cloud-controller-manager -G cloud-controller-manager --uid 1000

COPY --from=builder /workspace/cloud-controller-manager .

USER cloud-controller-manager:cloud-controller-manager
ENTRYPOINT ["/cloud-controller-manager"]

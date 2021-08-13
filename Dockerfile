# Build the manager binary
# See Red Hat catalog: https://catalog.redhat.com/software/containers/ubi8/go-toolset/5ce8713aac3db925c03774d1?container-tabs=overview
FROM registry.access.redhat.com/ubi8/go-toolset:1.15.14 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/
COPY pkg/ pkg/

USER root
# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o manager main.go

FROM registry.access.redhat.com/ubi8/ubi-minimal:latest
WORKDIR /
COPY --from=builder /workspace/manager .
RUN chown -R 1001 manager
USER 1001

ENTRYPOINT ["/manager"]

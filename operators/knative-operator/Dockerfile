# Build the manager binary
FROM golang:1.10.3 as builder

# Copy in the go src
WORKDIR /go/src/sigs.k8s.io/cluster-api-provider-gcp/operators/knative-operator
COPY pkg/    pkg/
COPY cmd/    cmd/
COPY vendor/ vendor/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager sigs.k8s.io/cluster-api-provider-gcp/operators/knative-operator/cmd/manager

# Copy the controller-manager into a thin image
FROM ubuntu:latest
WORKDIR /
ADD https://storage.googleapis.com/kubernetes-release/release/v1.12.0/bin/linux/amd64/kubectl /bin/kubectl
RUN chmod a+x /bin/kubectl
COPY --from=builder /go/src/sigs.k8s.io/cluster-api-provider-gcp/operators/knative-operator/manager .
COPY channels/ /channels/
ENTRYPOINT ["/manager"]

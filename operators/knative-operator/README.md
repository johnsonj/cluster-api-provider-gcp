# knative-operator

Proof of concept addon operator for deploying knative

# Installation

Prerequisite: istio, see [istio.yaml](istio.yaml) for an example

```bash
kubectl apply -f config/crds
kubectl create namespace knative-serving
kubectl create namespace knative-build
kubectl create namespace knative-monitoring
kubectl label namespace knative-serving istio-injection=enabled
kubectl apply -f config/samples/addons_v1alpha1_knative.yaml
```

# View Status
```bash
kubectl -n kube-serving get pods
kubectl -n kube-serving get kubeserving -oyaml

kubectl -n kube-build get pods
kubectl -n kube-build get kubebuild -oyaml
```

# Update version
```bash
kubectl -n kube-system edit knatives

# update 'channel' to 'alpha'

# observe changes to v0.2.2
kubectl -n kube-build get kubebuild -oyaml

kubectl -n kube-serving get pods
kubectl -n kube-serving get kubeserving -oyaml
```
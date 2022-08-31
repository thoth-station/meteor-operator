# Meteor Operator

## Prerequisites

- Openshift Pipelines

## Custom Resources

### Meteor

Meteor represents a repository that is built and deployed by this Operator.

```yaml
apiVersion: meteor.zone/v1alpha1
kind: Meteor
metadata:
  name: demo
spec:
  url: github.com/aicoe-aiops/meteor-demo
  ref: main
  ttl: 100000 # Time to live in seconds, defaults to 24h
```

## Run operator locally

### Interactive debugging

To debug/run the operator locally, while it's still connected to a cluster and listens to events from this cluster use these steps.

#### Setup

Prerequisites: `dlv` and `VSCode`

Install `dlv` via `go get -u github.com/go-delve/delve/cmd/dlv`

#### Launch

1. Log in to your cluster via `oc login`
2. Install CRDs via `make install`
3. Start VSCode debugging session by selecting `Meteor Operator` profile

### Run from local machine without debugging

1. Log in to your cluster via `oc login`
2. Run `make install run`

## Quick deploy via operator-sdk

Run following commands. Operator will be deployed to `aicoe-meteor` namespace

```sh
podman login ...

make docker-build

podman tag controller:latest quay.io/<your_account>/meteor-operator:latest
podman push quay.io/<your_account>/meteor-operator:latest

make deploy

kustomize build config/dev | oc apply -f -
```

## Testing locally with `envtest`

see https://book.kubebuilder.io/reference/envtest.html for details, extract:

```sh
export K8S_VERSION=1.21.2
curl -sSLo envtest-bins.tar.gz "https://go.kubebuilder.io/test-tools/${K8S_VERSION}/$(go env GOOS)/$(go env GOARCH)"
sudo mkdir /usr/local/kubebuilder
sudo chown $(whoami) /usr/local/kubebuilder
tar -C /usr/local/kubebuilder --strip-components=1 -zvxf envtest-bins.tar.gz
make test SKIP_FETCH_TOOLS=1 KUBEBUILDER_ASSETS=/usr/local/kubebuilder ENABLE_WEBHOOKS=false

```

## Deploying a local cluster with `kind`

The following steps will set up a local Kubernetes cluster for testing, using [kind](https://kind.sigs.k8s.io/):

```sh
kind create cluster --config hack/kind-config.yaml
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.8.0/cert-manager.yaml
kubectl apply -f https://raw.githubusercontent.com/kubernetes/dashboard/v2.5.0/aio/deploy/recommended.yaml
kubectl apply -f hack/dashboard-adminuser.yaml
kubectl apply -f https://storage.googleapis.com/tekton-releases/pipeline/latest/release.yaml
kubectl apply -f https://github.com/tektoncd/dashboard/releases/latest/download/tekton-dashboard-release.yaml
curl -s https://api.hub.tekton.dev/v1/resource/tekton/task/openshift-client/0.2/raw | sed -e s/Task/ClusterTask/ | kubectl apply -f -

kubectl -n tekton-pipelines port-forward svc/tekton-dashboard 9097:9097

export T=$(kubectl -n kubernetes-dashboard create token admin-user)
```

Run `kubectl proxy` to expose 8001 and access http://localhost:8001/api/v1/namespaces/kubernetes-dashboard/services/https:kubernetes-dashboard:/proxy/ for the kubernetes dashboard.

Use `kubectl port-forward -n tekton-pipelines service/tekton-dashboard 9097:9097` to expose the tekton dashboard, visit it at http://localhost:9097/

## Testing the operator against an existing cluster

Pre-requisite: a Kubernetes or OpenShift cluster, with the local `KUBECONFIG` configured to access it in the preferred target namespace.

Deploy the tekton pipelines and tasks CNBi Operator depends on:

`cat hack/cnbi-*.yaml | kubectl apply -f -`

`make install` will deploy all our CRD, and `make run` to run the controller locally but connected to the cluster.

## Known issues

- Webhooks are currently disabled due to certificate issues

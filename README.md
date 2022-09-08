# Meteor Operator

## Prerequisites

The cluster where the operator will be working on must have these components already deployed and available:

- Tekton / Openshift Pipelines
- [cert-manager](https://github.com/cert-manager/cert-manager)

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

# Development

General pre-requisites:

1. A Kubernetes or OpenShift cluster, with the local `KUBECONFIG` configured to access it in the preferred target namespace.

   - for OpenShift: `oc login`, and use `oc project XXX` to switch to the `XXX` target namespace
   - for a quick local Kubernetes cluster deployment, see the `kind` instructions below

2. The cluster must meet the prerequisites mentioned above (tekton, etc)

3. To deploy the tekton pipelines and tasks CNBi Operator depends on: `make install-pipelines`

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

## Quick build and deploy to a cluster

The following commands will build the image, push it to the container registry, and deploy the operator to the `meteor-system` namespace of your currently configured cluster (as per your `$KUBECONFIG`):

If you want to use a custom image name/version:

```sh
export IMAGE_TAG_BASE=quay.io/myorg/meteor-operator
export VERSION=0.0.7
```

To build, push and deploy:

```sh
podman login ...  # to make sure you can push the image to the public registry

make docker-build
make docker-push
make deploy
```

## Testing locally with `envtest`

see <https://book.kubebuilder.io/reference/envtest.html> for details, extract:

```sh
export K8S_VERSION=1.21.2
curl -sSLo envtest-bins.tar.gz "https://go.kubebuilder.io/test-tools/${K8S_VERSION}/$(go env GOOS)/$(go env GOARCH)"
sudo mkdir /usr/local/kubebuilder
sudo chown $(whoami) /usr/local/kubebuilder
tar -C /usr/local/kubebuilder --strip-components=1 -zvxf envtest-bins.tar.gz
make test SKIP_FETCH_TOOLS=1 KUBEBUILDER_ASSETS=/usr/local/kubebuilder ENABLE_WEBHOOKS=false

```

## Deploying a local cluster with `kind`

Using `make kind-create` will set up a local Kubernetes cluster for testing, using [kind](https://kind.sigs.k8s.io/).
`make kind-load-img` will build and load the operator container image into the cluster.

`export T=$(kubectl -n kubernetes-dashboard create token admin-user)` will get the admin user token for the
Kubernetes dashboard.

Run `kubectl proxy` to expose 8001 and access <http://localhost:8001/api/v1/namespaces/kubernetes-dashboard/services/https:kubernetes-dashboard:/proxy/> for the kubernetes dashboard.

Use `kubectl port-forward -n tekton-pipelines service/tekton-dashboard 9097:9097` to expose the tekton dashboard, visit it at <http://localhost:9097/>

## Testing the operator against an existing cluster

1. `make install` will deploy all our CRD
2. `make run` will run the controller locally but connected to the cluster.

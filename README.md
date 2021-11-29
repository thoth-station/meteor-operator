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

## Known issues

- Webhooks are currently disabled due to certificate issues

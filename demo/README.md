# Demos for Meteor operator

## Custom Notebook Image (CNBi) build of a git repo

The [cnbi-build-demo.sh](./cnbi-build-demo.sh) script uses
[demo-magic.sh](./demo-magic.sh) to present a demo of a `CustomNBImage` of type
`GitRepository`: [op1st-ds.yaml](./op1st-ds.yaml)

Pre-requisites and dependencies:
- `oc` and `tkn` clients
- An OpenShift cluster already configured
- meteor-operator with CNBi support already deployed in the cluster
- KUBECONFIG with a context pointing at the target cluster and namespace

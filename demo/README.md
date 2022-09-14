# Demos for Meteor operator

Pre-requisites and dependencies:
- `oc` and `tkn` clients
- An OpenShift cluster already configured
- meteor-operator with CNBi support already deployed in the cluster
- KUBECONFIG with a context pointing at the target cluster and namespace

## Custom Notebook Image (CNBi) build of a git repo

The [cnbi-build-demo.sh](./cnbi-build-demo.sh) script uses
[demo-magic.sh](./demo-magic.sh) to present a demo of a `CustomNBImage` of type
`GitRepository`: [op1st-ds.yaml](./op1st-ds.yaml)

An asciinema recording of a cript run is available as
[op1st-ds.cast](./op1st-ds.cast). Use `asciinema play op1st-ds.cast` to
reproduce it.

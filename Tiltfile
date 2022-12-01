# -*- mode: Python -*-

kubectl_cmd = "kubectl"

if str(local("command -v " + kubectl_cmd + " || true", quiet = True)) == "":
    fail("Required command '" + kubectl_cmd + "' not found in PATH")

load("ext://uibutton", "cmd_button", "location", "text_input")
load('ext://podman', 'podman_build')

settings = {
    "enable_providers": ["docker"],
    "deploy_cert_manager": True,
    "kind_cluster_name": "cre-dev",
    "preload_images_for_kind": True,
    "debug": {},
    "cert_manager_version": "v1.9.1",
    "kubernetes_version": "v1.25.0",
    "kind_version": "v0.15.0",
}

if "allowed_contexts" in settings:
    allow_k8s_contexts(settings.get("allowed_contexts"))

if "default_registry" in settings:
    default_registry(settings.get("default_registry"))

load("ext://cert_manager", "deploy_cert_manager")

if settings.get("deploy_cert_manager"):
    deploy_cert_manager()


k8s_yaml('config/crd/test/tekton-pipeline-v0.39.0.yaml')
k8s_yaml('hack/openshift-client-v0.2.yaml')

k8s_yaml('hack/cre-gitrepo.yaml')
k8s_yaml('hack/cre-import.yaml')
k8s_yaml('hack/cre-prepare.yaml')

k8s_yaml(kustomize('config/default'))

podman_build('quay.io/thoth-station/meteor-operator:v0.2.0', '.')

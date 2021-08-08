package controllers

import "time"

const (
	MeteorPipelineLabel   = "meteor.operate-first.cloud/pipeline"
	MeteorDeploymentLabel = "meteor.operate-first.cloud/deployment"
	MeteorLabel           = "meteor.operate-first.cloud/meteor"

	ImageFormatter = "image-registry.openshift-image-registry.svc:5000/%s/%s-%s"
	requeue        = 10 * time.Second
)

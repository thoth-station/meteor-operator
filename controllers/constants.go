package controllers

import "time"

const (
	ImageFormatter = "image-registry.openshift-image-registry.svc:5000/%s/%s-%s"
	RequeueAfter   = 10 * time.Second
	Group          = "meteor.operate-first.cloud"
)

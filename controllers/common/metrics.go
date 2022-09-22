package common

import (
	"github.com/prometheus/client_golang/prometheus"
	meteorv1alpha1 "github.com/thoth-station/meteor-operator/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

const MeteorSubsystem = "meteor_operator"

var (
	MeteorCreated = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: MeteorSubsystem,
			Name:      "meteor_total",
			Help:      "Number of Meteors",
		},
		[]string{"meteor", "url", "ref"},
	)
	MeteorDeleted = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: MeteorSubsystem,
			Name:      "meteor_deleted_total",
			Help:      "Number of Meteors deleted",
		},
		[]string{"meteor", "url", "ref"},
	)
	MeteorPhase = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: MeteorSubsystem,
			Name:      "meteor_phase_total",
			Help:      "Gauge of current meteor phase",
		},
		[]string{"meteor", "phase"},
	)
	MeteorRemainingTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: MeteorSubsystem,
			Name:      "meteor_remaining_ttl_bucket",
			Help:      "Remaining TTL for a Meteor",
			Buckets:   prometheus.LinearBuckets(0, 3600, 48),
		},
		[]string{"meteor"},
	)
)

func InitMetrics() {
	metrics.Registry.MustRegister(MeteorCreated, MeteorDeleted, MeteorPhase, MeteorRemainingTime)
}

func MetricsBeforeReconcile(m *meteorv1alpha1.Meteor) {
	if m.Status.ExpirationTimestamp.IsZero() {
		// First time reconciling this meteor
		MeteorCreated.WithLabelValues(m.GetName(), m.Spec.Url, m.Spec.Ref).Inc()
	}
	MeteorPhase.WithLabelValues(m.GetName(), m.Status.Phase).Set(0)

	if !m.ObjectMeta.DeletionTimestamp.IsZero() {
		// Being deleted
		MeteorDeleted.WithLabelValues(m.GetName(), m.Spec.Url, m.Spec.Ref).Inc()
	}
}

func MetricsAfterReconcile(m *meteorv1alpha1.Meteor) {
	MeteorRemainingTime.WithLabelValues(m.GetName()).Observe(m.GetRemainingTTL())
	MeteorPhase.WithLabelValues(m.GetName(), m.Status.Phase).Set(1)
}

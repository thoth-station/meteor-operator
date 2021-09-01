package metrics

import (
	meteorv1alpha1 "github.com/aicoe/meteor-operator/api/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	MeteorCreated = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "meteor_total",
			Help: "Number of Meteors",
		},
		[]string{"meteor", "url", "ref"},
	)
	MeteorDeleted = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "meteor_deleted_total",
			Help: "Number of Meteors deleted",
		},
		[]string{"meteor", "url", "ref"},
	)
	MeteorPhase = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "meteor_phase_total",
			Help: "Gauge of current meteor phase",
		},
		[]string{"meteor", "phase"},
	)
	MeteorRemainingTime = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name: "meteor_remaining_ttl_bucket",
			Help: "Remaining TTL for a Meteor",
		},
	)
)

func Init() {
	metrics.Registry.MustRegister(MeteorCreated, MeteorDeleted, MeteorPhase, MeteorRemainingTime)
}

func BeforeReconcile(m *meteorv1alpha1.Meteor) {
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

func AfterReconcile(m *meteorv1alpha1.Meteor) {
	MeteorRemainingTime.Observe(m.GetRemainingTTL())
	MeteorPhase.WithLabelValues(m.GetName(), m.Status.Phase).Set(1)
}

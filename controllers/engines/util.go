package engines

import (
	a1 "github.com/premAI-io/saas-controller/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
)

func mergeProbe(src *a1.Probe, dst *v1.Probe) {
	if src == nil {
		return
	}

	if src.InitialDelaySeconds != nil {
		dst.InitialDelaySeconds = *src.InitialDelaySeconds
	}
	if src.PeriodSeconds != 0 {
		dst.PeriodSeconds = src.PeriodSeconds
	}
	if src.TimeoutSeconds != 0 {
		dst.TimeoutSeconds = src.TimeoutSeconds
	}
	if src.SuccessThreshold != 0 {
		dst.SuccessThreshold = src.SuccessThreshold
	}
	if src.FailureThreshold != 0 {
		dst.FailureThreshold = src.FailureThreshold
	}
}

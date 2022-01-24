package controllers

import (
	"reflect"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type ConfigMapChangedPredicate struct {
	predicate.Funcs
}

func (ConfigMapChangedPredicate) Update(e event.UpdateEvent) bool {
	if e.ObjectOld == nil || e.ObjectNew == nil {
		return false
	}

	oldSource, ok := e.ObjectOld.(*corev1.ConfigMap)
	if !ok {
		return false
	}

	newSource, ok := e.ObjectNew.(*corev1.ConfigMap)
	if !ok {
		return false
	}

	if oldSource.Data == nil && newSource.Data != nil {
		return true
	}

	if oldSource.Data != nil && newSource.Data != nil && !reflect.DeepEqual(oldSource.Data, newSource.Data) {
		return true
	}

	return false
}

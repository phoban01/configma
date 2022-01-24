/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"regexp"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	v1alpha1 "github.com/phoban01/configma/api/v1alpha1"
)

//TODO: add suspend field

// ConfigMatchReconciler reconciles a ConfigMatch object
type ConfigMatchReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=util.phoban.io,resources=configmatches,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=util.phoban.io,resources=configmatches/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=util.phoban.io,resources=configmatches/finalizers,verbs=update

//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is the main reconciler loop.
func (r *ConfigMatchReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// your logic here
	obj := &v1alpha1.ConfigMatch{}
	if err := r.Client.Get(ctx, req.NamespacedName, obj); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	pattern, err := regexp.Compile(obj.Spec.SourceRef.Pattern)
	if err != nil {
		logger.Error(err, "error compiling regex")
		return ctrl.Result{}, err
	}

	switch obj.Spec.SourceRef.Kind {
	case "ConfigMap":
		logger.Info("reconciling ConfigMap")
		cmList := &corev1.ConfigMapList{}
		if err := r.Client.List(ctx, cmList, client.MatchingLabels{
			v1alpha1.LabelMatcher: obj.Spec.SourceRef.MatchGroup,
		}); err != nil {
			return ctrl.Result{}, err
		}

		if len(cmList.Items) == 0 {
			logger.Info("no configmaps found")
			return ctrl.Result{}, nil
		}

		var newest corev1.ConfigMap
		for _, cm := range cmList.Items {
			if pattern.MatchString(cm.Name) {
				if newest.CreationTimestamp.Before(&cm.CreationTimestamp) {
					newest = cm
				}
			}
		}

		target := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      obj.Spec.Target.Name,
				Namespace: obj.Spec.Target.Namespace,
			},
		}

		if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, target, func() error {
			if target.CreationTimestamp.IsZero() {
				if err := controllerutil.SetOwnerReference(obj, target, r.Scheme); err != nil {
					return err
				}
			}
			target.Data = newest.Data
			return nil
		}); err != nil {
			return ctrl.Result{}, err
		}
	case "Secret":
		secretList := &corev1.SecretList{}
		if err := r.Client.List(ctx, secretList, client.MatchingLabels{
			v1alpha1.LabelMatcher: obj.GetLabels()[v1alpha1.LabelMatcher],
		}); err != nil {
			return ctrl.Result{}, err
		}

		if len(secretList.Items) == 0 {
			return ctrl.Result{}, nil
		}

		var newest corev1.Secret
		for _, cm := range secretList.Items {
			if pattern.MatchString(cm.Name) {
				if newest.CreationTimestamp.Before(&cm.CreationTimestamp) {
					newest = cm
				}
			}
		}

		target := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      obj.Spec.Target.Name,
				Namespace: obj.Spec.Target.Namespace,
			},
		}

		if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, target, func() error {
			if target.CreationTimestamp.IsZero() {
				if err := controllerutil.SetOwnerReference(obj, target, r.Scheme); err != nil {
					return err
				}
			}
			target.Data = newest.Data
			return nil
		}); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ConfigMatchReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetCache().IndexField(context.TODO(),
		&v1alpha1.ConfigMatch{}, v1alpha1.MatchLabelIndexKey, r.indexMatchLabel()); err != nil {
		return fmt.Errorf("failed setting index fields: %w", err)
	}

	selector, err := predicate.LabelSelectorPredicate(
		metav1.LabelSelector{
			MatchExpressions: []metav1.LabelSelectorRequirement{
				{
					Key:      v1alpha1.LabelMatcher,
					Operator: metav1.LabelSelectorOpExists,
					Values:   []string{},
				},
			},
		})
	if err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.ConfigMatch{}).
		Watches(
			&source.Kind{Type: &corev1.ConfigMap{}},
			handler.EnqueueRequestsFromMapFunc(r.requestsForConfigMap()),
			builder.WithPredicates(predicate.And(
				ConfigMapChangedPredicate{},
				selector,
			))).
		Watches(
			&source.Kind{Type: &corev1.Secret{}},
			handler.EnqueueRequestsFromMapFunc(r.requestsForSecret()),
			builder.WithPredicates(predicate.And(
				ConfigMapChangedPredicate{},
				selector,
			))).
		Complete(r)
}

func (r *ConfigMatchReconciler) requestsForConfigMap() func(obj client.Object) []reconcile.Request {
	return func(obj client.Object) []reconcile.Request {
		cm, ok := obj.(*corev1.ConfigMap)
		if !ok {
			return nil
		}

		ctx := context.Background()

		var listOption client.ListOption
		if value, ok := cm.GetLabels()[v1alpha1.LabelMatcher]; ok {
			listOption = client.MatchingFields{v1alpha1.MatchLabelIndexKey: value}
		}

		matches := &v1alpha1.ConfigMatchList{}
		if err := r.Client.List(ctx, matches, listOption); err != nil {
			return nil
		}

		var reqs []reconcile.Request
		for _, match := range matches.Items {
			pattern := regexp.MustCompile(match.Spec.SourceRef.Pattern)
			if pattern.MatchString(cm.Name) {
				reqs = append(reqs, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      match.Name,
						Namespace: match.Namespace,
					},
				})
			}
		}

		return reqs
	}
}

func (r *ConfigMatchReconciler) requestsForSecret() func(obj client.Object) []reconcile.Request {
	return func(obj client.Object) []reconcile.Request {
		secret, ok := obj.(*corev1.Secret)
		if !ok {
			return nil
		}

		ctx := context.Background()

		var listOption client.ListOption
		if value, ok := secret.GetLabels()[v1alpha1.LabelMatcher]; ok {
			listOption = client.MatchingLabels{v1alpha1.LabelMatcher: value}
		}

		matches := &v1alpha1.ConfigMatchList{}
		if err := r.Client.List(ctx, matches, listOption); err != nil {
			return nil
		}

		var reqs []reconcile.Request
		for _, match := range matches.Items {
			pattern := regexp.MustCompile(match.Spec.SourceRef.Pattern)
			if pattern.MatchString(secret.Name) {
				reqs = append(reqs, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      match.Name,
						Namespace: match.Namespace,
					},
				})
			}
		}

		return reqs
	}
}

func (r *ConfigMatchReconciler) indexMatchLabel() func(o client.Object) []string {
	return func(o client.Object) []string {
		cm, ok := o.(*v1alpha1.ConfigMatch)
		if !ok {
			panic(fmt.Sprintf("Expected a ConfigMatch, got %T", o))
		}

		return []string{cm.Spec.SourceRef.MatchGroup}
	}
}

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

package bpfdoperator

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"

	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	bpfdiov1alpha1 "github.com/bpfd-dev/bpfd/bpfd-operator/apis/v1alpha1"
	"github.com/bpfd-dev/bpfd/bpfd-operator/internal"
)

//+kubebuilder:rbac:groups=bpfd.dev,resources=uprobeprograms,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=bpfd.dev,resources=uprobeprograms/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=bpfd.dev,resources=uprobeprograms/finalizers,verbs=update

type UprobeProgramReconciler struct {
	ReconcilerCommon
}

func (r *UprobeProgramReconciler) getRecCommon() *ReconcilerCommon {
	return &r.ReconcilerCommon
}

func (r *UprobeProgramReconciler) getFinalizer() string {
	return internal.UprobeProgramControllerFinalizer
}

// SetupWithManager sets up the controller with the Manager.
func (r *UprobeProgramReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&bpfdiov1alpha1.UprobeProgram{}).
		// Watch bpfPrograms which are owned by UprobePrograms
		Watches(
			&source.Kind{Type: &bpfdiov1alpha1.BpfProgram{}},
			&handler.EnqueueRequestForObject{},
			builder.WithPredicates(predicate.And(statusChangedPredicate(), internal.BpfProgramTypePredicate(internal.UprobeString))),
		).
		Complete(r)
}

func (r *UprobeProgramReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Logger = log.FromContext(ctx)

	uprobeProgram := &bpfdiov1alpha1.UprobeProgram{}
	if err := r.Get(ctx, req.NamespacedName, uprobeProgram); err != nil {
		// Reconcile was triggered by bpfProgram event, get parent UprobeProgram Object.
		if errors.IsNotFound(err) {
			bpfProgram := &bpfdiov1alpha1.BpfProgram{}
			if err := r.Get(ctx, req.NamespacedName, bpfProgram); err != nil {
				if errors.IsNotFound(err) {
					r.Logger.V(1).Info("bpfProgram not found stale reconcile, exiting", "Name", req.NamespacedName)
				} else {
					r.Logger.Error(err, "failed getting bpfProgram Object", "Name", req.NamespacedName)
				}
				return ctrl.Result{}, nil
			}

			// Get owning UprobeProgram object from ownerRef
			ownerRef := metav1.GetControllerOf(bpfProgram)
			if ownerRef == nil {
				return ctrl.Result{Requeue: false}, fmt.Errorf("failed getting bpfProgram Object owner")
			}

			if err := r.Get(ctx, types.NamespacedName{Namespace: corev1.NamespaceAll, Name: ownerRef.Name}, uprobeProgram); err != nil {
				if errors.IsNotFound(err) {
					r.Logger.Info("Uprobe Program from ownerRef not found stale reconcile exiting", "Name", req.NamespacedName)
				} else {
					r.Logger.Error(err, "failed getting UprobeProgram Object from ownerRef", "Name", req.NamespacedName)
				}
				return ctrl.Result{}, nil
			}

		} else {
			r.Logger.Error(err, "failed getting UprobeProgram Object", "Name", req.NamespacedName)
			return ctrl.Result{}, nil
		}
	}

	return reconcileBpfProgram(ctx, r, uprobeProgram)
}

func (r *UprobeProgramReconciler) updateStatus(ctx context.Context, name string, cond bpfdiov1alpha1.ProgramConditionType, message string) (ctrl.Result, error) {
	// Sometimes we end up with a stale UprobeProgram due to races, do this
	// get to ensure we're up to date before attempting a status update.
	prog := &bpfdiov1alpha1.UprobeProgram{}
	if err := r.Get(ctx, types.NamespacedName{Namespace: corev1.NamespaceAll, Name: name}, prog); err != nil {
		r.Logger.V(1).Info("failed to get fresh UprobeProgram object...requeuing")
		return ctrl.Result{Requeue: true, RequeueAfter: retryDurationOperator}, nil
	}

	if prog.Status.Conditions != nil {
		// Get most recent condition
		recentIdx := len(prog.Status.Conditions) - 1

		// If the most recent condition is the same as input, just return.
		if prog.Status.Conditions[recentIdx].Type == string(cond) {
			return ctrl.Result{}, nil
		} else {
			// Remove the input condition from the list if it exists (may no exist)
			// because the SetStatusCondition() doesn't append if it is already in the list.
			meta.RemoveStatusCondition(&prog.Status.Conditions, string(cond))
		}
	}
	meta.SetStatusCondition(&prog.Status.Conditions, cond.Condition(message))

	if err := r.Status().Update(ctx, prog); err != nil {
		r.Logger.V(1).Info("failed to set UprobeProgram object status...requeuing")
		return ctrl.Result{Requeue: true, RequeueAfter: retryDurationOperator}, nil
	}

	return ctrl.Result{}, nil
}

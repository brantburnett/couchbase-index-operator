/*
Copyright 2021 Brant Burnett

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
	"reflect"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	v1beta1 "github.com/brantburnett/couchbase-index-operator/api/v1beta1"
	"github.com/go-logr/logr"
)

const indexSetFinalizer string = "couchbase.btburnett.com/indices"

// CouchbaseIndexSetReconciler reconciles a CouchbaseIndexSet object
type CouchbaseIndexSetReconciler struct {
	client.Client
	record.EventRecorder
	Scheme    *runtime.Scheme
	CbimImage string
}

type CouchbaseIndexSetReconcileContext struct {
	Ctx     context.Context
	Request ctrl.Request
	logr.Logger
	Reconciler *CouchbaseIndexSetReconciler

	IndexSet           v1beta1.CouchbaseIndexSet
	ConnectionString   string
	AdminSecretName    string
	DeletingIndexNames []string
	IsDeleting         bool
}

//+kubebuilder:rbac:groups=couchbase.btburnett.com,namespace=system,resources=couchbaseindexsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=couchbase.btburnett.com,namespace=system,resources=couchbaseindexsets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=couchbase.btburnett.com,namespace=system,resources=couchbaseindexsets/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *CouchbaseIndexSetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	context := CouchbaseIndexSetReconcileContext{
		Ctx:        ctx,
		Request:    req,
		Logger:     log.FromContext(ctx),
		Reconciler: r,
	}

	if err := r.Get(ctx, req.NamespacedName, &context.IndexSet); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	context.IsDeleting = context.IndexSet.DeletionTimestamp != nil

	context.V(1).Info("reconciling")

	if !controllerutil.ContainsFinalizer(&context.IndexSet, indexSetFinalizer) {
		if !context.IsDeleting {
			// Add the finalizer if missing

			if err := context.addFinalizer(); err != nil {
				// Don't continue if we can't add the finalizer, we don't want to leave orphaned indices
				return ctrl.Result{}, err
			}
		} else {
			// We've already done the finalization, nothing to do

			return ctrl.Result{}, nil
		}
	}

	primaryResult, primaryErr := r.primaryReconcile(&context)

	// Apply any status updates
	if err := r.updateStatus(&context); err != nil {
		return ctrl.Result{}, err
	}

	return primaryResult, primaryErr
}

// Performs primary reconcilation. Any actions in this context may change the status of the index set, it is expected
// that the caller will commit the changes.
func (r *CouchbaseIndexSetReconciler) primaryReconcile(context *CouchbaseIndexSetReconcileContext) (ctrl.Result, error) {
	// Update the status with the number of indices in the array
	// This is useful for the print columns display for kubectl
	context.IndexSet.Status.IndexCount = pointer.Int32Ptr(int32(len(context.IndexSet.Spec.Indices)))

	if ok, result, err := context.getConnectionInfo(); !ok {
		return result, err
	}

	return context.reconcileJob()
}

func (context *CouchbaseIndexSetReconcileContext) addFinalizer() error {
	controllerutil.AddFinalizer(&context.IndexSet, indexSetFinalizer)

	return context.Reconciler.Update(context.Ctx, &context.IndexSet)
}

func (context *CouchbaseIndexSetReconcileContext) removeFinalizer() error {
	controllerutil.RemoveFinalizer(&context.IndexSet, indexSetFinalizer)

	return context.Reconciler.Update(context.Ctx, &context.IndexSet)
}

func ignoreStatusChangePredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			if newJob, ok := e.ObjectNew.(*batchv1.Job); ok {
				if oldJob, ok := e.ObjectOld.(*batchv1.Job); ok {
					// We only care if the conditions change on a Job

					return !reflect.DeepEqual(oldJob.Status.Conditions, newJob.Status.Conditions)
				}
			}

			// We don't want to reconcile every time the status changes on the CouchbaseIndexSet or ConfigMap
			return e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration() ||
				!reflect.DeepEqual(e.ObjectOld.GetFinalizers(), e.ObjectNew.GetFinalizers())
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			if _, ok := e.Object.(*batchv1.Job); ok {
				// Ignore deletes of jobs to reduce reconcile cycles
				return false
			}

			return true
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *CouchbaseIndexSetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.EventRecorder = mgr.GetEventRecorderFor("couchbase-index-set-controller")

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.CouchbaseIndexSet{}).
		Owns(&batchv1.Job{}).
		Owns(&corev1.ConfigMap{}).
		WithEventFilter(ignoreStatusChangePredicate()).
		WithOptions(controller.Options{}).
		Complete(r)
}

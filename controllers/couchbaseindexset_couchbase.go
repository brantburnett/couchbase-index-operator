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
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	couchbasev2 "github.com/brantburnett/couchbase-index-operator/couchbase/v2"
)

func (context *CouchbaseIndexSetReconcileContext) getConnectionInfo() (bool, ctrl.Result, error) {
	if context.IndexSet.Spec.Cluster.ClusterRef != nil {
		if getResult, err := context.getCluster(); !getResult.IsZero() || err != nil {
			if context.IsDeleting && err == nil {
				// The cluster can't be found, etc. In this case during a delete we don't need to worry
				// about cleanup, just remove the finalizer.

				controllerutil.RemoveFinalizer(&context.IndexSet, indexSetFinalizer)
				if err := context.Reconciler.Update(context.Ctx, &context.IndexSet); err != nil {
					return false, ctrl.Result{}, err
				}

				return false, ctrl.Result{}, nil
			}

			return false, getResult, err
		}
	} else if context.IndexSet.Spec.Cluster.Manual != nil {
		context.ConnectionString = context.IndexSet.Spec.Cluster.Manual.ConnectionString
		context.AdminSecretName = context.IndexSet.Spec.Cluster.Manual.SecretName
	} else {
		setNotReady(&context.IndexSet, IndexSetReadyReasonCouchbaseError, "Missing connection info")

		return false, ctrl.Result{}, nil
	}

	// Success
	return true, ctrl.Result{}, nil
}

func (context *CouchbaseIndexSetReconcileContext) getCluster() (ctrl.Result, error) {
	clusterName := types.NamespacedName{
		Namespace: context.IndexSet.Namespace,
		Name:      context.IndexSet.Spec.Cluster.ClusterRef.Name,
	}

	cluster := couchbasev2.CouchbaseCluster{}
	if err := context.Reconciler.Get(context.Ctx, clusterName, &cluster); err != nil {
		if apierrors.IsNotFound(err) {
			setNotReady(&context.IndexSet, IndexSetReadyReasonCouchbaseError, "Cluster is not found")
			return ctrl.Result{Requeue: true}, nil
		}

		setNotReady(&context.IndexSet, IndexSetReadyReasonCouchbaseError, err.Error())
		return ctrl.Result{}, err
	}

	context.ConnectionString = fmt.Sprintf("couchbase://%s-srv", cluster.ObjectMeta.Name)
	if context.IndexSet.Spec.Cluster.ClusterRef.SecretName != nil {
		// Prefer the secret name if provided in our spec
		context.AdminSecretName = *context.IndexSet.Spec.Cluster.ClusterRef.SecretName
	} else {
		// Fallback to the admin secret name
		context.AdminSecretName = cluster.Spec.Security.AdminSecret
	}

	available := false
	for _, condition := range cluster.Status.Conditions {
		if condition.Type == "Available" && condition.Status == v1.ConditionTrue {
			available = true
			break
		}
	}

	if !available {
		setNotReady(&context.IndexSet, IndexSetReadyReasonCouchbaseError, "Cluster is not available")
		return ctrl.Result{Requeue: true}, nil
	}

	if cluster.Spec.Buckets.Managed {
		// If the cluster is managing buckets, we can confirm the bucket exists via the cluster status

		foundBucket := false
		for _, bucket := range cluster.Status.Buckets {
			if bucket.Name == context.IndexSet.Spec.BucketName {
				foundBucket = true
				break
			}
		}

		if !foundBucket {
			setNotReady(&context.IndexSet, IndexSetReadyReasonCouchbaseError, "Bucket is not found")
			return ctrl.Result{Requeue: true}, nil
		}
	}

	return ctrl.Result{}, nil
}

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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1beta1 "github.com/brantburnett/couchbase-index-operator/api/v1beta1"
)

type IndexSetSyncingReason string
type IndexSetReadyReason string

const (
	ConditionTypeSyncing string = "Syncing"
	ConditionTypeReady   string = "Ready"

	IndexSetSyncingReasonNotSyncing IndexSetSyncingReason = "NotSyncing"
	IndexSetSyncingReasonSyncing    IndexSetSyncingReason = "Syncing"

	IndexSetReadyReasonUnknown        IndexSetReadyReason = "Unknown"
	IndexSetReadyReasonInSync         IndexSetReadyReason = "InSync"
	IndexSetReadyReasonOutOfSync      IndexSetReadyReason = "OutOfSync"
	IndexSetReadyReasonPaused         IndexSetReadyReason = "Paused"
	IndexSetReadyReasonJobFailed      IndexSetReadyReason = "JobFailed"
	IndexSetReadyReasonConfigMapError IndexSetReadyReason = "ConfigMapError"
	IndexSetReadyReasonCouchbaseError IndexSetReadyReason = "CouchbaseError"
)

func getStatus(status bool) v1.ConditionStatus {
	if status {
		return v1.ConditionTrue
	} else {
		return v1.ConditionFalse
	}
}

func setSyncing(indexSet *v1beta1.CouchbaseIndexSet) {
	setSyncStatus(indexSet, true, IndexSetSyncingReasonSyncing, "Sync in progress")
}

func setNotSyncing(indexSet *v1beta1.CouchbaseIndexSet) {
	setSyncStatus(indexSet, false, IndexSetSyncingReasonNotSyncing, "Sync not in progress")
}

func setSyncStatus(indexSet *v1beta1.CouchbaseIndexSet, status bool, reason IndexSetSyncingReason, message string) {
	meta.SetStatusCondition(&indexSet.Status.Conditions, v1.Condition{
		Type:               ConditionTypeSyncing,
		Status:             getStatus(status),
		Message:            message,
		Reason:             string(reason),
		ObservedGeneration: indexSet.Generation,
	})
}

func setNotReady(indexSet *v1beta1.CouchbaseIndexSet, reason IndexSetReadyReason, message string) {
	setReadyStatus(indexSet, false, reason, message)
}

func setReadyInSync(indexSet *v1beta1.CouchbaseIndexSet) {
	setReadyStatus(indexSet, true, IndexSetReadyReasonInSync, "Indices are in sync")
}

func setReadyStatus(indexSet *v1beta1.CouchbaseIndexSet, status bool, reason IndexSetReadyReason, message string) {
	meta.SetStatusCondition(&indexSet.Status.Conditions, v1.Condition{
		Type:               ConditionTypeReady,
		Status:             getStatus(status),
		Message:            message,
		Reason:             string(reason),
		ObservedGeneration: indexSet.Generation,
	})
}

func getCurrentStateFromIndexSet(indexSet *v1beta1.CouchbaseIndexSet) IndexSetReadyReason {
	readyCondition := meta.FindStatusCondition(indexSet.Status.Conditions, ConditionTypeReady)

	return getCurrentStateFromCondition(readyCondition)
}

func getCurrentStateFromCondition(readyCondition *v1.Condition) IndexSetReadyReason {
	if readyCondition != nil {
		return IndexSetReadyReason(readyCondition.Reason)
	} else {
		return IndexSetReadyReasonUnknown
	}
}

func (r *CouchbaseIndexSetReconciler) updateStatus(context *CouchbaseIndexSetReconcileContext) error {
	// Because the object may have been modified, we need to make sure we have the latest object and just replace the status
	// This can help with the cache being out of date or external mutations. We try first without a Get, then 2 more times
	// with a fresh get, before we give up

	var err error
	for i := 0; i < 3; i++ {
		if i > 0 {
			// If this isn't the first attempt, then reload the latest index set via the API

			var tempIndexSet v1beta1.CouchbaseIndexSet
			if err = r.Get(context.Ctx, context.Request.NamespacedName, &tempIndexSet); err != nil {
				if apierrors.IsNotFound(err) {
					// Not found we can ignore, means the index set has been deleted, but we don't want to contine processing
					return nil
				}

				// Any other error break out to log and return
				break
			}

			// Backup the status, apply to the temp copy, and replace our index set
			tempStatus := context.IndexSet.Status
			context.IndexSet = tempIndexSet
			context.IndexSet.Status = tempStatus
		}

		if err = r.Status().Update(context.Ctx, &context.IndexSet); err == nil {
			return nil
		}
	}

	if err != nil {
		context.Error(err, "unable to update status")
	}

	return err
}

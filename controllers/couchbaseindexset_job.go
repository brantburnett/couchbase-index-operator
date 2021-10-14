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
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	v1beta1 "github.com/brantburnett/couchbase-index-operator/api/v1beta1"
	cbim "github.com/brantburnett/couchbase-index-operator/cbim"
)

const gsiAnnotationKey = "couchbase.btburnett.com/gsi"

type gsiAnnotation struct {
	Adding   []string `json:"adding,omitempty"`
	Deleting []string `json:"deleting,omitempty"`
}

type jobLookupResult struct {
	MostRecentJob *batchv1.Job
	OldJobs       []*batchv1.Job
}

func getTimeToNextSync(lastUpdate time.Time, syncInterval time.Duration) time.Duration {
	nextTransitionTime := lastUpdate.Add(syncInterval)

	timeToNextSync := time.Until(nextTransitionTime)

	if timeToNextSync > 0 {
		return timeToNextSync
	} else {
		return 0
	}
}

func isCurrentJob(context *CouchbaseIndexSetReconcileContext, job *batchv1.Job) bool {
	if context.IsDeleting {
		return job.Labels["deletion"] == "true"
	} else {
		jobGeneration := int64(0)
		var err error
		if jobGeneration, err = strconv.ParseInt(job.GetLabels()["generation"], 10, 64); err != nil {
			jobGeneration = 0
		}

		return jobGeneration == context.IndexSet.Generation
	}
}

func updateIndexStatus(indexSet *v1beta1.CouchbaseIndexSet, gsiAnnotationValue string) {
	if gsiAnnotationValue == "" {
		return
	}

	gsiAnnotation := gsiAnnotation{}
	if err := json.Unmarshal([]byte(gsiAnnotationValue), &gsiAnnotation); err != nil {
		return
	}

	hashSet := map[string]bool{}
	for _, v := range indexSet.Status.Indices {
		hashSet[v] = true
	}
	for _, v := range gsiAnnotation.Adding {
		hashSet[v] = true
	}
	for _, v := range gsiAnnotation.Deleting {
		delete(hashSet, v)
	}

	newList := make([]string, len(hashSet))
	i := 0
	for indexName := range hashSet {
		newList[i] = indexName
		i++
	}

	sort.Strings(newList)

	indexSet.Status.Indices = newList
}

func (context *CouchbaseIndexSetReconcileContext) getMostRecentJob() (jobLookupResult, error) {
	var labelRequirement *labels.Requirement
	var err error
	if labelRequirement, err = labels.NewRequirement("controller-uid", selection.Equals, []string{string(context.IndexSet.GetUID())}); err != nil {
		return jobLookupResult{}, err
	}

	// Find the most recent job, deleting any jobs found which are both old and not the most recent
	result := jobLookupResult{}
	mostRecentJobTimestamp := time.Time{}
	continuationToken := ""
	jobCount := 0

	for moreItems := true; moreItems; {
		var jobList = batchv1.JobList{}
		err = context.Reconciler.List(context.Ctx, &jobList, &client.ListOptions{
			LabelSelector: labels.NewSelector().Add(*labelRequirement),
			Namespace:     context.IndexSet.Namespace,
			Limit:         100,
			Continue:      continuationToken,
		})
		if err != nil {
			return jobLookupResult{}, err
		}

		continuationToken = jobList.Continue
		if continuationToken == "" {
			// No more items
			moreItems = false
		}

		for _, v := range jobList.Items {
			jobCount++
			creationTimestamp := v.GetCreationTimestamp().Time

			if creationTimestamp.After(mostRecentJobTimestamp) {
				if result.MostRecentJob != nil {
					// Let's cleanup the older job before we move on
					result.OldJobs = queueJobIfOld(result.OldJobs, result.MostRecentJob)
				}

				mostRecentJobTimestamp = creationTimestamp
				result.MostRecentJob = v.DeepCopy()
			} else {
				// Let's cleanup the older job before we move on
				result.OldJobs = queueJobIfOld(result.OldJobs, v.DeepCopy())
			}
		}
	}

	return result, nil
}

func (context *CouchbaseIndexSetReconcileContext) reconcileJob() (ctrl.Result, error) {
	var job *batchv1.Job = nil

	if lookupResult, err := context.getMostRecentJob(); err != nil {
		return ctrl.Result{}, err
	} else {
		// Delete any old jobs, ignoring errors and continuing with reconcile
		for _, v := range lookupResult.OldJobs {
			_ = context.deleteJob(v)
		}

		job = lookupResult.MostRecentJob
	}

	isCurrentJob := job != nil && isCurrentJob(context, job)
	if !isCurrentJob {
		// The current job isn't for the latest spec, so we're out of date, make sure we've cleared the ready state
		// We do this regardless of whether a job is running currently
		setNotReady(&context.IndexSet, IndexSetReadyReasonOutOfSync, "Indices are out of sync")
	}

	// Get the status and generation of the most recent job
	jobStatus := getJobStatus(job)
	if jobStatus == jobRunning {
		// We can't start another job until the current one completes
		// Mark the condition as syncing and the next status change of the job will trigger another reconcile

		setSyncing(&context.IndexSet)
		context.V(1).Info("Sync is still in progress")
		return ctrl.Result{}, nil
	} else {
		// We're not syncing, clear the condition
		setNotSyncing(&context.IndexSet)

		if jobStatus == jobCompleted {
			// We always want to track indices, even if we're about to start a fresh run, so do that first
			updateIndexStatus(&context.IndexSet, job.GetAnnotations()[gsiAnnotationKey])
		}
	}

	if isCurrentJob {
		// If this job is the current job, handle status updates for success/failure and sleeps
		// Note that if we've reached this code, the status is definitely etiher jobFailed or jobCompleted

		switch jobStatus {
		case jobFailed:
			// We want to resync immediately (no sleep) if the index set has changed
			// Make sure we don't flood ourselves if we're failing by waiting 1 minute before retrying
			// Jobs don't include a failed time, so we do the best we can using the start time
			timeToNextSync := getTimeToNextSync(job.Status.StartTime.Time, time.Minute)

			if timeToNextSync > 0 {
				if getCurrentStateFromIndexSet(&context.IndexSet) != IndexSetReadyReasonJobFailed {
					context.V(1).Info("Index sync job failed")
					context.Reconciler.Event(&context.IndexSet, "Warning", "SyncFailed", "Sync failed")

					setNotReady(&context.IndexSet, IndexSetReadyReasonJobFailed, "Sync failed")
				}

				return ctrl.Result{RequeueAfter: timeToNextSync}, nil
			}

			// Fall through to create a new job below

		case jobCompleted:
			if context.IsDeleting {
				context.V(1).Info("Cleanup job successful")

				// We're done with index cleanup, remove the finalizer
				if err := context.removeFinalizer(); err != nil {
					return ctrl.Result{}, err
				}

				return ctrl.Result{}, nil
			}

			// Since the most recent job was successful, sleep 5 minutes from the time it completed
			timeToNextSync := getTimeToNextSync(job.Status.CompletionTime.Time, time.Minute*5)

			if timeToNextSync > 0 {
				if getCurrentStateFromIndexSet(&context.IndexSet) != IndexSetReadyReasonInSync {
					context.V(1).Info("Index sync job succeeded")
					context.Reconciler.Event(&context.IndexSet, "Normal", "SyncComplete", "Sync completed")

					setReadyInSync(&context.IndexSet)
				}

				return ctrl.Result{RequeueAfter: timeToNextSync}, nil
			}

			// Fall through to create a new job below
		}
	}

	// Update the config map before starting the job

	if configMapName, err := context.reconcileConfigMap(); err != nil {
		setNotReady(&context.IndexSet, IndexSetReadyReasonConfigMapError, err.Error())

		return ctrl.Result{}, err
	} else {
		context.IndexSet.Status.ConfigMapName = configMapName
	}

	// Create the job

	if err := context.createJob(); err != nil {
		setNotReady(&context.IndexSet, IndexSetReadyReasonJobFailed, "Failed to create sync job")
		return ctrl.Result{}, err
	}

	// Since we started a new job, we can clean up the old one if it's old
	if job != nil {
		_ = context.deleteJobIfOld(job)
	}

	return ctrl.Result{}, nil
}

func (context *CouchbaseIndexSetReconcileContext) createJob() error {
	context.V(1).Info("Creating index sync job")

	gsiAnnotation := gsiAnnotation{
		Adding:   make([]string, len(context.IndexSet.Spec.Indices)),
		Deleting: make([]string, len(context.DeletingIndexes)),
	}
	for i, v := range context.IndexSet.Spec.Indices {
		gsiAnnotation.Adding[i] = cbim.GetIndexIdentifier(v).ToString()
	}
	for i, v := range context.DeletingIndexes {
		gsiAnnotation.Deleting[i] = v.ToString()
	}
	gsiAnnotationValue, _ := json.Marshal(gsiAnnotation)

	labels := map[string]string{
		"controller-uid": string(context.IndexSet.GetUID()),
		"generation":     fmt.Sprintf("%d", context.IndexSet.GetGeneration()),
		"bucketName":     context.IndexSet.Spec.BucketName,
	}

	if context.IndexSet.Spec.Cluster.ClusterRef != nil {
		labels["clusterName"] = context.IndexSet.Spec.Cluster.ClusterRef.Name
	}
	if context.IsDeleting {
		labels["deletion"] = "true"
	}

	activeDeadlineSeconds := context.IndexSet.Spec.ActiveDeadlineSeconds
	if activeDeadlineSeconds == nil {
		activeDeadlineSeconds = pointer.Int64Ptr(600)
	}

	backoffLimit := context.IndexSet.Spec.BackoffLimit
	if backoffLimit == nil {
		backoffLimit = pointer.Int32Ptr(2)
	}

	job := batchv1.Job{
		ObjectMeta: v1.ObjectMeta{
			Namespace:    context.IndexSet.GetNamespace(),
			GenerateName: fmt.Sprintf("%s-", context.IndexSet.Name),
			Labels:       labels,
			Annotations: map[string]string{
				gsiAnnotationKey: string(gsiAnnotationValue),
			},
		},
		Spec: batchv1.JobSpec{
			ActiveDeadlineSeconds: activeDeadlineSeconds,
			BackoffLimit:          backoffLimit,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:  "couchbase-index-manager",
							Image: context.Reconciler.CbimImage,
							Args: []string{
								"-c",
								context.ConnectionString,
								"-u",
								"$(USERNAME)",
								"-p",
								"$(PASSWORD)",
								"sync",
								"--force",
								context.IndexSet.Spec.BucketName,
								"/spec",
							},
							Env: []corev1.EnvVar{
								{
									Name: "USERNAME",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											Key: "username",
											LocalObjectReference: corev1.LocalObjectReference{
												Name: context.AdminSecretName,
											},
										},
									},
								},
								{
									Name: "PASSWORD",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											Key: "password",
											LocalObjectReference: corev1.LocalObjectReference{
												Name: context.AdminSecretName,
											},
										},
									},
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "indexspec",
									ReadOnly:  true,
									MountPath: "/spec",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "indexspec",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: context.IndexSet.Status.ConfigMapName,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	controllerutil.SetControllerReference(&context.IndexSet, &job, context.Reconciler.Scheme)

	if err := context.Reconciler.Create(context.Ctx, &job); err != nil {
		return err
	}

	context.Info("Created index sync job", "jobName", job.GetName())
	setSyncing(&context.IndexSet)
	context.Reconciler.Event(&context.IndexSet, "Normal", "SyncStarted", "Sync started")

	return nil
}

func isJobOld(job *batchv1.Job) bool {
	return job.CreationTimestamp.Time.Before(time.Now().Add(time.Minute * -15))
}

func queueJobIfOld(queue []*batchv1.Job, job *batchv1.Job) []*batchv1.Job {
	if isJobOld(job) {
		return append(queue, job)
	}

	return queue
}

func (context *CouchbaseIndexSetReconcileContext) deleteJobIfOld(job *batchv1.Job) error {
	if isJobOld(job) {
		return context.deleteJob(job)
	}

	return nil
}

func (context *CouchbaseIndexSetReconcileContext) deleteJob(job *batchv1.Job) error {
	context.V(1).Info("Deleting job", "jobName", job.GetName())

	deletePropagationBackground := v1.DeletePropagationBackground
	deleteOptions := client.DeleteOptions{
		PropagationPolicy: &deletePropagationBackground,
	}

	if err := context.Reconciler.Delete(context.Ctx, job, &deleteOptions); client.IgnoreNotFound(err) != nil {
		context.Error(err, "Failed to delete Job", "jobName", job.GetName())

		return err
	}

	return nil
}

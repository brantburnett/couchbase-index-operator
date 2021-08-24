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
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

type jobStatus int

const (
	jobNotFound jobStatus = iota
	jobRunning
	jobFailed
	jobCompleted
)

func isJobStatusConditionTrue(conditions []batchv1.JobCondition, conditionType batchv1.JobConditionType) bool {
	return isJobStatusConditionPresentAndEqual(conditions, conditionType, corev1.ConditionTrue)
}

func isJobStatusConditionPresentAndEqual(conditions []batchv1.JobCondition, conditionType batchv1.JobConditionType, status corev1.ConditionStatus) bool {
	if conditions == nil {
		return false
	}

	for _, condition := range conditions {
		if condition.Type == conditionType {
			return condition.Status == status
		}
	}
	return false
}

func getJobStatus(job *batchv1.Job) jobStatus {
	if job == nil {
		return jobNotFound
	} else if isJobStatusConditionTrue(job.Status.Conditions, "Failed") {
		return jobFailed
	} else if isJobStatusConditionTrue(job.Status.Conditions, "Complete") {
		return jobCompleted
	} else {
		return jobRunning
	}
}

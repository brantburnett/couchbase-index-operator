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
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/brantburnett/couchbase-index-operator/cbim"
)

func (context *CouchbaseIndexSetReconcileContext) reconcileConfigMap() (string, error) {
	name := types.NamespacedName{
		Namespace: context.IndexSet.Namespace,
		Name:      context.IndexSet.Name + "-indexspec",
	}

	var (
		configMap corev1.ConfigMap
		isNew     bool = false
	)
	if err := context.Reconciler.Get(context.Ctx, name, &configMap); err != nil {
		if !apierrors.IsNotFound(err) {
			context.Error(err, "unable to fetch ConfigMap")
			return "", err
		}

		// Config map doesn't exist, so make a new one
		isNew = true
		configMap = corev1.ConfigMap{
			ObjectMeta: v1.ObjectMeta{
				Namespace: name.Namespace,
				Name:      name.Name,
			},
		}
	}

	yaml, err := cbim.GenerateYaml(&context.IndexSet, &context.DeletingIndexes)
	if err != nil {
		context.Error(err, "Error generating index spec")
		return "", err
	}

	if configMap.Data == nil ||
		len(configMap.Data) != 1 ||
		configMap.Data["indices.yaml"] != yaml {

		configMap.Data = map[string]string{
			"indices.yaml": yaml,
		}
	} else {
		// Already in sync, do nothing
		return name.Name, nil
	}

	controllerutil.SetControllerReference(&context.IndexSet, &configMap, context.Reconciler.Scheme)

	if isNew {
		context.Info("Creating config map")

		if err := context.Reconciler.Create(context.Ctx, &configMap); err != nil {
			context.Error(err, "Error creating config map")
			return "", err
		}
	} else {
		context.Info("Updating config map")

		if err := context.Reconciler.Update(context.Ctx, &configMap); err != nil {
			context.Error(err, "Error updating config map")
			return "", err
		}
	}

	return name.Name, nil
}

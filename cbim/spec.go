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

package cbim

import (
	"strings"

	"k8s.io/utils/pointer"
	"sigs.k8s.io/yaml"

	couchbasev1beta1 "github.com/brantburnett/couchbase-index-operator/api/v1beta1"
)

func GenerateYaml(indexSet *couchbasev1beta1.CouchbaseIndexSet, deletingIndexes *[]GlobalSecondaryIndexIdentifier) (string, error) {
	var sb strings.Builder

	definedIndexes := map[GlobalSecondaryIndexIdentifier]bool{}

	if indexSet.GetDeletionTimestamp() == nil {
		// Only create indices if we're not deleting the index set
		// If we are deleting, this will leave definedIndexNames empty so all indices are deleted

		for _, gsi := range indexSet.Spec.Indices {
			definedIndexes[GetIndexIdentifier(gsi)] = true

			if err := addIndexSpec(&sb, createIndexSpec(&gsi)); err != nil {
				return "", err
			}
		}
	}

	*deletingIndexes = []GlobalSecondaryIndexIdentifier{}
	for _, index := range indexSet.Status.Indices {
		if indexIdentifier, err := ParseIndexIdentifierString(index); err == nil {
			if !definedIndexes[indexIdentifier] {
				*deletingIndexes = append(*deletingIndexes, indexIdentifier)

				if err := addIndexSpec(&sb, createIndexDeleteSpec(indexIdentifier)); err != nil {
					return "", err
				}
			}
		}
	}

	return sb.String(), nil
}

func addIndexSpec(sb *strings.Builder, spec IndexSpec) error {
	yaml, err := yaml.Marshal(spec)
	if err != nil {
		return err
	}

	if sb.Len() > 0 {
		sb.WriteString("---\n")
	}
	sb.WriteString(string(yaml))

	return nil
}

func createIndexSpec(gsi *couchbasev1beta1.GlobalSecondaryIndex) IndexSpec {
	return IndexSpec{
		Name:               gsi.Name,
		Scope:              gsi.ScopeName,
		Collection:         gsi.CollectionName,
		IndexKey:           &gsi.IndexKey,
		Condition:          gsi.Condition,
		NumReplicas:        gsi.NumReplicas,
		RetainDeletedXattr: gsi.RetainDeletedXAttr,
		Partition:          mapPartition(gsi.Partition),
	}
}

func mapPartition(partition *couchbasev1beta1.GlobalSecondaryIndexPartition) *PartitionSpec {
	if partition == nil {
		return nil
	}

	return &PartitionSpec{
		Expressions:   partition.Expressions,
		Strategy:      mapPartitionStrategy(partition.Strategy),
		NumPartitions: partition.NumPartitions,
	}
}

func mapPartitionStrategy(strategy *string) *string {
	if strategy == nil {
		return nil
	}

	result := strings.ToLower(*strategy)
	return &result
}

func createIndexDeleteSpec(indexIdentifier GlobalSecondaryIndexIdentifier) IndexSpec {
	return IndexSpec{
		Name:       indexIdentifier.Name,
		Scope:      &indexIdentifier.ScopeName,
		Collection: &indexIdentifier.CollectionName,
		Lifecycle: &LifecycleSpec{
			Drop: pointer.BoolPtr(true),
		},
	}
}

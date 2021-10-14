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

const (
	StrategyHash = "hash"
)

type PartitionSpec struct {
	Expressions   []string `json:"exprs"`
	Strategy      *string  `json:"strategy,omitempty"`
	NumPartitions *int     `json:"num_partition,omitempty"`
}

type LifecycleSpec struct {
	Drop *bool `json:"drop,omitempty"`
}

type IndexSpec struct {
	Type               string         `json:"type,omitempty"`
	Name               string         `json:"name"`
	Scope              *string        `json:"scope,omitempty"`
	Collection         *string        `json:"collection,omitempty"`
	IsPrimary          *bool          `json:"is_primary,omitempty"`
	IndexKey           *[]string      `json:"index_key,omitempty"`
	Condition          *string        `json:"condition,omitempty"`
	RetainDeletedXattr *bool          `json:"retain_deleted_xattr,omitempty"`
	Partition          *PartitionSpec `json:"partition,omitempty"`
	ManualReplica      *bool          `json:"manual_replica,omitempty"`
	NumReplicas        *int           `json:"num_replica,omitempty"`
	Nodes              *[]string      `json:"nodes,omitempty"`
	Lifecycle          *LifecycleSpec `json:"lifecycle,omitempty"`
}

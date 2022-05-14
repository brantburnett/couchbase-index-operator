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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// Defines partition information for a partitioned index
type GlobalSecondaryIndexPartition struct {
	//+kubebuilder:validation:MinItems:=1
	// Attributes to be used to partition documents across nodes
	Expressions []string `json:"expressions"`
	//+kubebuilder:default:=Hash
	//+kubebuilder:validation:Enum:=Hash
	// Partition strategy to use, defaults to Hash (which is currently the only option)
	Strategy *string `json:"strategy,omitempty"`
	//+kubebuilder:validation:Minimum:=2
	NumPartitions *int `json:"numPartitions,omitempty"`
}

// Defines the desired state of a Couchbase Global Secondary Index
type GlobalSecondaryIndex struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	//+kubebuilder:validation:MinLength:=1
	//+kubebuilder:validation:Pattern:=^[A-Za-z][A-Za-z0-9#_\-]*$
	// Name of the index
	Name string `json:"name"`
	//+kubebuilder:validation:MinLength:=1
	//+kubebuilder:validation:Pattern:="^_default$|^[A-Za-z0-9\\-][A-Za-z0-9_\\-%]*$"
	// Name of the index's scope, assumes "_default" if not present
	ScopeName *string `json:"scopeName,omitempty"`
	//+kubebuilder:validation:MinLength:=1
	//+kubebuilder:validation:Pattern:="^_default$|^[A-Za-z0-9\\-][A-Za-z0-9_\\-%]*$"
	// Name of the index's collection, assumes "_default" if not present
	CollectionName *string `json:"collectionName,omitempty"`
	//+kubebuilder:validation:MinItems:=1
	// List of properties or deterministic functions which make up the index key
	IndexKey []string `json:"indexKey"`
	// Conditions to filter documents included on the index
	Condition *string `json:"condition,omitempty"`
	//+kubebuilder:validation:Minimum:=0
	// Number of replicas
	NumReplicas *int `json:"numReplicas,omitempty"`
	// Enable for Sync Gateway indices to preserve deleted XAttrs
	RetainDeletedXAttr *bool `json:"retainDeletedXAttr,omitempty"`
	// Defines partition information for a partitioned index
	Partition *GlobalSecondaryIndexPartition `json:"partition,omitempty"`
}

type CouchbaseClusterRef struct {
	// Name of the CouchbaseCluster resource in Kubernetes. This resource must be in the same namespace.
	Name string `json:"name"`
	// Optional name of a secret containing a username and password. If not present, uses the AdminSecretName found on the CouchbaseCluster resource.
	SecretName *string `json:"secretName,omitempty"`
}

type CouchbaseClusterManual struct {
	//+kubebuilder:validation:Pattern:="^couchbases?:\\/\\/(([\\w\\d\\-\\_]+\\.)*[\\w\\d\\-\\_]+,)*([\\w\\d\\-\\_]+\\.)*[\\w\\d\\-\\_]+(:\\d+)?\\/?$"
	// Couchbase connection string, in "couchbase://" format
	ConnectionString string `json:"connectionString"`
	// Name of a secret containing a username and password
	SecretName string `json:"secretName"`
}

//+kubebuilder:validation:MinProperties:=1
//+kubebuilder:validation:MaxProperties:=1
// Defines how to connect to a Couchbase cluster
type CouchbaseCluster struct {
	// Connect via a CouchbaseCluster resource in Kubernetes
	ClusterRef *CouchbaseClusterRef `json:"clusterRef,omitempty"`
	// Connect via manual connection information
	Manual *CouchbaseClusterManual `json:"manual,omitempty"`
}

// Defines the desired state of a set of Couchbase indices
type CouchbaseIndexSetSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Defines how to connect to a Couchbase cluster
	Cluster CouchbaseCluster `json:"cluster"`
	// Name of the bucket
	BucketName string `json:"bucketName"`
	//+listType:=map
	//+listMapKey:=name
	// List of global secondary indices
	Indices []GlobalSecondaryIndex `json:"indices,omitempty"`
	//+kubebuilder:default:=600
	//+kubebuilder:validation:Optional
	//+kubebuilder:validation:Minimum:=1
	// Specifies the duration in seconds relative to the startTime that a sync attempt may be active before the system tries to terminate it; value must be positive integer
	ActiveDeadlineSeconds *int64 `json:"activeDeadlineSeconds,omitempty"`
	//+kubebuilder:default:=2
	//+kubebuilder:validation:Optional
	//+kubebuilder:validation:Minimum:=0
	// Specifies the number of retries before marking a sync attempt as failed.
	BackoffLimit *int32 `json:"backoffLimit,omitempty"`
	//+kubebuilder:default=false
	//+kubebuilder:validation:Optional
	// Pauses index synchronization for this index set. Deleting the index set will still perform cleanup.
	Paused *bool `json:"paused"`
}

// Defines the observed state of CouchbaseIndexSet
type CouchbaseIndexSetStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	//+listType:=map
	//+listMapKey:=type
	// Conditions represent the latest available observations of an object's state
	Conditions []metav1.Condition `json:"conditions"`
	// Name of the generated config map
	ConfigMapName string `json:"configMapName,omitempty"`
	//+listType:=atomic
	// List of global secondary indices created and managed by this resource
	Indices []string `json:"indices,omitempty"`
	// Number of indices
	IndexCount *int32 `json:"indexCount"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

//+kubebuilder:printcolumn:name="Bucket",type=string,JSONPath=`.spec.bucketName`
//+kubebuilder:printcolumn:name="Indices",type=integer,JSONPath=`.status.indexCount`
//+kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// Defines a set of Couchbase indices
type CouchbaseIndexSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CouchbaseIndexSetSpec   `json:"spec,omitempty"`
	Status CouchbaseIndexSetStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CouchbaseIndexSetList contains a list of CouchbaseIndexSet
type CouchbaseIndexSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CouchbaseIndexSet `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CouchbaseIndexSet{}, &CouchbaseIndexSetList{})
}

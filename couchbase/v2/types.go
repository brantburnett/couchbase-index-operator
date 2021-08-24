package v2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: This file only partially defines the Couchbase CRDs, focusing on fields we care about
// Important: Run "make" to regenerate code after modifying this file

const (
	BucketTypeCouchbase = "couchbase"
)

type Security struct {
	AdminSecret string `json:"adminSecret,omitempty"`
}

type Buckets struct {
	Managed bool `json:"managed"`
}

type Bucket struct {
	Name string `json:"name,omitempty"`
	Type string `json:"type,omitempty"`
}

type CouchbaseClusterSpec struct {
	Security Security `json:"security"`
	Buckets  Buckets  `json:"buckets"`
}

type CouchbaseClusterStatus struct {
	Conditions []metav1.Condition `json:"conditions"`
	Buckets    []Bucket           `json:"buckets,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

type CouchbaseCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CouchbaseClusterSpec   `json:"spec,omitempty"`
	Status CouchbaseClusterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CouchbaseIndexSetList contains a list of CouchbaseIndexSet
type CouchbaseClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CouchbaseCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CouchbaseCluster{}, &CouchbaseClusterList{})
}

//go:build !ignore_autogenerated
// +build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1beta1

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CouchbaseCluster) DeepCopyInto(out *CouchbaseCluster) {
	*out = *in
	if in.ClusterRef != nil {
		in, out := &in.ClusterRef, &out.ClusterRef
		*out = new(CouchbaseClusterRef)
		(*in).DeepCopyInto(*out)
	}
	if in.Manual != nil {
		in, out := &in.Manual, &out.Manual
		*out = new(CouchbaseClusterManual)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CouchbaseCluster.
func (in *CouchbaseCluster) DeepCopy() *CouchbaseCluster {
	if in == nil {
		return nil
	}
	out := new(CouchbaseCluster)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CouchbaseClusterManual) DeepCopyInto(out *CouchbaseClusterManual) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CouchbaseClusterManual.
func (in *CouchbaseClusterManual) DeepCopy() *CouchbaseClusterManual {
	if in == nil {
		return nil
	}
	out := new(CouchbaseClusterManual)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CouchbaseClusterRef) DeepCopyInto(out *CouchbaseClusterRef) {
	*out = *in
	if in.SecretName != nil {
		in, out := &in.SecretName, &out.SecretName
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CouchbaseClusterRef.
func (in *CouchbaseClusterRef) DeepCopy() *CouchbaseClusterRef {
	if in == nil {
		return nil
	}
	out := new(CouchbaseClusterRef)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CouchbaseIndexSet) DeepCopyInto(out *CouchbaseIndexSet) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CouchbaseIndexSet.
func (in *CouchbaseIndexSet) DeepCopy() *CouchbaseIndexSet {
	if in == nil {
		return nil
	}
	out := new(CouchbaseIndexSet)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CouchbaseIndexSet) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CouchbaseIndexSetList) DeepCopyInto(out *CouchbaseIndexSetList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]CouchbaseIndexSet, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CouchbaseIndexSetList.
func (in *CouchbaseIndexSetList) DeepCopy() *CouchbaseIndexSetList {
	if in == nil {
		return nil
	}
	out := new(CouchbaseIndexSetList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CouchbaseIndexSetList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CouchbaseIndexSetSpec) DeepCopyInto(out *CouchbaseIndexSetSpec) {
	*out = *in
	in.Cluster.DeepCopyInto(&out.Cluster)
	if in.Indices != nil {
		in, out := &in.Indices, &out.Indices
		*out = make([]GlobalSecondaryIndex, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.ActiveDeadlineSeconds != nil {
		in, out := &in.ActiveDeadlineSeconds, &out.ActiveDeadlineSeconds
		*out = new(int64)
		**out = **in
	}
	if in.BackoffLimit != nil {
		in, out := &in.BackoffLimit, &out.BackoffLimit
		*out = new(int32)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CouchbaseIndexSetSpec.
func (in *CouchbaseIndexSetSpec) DeepCopy() *CouchbaseIndexSetSpec {
	if in == nil {
		return nil
	}
	out := new(CouchbaseIndexSetSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CouchbaseIndexSetStatus) DeepCopyInto(out *CouchbaseIndexSetStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Indices != nil {
		in, out := &in.Indices, &out.Indices
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.IndexCount != nil {
		in, out := &in.IndexCount, &out.IndexCount
		*out = new(int32)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CouchbaseIndexSetStatus.
func (in *CouchbaseIndexSetStatus) DeepCopy() *CouchbaseIndexSetStatus {
	if in == nil {
		return nil
	}
	out := new(CouchbaseIndexSetStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GlobalSecondaryIndex) DeepCopyInto(out *GlobalSecondaryIndex) {
	*out = *in
	if in.ScopeName != nil {
		in, out := &in.ScopeName, &out.ScopeName
		*out = new(string)
		**out = **in
	}
	if in.CollectionName != nil {
		in, out := &in.CollectionName, &out.CollectionName
		*out = new(string)
		**out = **in
	}
	if in.IndexKey != nil {
		in, out := &in.IndexKey, &out.IndexKey
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Condition != nil {
		in, out := &in.Condition, &out.Condition
		*out = new(string)
		**out = **in
	}
	if in.NumReplicas != nil {
		in, out := &in.NumReplicas, &out.NumReplicas
		*out = new(int)
		**out = **in
	}
	if in.RetainDeletedXAttr != nil {
		in, out := &in.RetainDeletedXAttr, &out.RetainDeletedXAttr
		*out = new(bool)
		**out = **in
	}
	if in.Partition != nil {
		in, out := &in.Partition, &out.Partition
		*out = new(GlobalSecondaryIndexPartition)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GlobalSecondaryIndex.
func (in *GlobalSecondaryIndex) DeepCopy() *GlobalSecondaryIndex {
	if in == nil {
		return nil
	}
	out := new(GlobalSecondaryIndex)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GlobalSecondaryIndexPartition) DeepCopyInto(out *GlobalSecondaryIndexPartition) {
	*out = *in
	if in.Expressions != nil {
		in, out := &in.Expressions, &out.Expressions
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Strategy != nil {
		in, out := &in.Strategy, &out.Strategy
		*out = new(string)
		**out = **in
	}
	if in.NumPartitions != nil {
		in, out := &in.NumPartitions, &out.NumPartitions
		*out = new(int)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GlobalSecondaryIndexPartition.
func (in *GlobalSecondaryIndexPartition) DeepCopy() *GlobalSecondaryIndexPartition {
	if in == nil {
		return nil
	}
	out := new(GlobalSecondaryIndexPartition)
	in.DeepCopyInto(out)
	return out
}

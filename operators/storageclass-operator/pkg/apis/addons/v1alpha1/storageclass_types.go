/*
Copyright 2018 The Kubernetes Authors.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/alpha/patterns/addon"
)

// StorageClassSpec defines the desired state of StorageClass
type StorageClassSpec struct {
	addon.CommonSpec
}

// StorageClassStatus defines the observed state of StorageClass
type StorageClassStatus struct {
	addon.CommonStatus
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// StorageClass is the Schema for the storageclasses API
// +k8s:openapi-gen=true
type StorageClass struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StorageClassSpec   `json:"spec,omitempty"`
	Status StorageClassStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// StorageClassList contains a list of StorageClass
type StorageClassList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StorageClass `json:"items"`
}

func init() {
	SchemeBuilder.Register(&StorageClass{}, &StorageClassList{})
}

var _ addon.CommonObject = &StorageClass{}

func (c *StorageClass) ComponentName() string {
	return "storageclass"
}

func (c *StorageClass) CommonSpec() addon.CommonSpec {
	return c.Spec.CommonSpec
}

func (c *StorageClass) GetCommonStatus() addon.CommonStatus {
	return c.Status.CommonStatus
}

func (c *StorageClass) SetCommonStatus(s addon.CommonStatus) {
	c.Status.CommonStatus = s
}
